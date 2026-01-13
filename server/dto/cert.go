package dto

import (
	"time"

	"github.com/HUAHUAI23/RuiQi/server/model"
)

// CertificateCreateRequest 创建证书请求
// @Description 创建证书的请求参数
type CertificateCreateRequest struct {
	Name        string    `json:"name" example:"example-cert"`               // 证书名称/别名
	Description string    `json:"description" example:"用于example.com的证书"`    // 证书描述
	PublicKey   string    `json:"publicKey" binding:"required"`              // 公钥内容（PEM格式）
	PrivateKey  string    `json:"privateKey" binding:"required"`             // 私钥内容（PEM格式）
	ExpireDate  time.Time `json:"expireDate" example:"2023-12-31T23:59:59Z"` // 证书过期日期
	IssuerName  string    `json:"issuerName" example:"Let's Encrypt"`        // 颁发机构
	FingerPrint string    `json:"fingerPrint" example:"AA:BB:CC:DD:..."`     // 证书指纹
	Domains     []string  `json:"domains" example:"[\"example.com\"]"`       // 证书绑定的域名列表
}

// CertificateUpdateRequest 更新证书请求
// @Description 更新证书的请求参数
type CertificateUpdateRequest struct {
	Name        string    `json:"name,omitempty" example:"example-cert"`               // 证书名称/别名
	Description string    `json:"description,omitempty" example:"用于example.com的证书"`    // 证书描述
	PublicKey   string    `json:"publicKey,omitempty"`                                 // 公钥内容（PEM格式）
	PrivateKey  string    `json:"privateKey,omitempty"`                                // 私钥内容（PEM格式）
	ExpireDate  time.Time `json:"expireDate,omitempty" example:"2023-12-31T23:59:59Z"` // 证书过期日期
	IssuerName  string    `json:"issuerName,omitempty" example:"Let's Encrypt"`        // 颁发机构
	FingerPrint string    `json:"fingerPrint,omitempty" example:"AA:BB:CC:DD:..."`     // 证书指纹
	Domains     []string  `json:"domains,omitempty" example:"[\"example.com\"]"`       // 证书绑定的域名列表
}

// CertificateListResponse 证书列表响应
// @Description 证书列表响应
type CertificateListResponse struct {
	Total int64                    `json:"total"` // 总数
	Items []model.CertificateStore `json:"items"` // 证书列表
}
