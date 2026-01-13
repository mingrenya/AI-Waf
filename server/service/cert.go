package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/dto"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/HUAHUAI23/RuiQi/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrCertificateNotFound   = errors.New("证书不存在")
	ErrCertificateNameExists = errors.New("证书名称已存在")
	ErrInvalidCertificate    = errors.New("无效的证书格式")
)

// CertificateService 证书服务接口
type CertificateService interface {
	CreateCertificate(ctx context.Context, req *dto.CertificateCreateRequest) (*model.CertificateStore, error)
	GetCertificates(ctx context.Context, pageStr, sizeStr string) ([]model.CertificateStore, int64, error)
	GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error)
	UpdateCertificate(ctx context.Context, id bson.ObjectID, req *dto.CertificateUpdateRequest) (*model.CertificateStore, error)
	DeleteCertificate(ctx context.Context, id bson.ObjectID) error
}

// CertificateServiceImpl 证书服务实现
type CertificateServiceImpl struct {
	certRepo repository.CertificateRepository
	logger   zerolog.Logger
}

// NewCertificateService 创建证书服务
func NewCertificateService(certRepo repository.CertificateRepository) CertificateService {
	logger := config.GetServiceLogger("certificate")
	return &CertificateServiceImpl{
		certRepo: certRepo,
		logger:   logger,
	}
}

// CreateCertificate 创建证书
func (s *CertificateServiceImpl) CreateCertificate(ctx context.Context, req *dto.CertificateCreateRequest) (*model.CertificateStore, error) {
	// 检查证书名称是否已存在
	if req.Name != "" {
		exists, err := s.certRepo.CheckCertificateNameExists(ctx, req.Name, bson.NilObjectID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCertificateNameExists
		}
	}

	// 创建新证书
	cert := model.NewCertificateStore()
	cert.Name = req.Name
	cert.Description = req.Description
	cert.PublicKey = req.PublicKey
	cert.PrivateKey = req.PrivateKey
	cert.ExpireDate = req.ExpireDate
	cert.IssuerName = req.IssuerName
	cert.FingerPrint = req.FingerPrint
	cert.Domains = req.Domains

	// 验证证书
	if err := model.ValidateCertificateStore(cert); err != nil {
		s.logger.Error().Err(err).Msg("证书验证失败")
		return nil, ErrInvalidCertificate
	}

	// 保存证书
	err := s.certRepo.CreateCertificate(ctx, cert)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建证书失败")
		return nil, err
	}

	s.logger.Info().Str("id", cert.ID.Hex()).Str("name", cert.Name).Msg("证书创建成功")
	return cert, nil
}

// GetCertificates 获取证书列表
func (s *CertificateServiceImpl) GetCertificates(ctx context.Context, pageStr, sizeStr string) ([]model.CertificateStore, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	certificates, total, err := s.certRepo.GetCertificates(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取证书列表失败")
		return nil, 0, err
	}

	// 转换为DTO
	responses := make([]model.CertificateStore, len(certificates))
	for i, cert := range certificates {
		responses[i] = model.CertificateStore{
			ID:          cert.ID,
			Name:        cert.Name,
			Description: cert.Description,
			PublicKey:   cert.PublicKey,
			PrivateKey:  cert.PrivateKey,
			ExpireDate:  cert.ExpireDate,
			IssuerName:  cert.IssuerName,
			FingerPrint: cert.FingerPrint,
			Domains:     cert.Domains,
			CreatedAt:   cert.CreatedAt,
			UpdatedAt:   cert.UpdatedAt,
		}
	}

	return responses, total, nil
}

// GetCertificateByID 根据ID获取证书
func (s *CertificateServiceImpl) GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error) {
	cert, err := s.certRepo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCertNotFound) {
			return nil, ErrCertificateNotFound
		}
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("获取证书失败")
		return nil, err
	}

	return cert, nil
}

// UpdateCertificate 更新证书
func (s *CertificateServiceImpl) UpdateCertificate(ctx context.Context, id bson.ObjectID, req *dto.CertificateUpdateRequest) (*model.CertificateStore, error) {
	// 获取现有证书
	cert, err := s.certRepo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCertNotFound) {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// 检查证书名称是否已存在（如果要更新名称）
	if req.Name != "" && req.Name != cert.Name {
		exists, err := s.certRepo.CheckCertificateNameExists(ctx, req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCertificateNameExists
		}
		cert.Name = req.Name
	}

	// 更新其他字段（只更新非空字段）
	if req.Description != "" {
		cert.Description = req.Description
	}
	if req.PublicKey != "" {
		cert.PublicKey = req.PublicKey
	}
	if req.PrivateKey != "" {
		cert.PrivateKey = req.PrivateKey
	}
	if !req.ExpireDate.IsZero() {
		cert.ExpireDate = req.ExpireDate
	}
	if req.IssuerName != "" {
		cert.IssuerName = req.IssuerName
	}
	if req.FingerPrint != "" {
		cert.FingerPrint = req.FingerPrint
	}
	if req.Domains != nil {
		cert.Domains = req.Domains
	}

	// 验证证书
	if err := model.ValidateCertificateStore(cert); err != nil {
		s.logger.Error().Err(err).Msg("证书验证失败")
		return nil, ErrInvalidCertificate
	}

	// 保存更新
	err = s.certRepo.UpdateCertificate(ctx, cert)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("更新证书失败")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", cert.Name).Msg("证书更新成功")
	return cert, nil
}

// DeleteCertificate 删除证书
func (s *CertificateServiceImpl) DeleteCertificate(ctx context.Context, id bson.ObjectID) error {
	// 检查证书是否存在
	_, err := s.certRepo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCertNotFound) {
			return ErrCertificateNotFound
		}
		return err
	}

	// 删除证书
	err = s.certRepo.DeleteCertificate(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除证书失败")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Msg("证书删除成功")
	return nil
}
