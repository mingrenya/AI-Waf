package model

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CertificateStore 代表证书库表
type CertificateStore struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // 证书ID
	Name        string        `bson:"name" json:"name"`                  // 证书名称/别名
	Description string        `bson:"description" json:"description"`    // 证书描述
	PublicKey   string        `bson:"publicKey" json:"publicKey"`        // 公钥内容（PEM格式）
	PrivateKey  string        `bson:"privateKey" json:"privateKey"`      // 私钥内容（PEM格式）
	ExpireDate  time.Time     `bson:"expireDate" json:"expireDate"`      // 证书过期日期
	IssuerName  string        `bson:"issuerName" json:"issuerName"`      // 颁发机构
	FingerPrint string        `bson:"fingerPrint" json:"fingerPrint"`    // 证书指纹
	Domains     []string      `bson:"domains" json:"domains"`            // 证书绑定的域名列表
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`        // 创建时间
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`        // 更新时间
}

// GetCollectionName 返回集合名称
func (c *CertificateStore) GetCollectionName() string {
	return "certificate"
}

// NewCertificateStore 创建一个新证书，设置默认值
func NewCertificateStore() *CertificateStore {
	now := time.Now()
	return &CertificateStore{
		Domains:   make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ValidateCertificateStore 验证证书配置有效性
func ValidateCertificateStore(cert *CertificateStore) error {
	// 只检查必需字段
	if cert.PublicKey == "" || cert.PrivateKey == "" {
		return ErrMissingRequiredField
	}
	return nil
}

// 通用错误
var (
	ErrMissingRequiredField = errors.New("缺少必填字段")
	ErrInvalidFormat        = errors.New("无效的格式")
)
