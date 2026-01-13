package service

import (
	"context"
	"strconv"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/dto"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/HUAHUAI23/RuiQi/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SiteService interface {
	CreateSite(ctx context.Context, req *dto.CreateSiteRequest) (*model.Site, error)
	GetSites(ctx context.Context, pageStr, sizeStr string) ([]model.Site, int64, error)
	GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error)
	UpdateSite(ctx context.Context, id bson.ObjectID, req *dto.UpdateSiteRequest) (*model.Site, error)
	DeleteSite(ctx context.Context, id bson.ObjectID) error
}

// SiteService 站点服务
type SiteServiceImpl struct {
	siteRepo repository.SiteRepository
	logger   zerolog.Logger
}

// NewSiteService 创建站点服务
func NewSiteService(siteRepo repository.SiteRepository) SiteService {
	logger := config.GetServiceLogger("site")
	return &SiteServiceImpl{
		siteRepo: siteRepo,
		logger:   logger,
	}
}

// CreateSite 创建站点
func (s *SiteServiceImpl) CreateSite(ctx context.Context, req *dto.CreateSiteRequest) (*model.Site, error) {
	// 创建新站点
	site := model.NewSite()
	site.Name = req.Name
	site.Domain = req.Domain
	site.ListenPort = req.ListenPort
	site.EnableHTTPS = req.EnableHTTPS
	site.WAFEnabled = req.WAFEnabled
	site.WAFMode = model.WAFModeFromString(req.WAFMode)
	site.ActiveStatus = req.ActiveStatus
	// 设置后端服务器
	site.Backend.Servers = make([]model.Server, len(req.Backend.Servers))
	for i, server := range req.Backend.Servers {
		site.Backend.Servers[i] = model.Server{
			Host: server.Host,
			Port: server.Port,
		}
	}

	// 如果启用HTTPS，设置证书信息
	if req.EnableHTTPS && req.Certificate != nil {
		site.Certificate = model.Certificate{
			CertName:    req.Certificate.CertName,
			PublicKey:   req.Certificate.PublicKey,
			PrivateKey:  req.Certificate.PrivateKey,
			ExpireDate:  req.Certificate.ExpireDate,
			IssuerName:  req.Certificate.IssuerName,
			FingerPrint: req.Certificate.FingerPrint,
		}
	}

	// 验证站点配置
	if err := model.ValidateSite(site); err != nil {
		s.logger.Error().Err(err).Msg("站点验证失败")
		return nil, err
	}

	// 检查域名和端口是否已存在
	err := s.siteRepo.CheckDomainPortExists(ctx, site)
	if err != nil {
		return nil, err
	}

	// 保存站点
	err = s.siteRepo.CreateSite(ctx, site)
	if err != nil {
		s.logger.Error().Err(err).Msg("创建站点失败")
		return nil, err
	}

	s.logger.Info().Str("name", site.Name).Str("domain", site.Domain).Msg("站点创建成功")
	return site, nil
}

// GetSites 获取站点列表
func (s *SiteServiceImpl) GetSites(ctx context.Context, pageStr, sizeStr string) ([]model.Site, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	sites, total, err := s.siteRepo.GetSites(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取站点列表失败")
		return nil, 0, err
	}

	return sites, total, nil
}

// GetSiteByID 根据ID获取站点
func (s *SiteServiceImpl) GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error) {
	site, err := s.siteRepo.GetSiteByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return site, nil
}

// UpdateSite 更新站点
func (s *SiteServiceImpl) UpdateSite(ctx context.Context, id bson.ObjectID, req *dto.UpdateSiteRequest) (*model.Site, error) {

	err := s.siteRepo.CheckDomainPortConflict(ctx, &model.Site{
		ID:         id,
		Domain:     req.Domain,
		ListenPort: req.ListenPort,
	})

	if err != nil {
		return nil, err
	}

	// 获取现有站点
	site, err := s.siteRepo.GetSiteByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新站点信息
	if req.Name != "" {
		site.Name = req.Name
	}
	if req.Domain != "" {
		site.Domain = req.Domain
	}
	if req.ListenPort != 0 {
		site.ListenPort = req.ListenPort
	}

	// 更新HTTPS设置
	site.EnableHTTPS = req.EnableHTTPS
	site.WAFEnabled = req.WAFEnabled
	if req.WAFMode != "" {
		site.WAFMode = model.WAFModeFromString(req.WAFMode)
	}
	site.ActiveStatus = req.ActiveStatus

	// 更新后端服务器
	if req.Backend != nil && len(req.Backend.Servers) > 0 {
		site.Backend.Servers = make([]model.Server, len(req.Backend.Servers))
		for i, server := range req.Backend.Servers {
			site.Backend.Servers[i] = model.Server{
				Host:  server.Host,
				Port:  server.Port,
				IsSSL: server.IsSSL,
			}
		}
	}

	// 更新证书信息
	if req.EnableHTTPS && req.Certificate != nil {
		site.Certificate = model.Certificate{
			CertName:    req.Certificate.CertName,
			PublicKey:   req.Certificate.PublicKey,
			PrivateKey:  req.Certificate.PrivateKey,
			ExpireDate:  req.Certificate.ExpireDate,
			IssuerName:  req.Certificate.IssuerName,
			FingerPrint: req.Certificate.FingerPrint,
		}
	}

	// 验证站点配置
	if err := model.ValidateSite(site); err != nil {
		s.logger.Error().Err(err).Msg("站点验证失败")
		return nil, err
	}

	// 保存更新
	err = s.siteRepo.UpdateSite(ctx, site)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("更新站点失败")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", site.Name).Msg("站点更新成功")
	return site, nil
}

// DeleteSite 删除站点
func (s *SiteServiceImpl) DeleteSite(ctx context.Context, id bson.ObjectID) error {
	// 检查站点是否存在
	site, err := s.siteRepo.GetSiteByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除站点
	err = s.siteRepo.DeleteSite(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除站点失败")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", site.Name).Msg("站点删除成功")
	return nil
}
