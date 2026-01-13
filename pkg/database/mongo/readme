package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/myapp/mongodb" // 假设上面的代码在这个包中
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// 用户文档结构
type User struct {
	ID       string    `bson:"_id,omitempty"`
	Username string    `bson:"username"`
	Email    string    `bson:"email"`
	Age      int       `bson:"age"`
	Created  time.Time `bson:"created"`
}

func main() {
	// 连接MongoDB
	uri := "mongodb://localhost:27017"
	client, err := mongodb.Connect(uri)
	if err != nil {
		log.Fatalf("连接MongoDB失败: %v", err)
	}

	// 确保在程序结束时断开连接
	defer func() {
		if err := mongodb.Disconnect(); err != nil {
			log.Printf("断开MongoDB连接失败: %v", err)
		}
	}()

	// 检查连接状态
	if err := mongodb.Ping(); err != nil {
		log.Fatalf("ping失败: %v", err)
	}
	fmt.Println("MongoDB连接成功!")

	// 获取数据库和集合
	coll, err := mongodb.GetCollection("testdb", "users")
	if err != nil {
		log.Fatalf("获取集合失败: %v", err)
	}

	// 插入文档
	if err := insertUser(coll); err != nil {
		log.Fatalf("插入用户失败: %v", err)
	}

	// 查询文档
	if err := findUsers(coll); err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	// 更新文档
	if err := updateUser(coll); err != nil {
		log.Fatalf("更新用户失败: %v", err)
	}

	// 删除文档
	if err := deleteUser(coll); err != nil {
		log.Fatalf("删除用户失败: %v", err)
	}

	// 演示重新连接功能
	fmt.Println("\n演示断开连接后重新连接:")
	if err := mongodb.Disconnect(); err != nil {
		log.Fatalf("断开连接失败: %v", err)
	}
	fmt.Println("已断开MongoDB连接")

	// 重新连接
	client, err = mongodb.Connect(uri)
	if err != nil {
		log.Fatalf("重新连接MongoDB失败: %v", err)
	}
	fmt.Println("成功重新连接到MongoDB!")

	// 验证连接是否有效
	if err := mongodb.Ping(); err != nil {
		log.Fatalf("重连后ping失败: %v", err)
	}
	fmt.Println("重连后连接状态正常")
}

// 插入用户文档
func insertUser(coll *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建用户文档
	user := User{
		Username: "zhangsan",
		Email:    "zhangsan@example.com",
		Age:      30,
		Created:  time.Now(),
	}

	// 插入文档
	result, err := coll.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("插入文档失败: %w", err)
	}
	fmt.Printf("插入用户成功，ID: %v\n", result.InsertedID)
	return nil
}

// 查询用户文档
func findUsers(coll *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建过滤器
	filter := bson.D{{Key: "age", Value: bson.D{{Key: "$gte", Value: 18}}}}

	// 执行查询
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 处理结果
	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return fmt.Errorf("解析查询结果失败: %w", err)
	}

	fmt.Printf("找到 %d 个用户:\n", len(users))
	for _, user := range users {
		fmt.Printf("- 用户: %s, 邮箱: %s, 年龄: %d\n", user.Username, user.Email, user.Age)
	}
	return nil
}

// 更新用户文档
func updateUser(coll *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建过滤器
	filter := bson.D{{Key: "username", Value: "zhangsan"}}

	// 创建更新内容
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "age", Value: 31}}}}

	// 执行更新
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新文档失败: %w", err)
	}

	fmt.Printf("更新用户成功，匹配: %d, 修改: %d\n", result.MatchedCount, result.ModifiedCount)
	return nil
}

// 删除用户文档
func deleteUser(coll *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建过滤器
	filter := bson.D{{Key: "username", Value: "zhangsan"}}

	// 执行删除
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}

	fmt.Printf("删除用户成功，删除数量: %d\n", result.DeletedCount)
	return nil
}