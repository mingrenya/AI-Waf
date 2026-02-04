package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// MiddlewareAudit 中间件审计工具
type MiddlewareAudit struct {
	routerFiles []string
	issues      []AuditIssue
}

// AuditIssue 审计发现的问题
type AuditIssue struct {
	File        string
	Line        int
	Severity    string // HIGH, MEDIUM, LOW
	Category    string
	Description string
	Suggestion  string
}

func main() {
	audit := &MiddlewareAudit{}

	// 查找所有路由文件
	if err := audit.findRouterFiles("./router"); err != nil {
		fmt.Printf("Error finding router files: %v\n", err)
		os.Exit(1)
	}

	// 审计每个文件
	for _, file := range audit.routerFiles {
		if err := audit.auditFile(file); err != nil {
			fmt.Printf("Error auditing %s: %v\n", file, err)
		}
	}

	// 输出报告
	audit.printReport()
}

func (a *MiddlewareAudit) findRouterFiles(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			a.routerFiles = append(a.routerFiles, path)
		}
		return nil
	})
}

func (a *MiddlewareAudit) auditFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 遍历AST
	ast.Inspect(node, func(n ast.Node) bool {
		// 检查函数调用
		if call, ok := n.(*ast.CallExpr); ok {
			a.checkRouteDefinition(fset, call, filename)
		}
		return true
	})

	return nil
}

func (a *MiddlewareAudit) checkRouteDefinition(fset *token.FileSet, call *ast.CallExpr, filename string) {
	// 获取函数名
	var funcName string
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		funcName = sel.Sel.Name
	}

	// 检查是否是路由定义
	httpMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	isRouteMethod := false
	for _, method := range httpMethods {
		if funcName == method {
			isRouteMethod = true
			break
		}
	}

	if !isRouteMethod {
		return
	}

	pos := fset.Position(call.Pos())

	// 检查参数数量（至少应该有路径和处理函数）
	if len(call.Args) < 2 {
		return
	}

	// 提取路径
	var path string
	if lit, ok := call.Args[0].(*ast.BasicLit); ok {
		path = strings.Trim(lit.Value, "\"")
	}

	// 检查中间件使用
	middlewares := a.extractMiddlewares(call.Args[1:])

	// 审计规则
	a.auditRoute(filename, pos.Line, funcName, path, middlewares)
}

func (a *MiddlewareAudit) extractMiddlewares(args []ast.Expr) []string {
	var middlewares []string
	for _, arg := range args {
		if sel, ok := arg.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok {
				middlewares = append(middlewares, x.Name+"."+sel.Sel.Name)
			}
		} else if call, ok := arg.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok {
					middlewares = append(middlewares, x.Name+"."+sel.Sel.Name)
				}
			}
		}
	}
	return middlewares
}

func (a *MiddlewareAudit) auditRoute(file string, line int, method, path string, middlewares []string) {
	// 规则1: POST/PUT/PATCH 应该验证 Content-Type
	if (method == "POST" || method == "PUT" || method == "PATCH") &&
		!a.hasMiddleware(middlewares, "ValidateJSONContentType") &&
		!a.hasMiddleware(middlewares, "ValidateContentType") {
		a.issues = append(a.issues, AuditIssue{
			File:        file,
			Line:        line,
			Severity:    "MEDIUM",
			Category:    "Validation",
			Description: fmt.Sprintf("%s %s: Missing Content-Type validation", method, path),
			Suggestion:  "Add middleware.ValidateJSONContentType() before handler",
		})
	}

	// 规则2: 带 :id 的路由应该验证ID格式
	if strings.Contains(path, ":id") &&
		!a.hasMiddleware(middlewares, "ValidateMongoID") &&
		!a.hasMiddleware(middlewares, "ValidateUUID") {
		a.issues = append(a.issues, AuditIssue{
			File:        file,
			Line:        line,
			Severity:    "HIGH",
			Category:    "Security",
			Description: fmt.Sprintf("%s %s: Missing ID format validation", method, path),
			Suggestion:  "Add middleware.ValidateMongoID(\"id\") or middleware.ValidateUUID(\"id\")",
		})
	}

	// 规则3: DELETE 操作应该有权限检查
	if method == "DELETE" && !a.hasMiddleware(middlewares, "HasPermission") {
		a.issues = append(a.issues, AuditIssue{
			File:        file,
			Line:        line,
			Severity:    "HIGH",
			Category:    "Security",
			Description: fmt.Sprintf("DELETE %s: Missing permission check", path),
			Suggestion:  "Add middleware.HasPermission(\"resource:delete\")",
		})
	}

	// 规则4: 认证路由（/api下的）应该有 JWTAuth
	if strings.HasPrefix(path, "/api") &&
		!strings.Contains(path, "/public") &&
		!strings.Contains(path, "/login") &&
		!a.hasMiddleware(middlewares, "JWTAuth") {
		a.issues = append(a.issues, AuditIssue{
			File:        file,
			Line:        line,
			Severity:    "HIGH",
			Category:    "Security",
			Description: fmt.Sprintf("%s %s: Missing JWT authentication", method, path),
			Suggestion:  "Add middleware.JWTAuth() in route group",
		})
	}

	// 规则5: 列表查询应该验证分页参数
	if method == "GET" &&
		(strings.HasSuffix(path, "s") || strings.Contains(path, "/list")) &&
		!strings.Contains(path, ":id") &&
		!a.hasMiddleware(middlewares, "ValidatePagination") {
		a.issues = append(a.issues, AuditIssue{
			File:        file,
			Line:        line,
			Severity:    "LOW",
			Category:    "Validation",
			Description: fmt.Sprintf("GET %s: Missing pagination validation", path),
			Suggestion:  "Add middleware.ValidatePagination()",
		})
	}

	// 规则6: 敏感接口应该禁用缓存
	sensitivePaths := []string{"/profile", "/password", "/token", "/secret"}
	for _, sensitive := range sensitivePaths {
		if strings.Contains(path, sensitive) && !a.hasMiddleware(middlewares, "NoCache") {
			a.issues = append(a.issues, AuditIssue{
				File:        file,
				Line:        line,
				Severity:    "MEDIUM",
				Category:    "Security",
				Description: fmt.Sprintf("%s %s: Sensitive endpoint should disable cache", method, path),
				Suggestion:  "Add middleware.NoCache()",
			})
			break
		}
	}
}

func (a *MiddlewareAudit) hasMiddleware(middlewares []string, name string) bool {
	for _, m := range middlewares {
		if strings.Contains(m, name) {
			return true
		}
	}
	return false
}

func (a *MiddlewareAudit) printReport() {
	fmt.Println("\n=== Gin Middleware Audit Report ===")

	if len(a.issues) == 0 {
		fmt.Println("✅ No issues found! All routes are properly configured.")
		return
	}

	// 按严重程度分组
	highIssues := []AuditIssue{}
	mediumIssues := []AuditIssue{}
	lowIssues := []AuditIssue{}

	for _, issue := range a.issues {
		switch issue.Severity {
		case "HIGH":
			highIssues = append(highIssues, issue)
		case "MEDIUM":
			mediumIssues = append(mediumIssues, issue)
		case "LOW":
			lowIssues = append(lowIssues, issue)
		}
	}

	// 打印统计
	fmt.Printf("Total Issues: %d\n", len(a.issues))
	fmt.Printf("  🔴 High:   %d\n", len(highIssues))
	fmt.Printf("  🟡 Medium: %d\n", len(mediumIssues))
	fmt.Printf("  🟢 Low:    %d\n\n", len(lowIssues))

	// 打印详细信息
	a.printIssuesByCategory("🔴 HIGH SEVERITY ISSUES", highIssues)
	a.printIssuesByCategory("🟡 MEDIUM SEVERITY ISSUES", mediumIssues)
	a.printIssuesByCategory("🟢 LOW SEVERITY ISSUES", lowIssues)
}

func (a *MiddlewareAudit) printIssuesByCategory(title string, issues []AuditIssue) {
	if len(issues) == 0 {
		return
	}

	fmt.Printf("\n%s\n", title)
	fmt.Println(strings.Repeat("=", len(title)))

	for i, issue := range issues {
		fmt.Printf("\n%d. [%s] %s\n", i+1, issue.Category, issue.Description)
		fmt.Printf("   File: %s:%d\n", issue.File, issue.Line)
		fmt.Printf("   💡 Suggestion: %s\n", issue.Suggestion)
	}
}
