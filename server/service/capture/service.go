package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/uuid"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/rs/zerolog"
)

// CaptureService 流量捕获服务
type CaptureService struct {
	repo           repository.CaptureRepository
	outputDir      string
	activeCaptures sync.Map // map[string]context.CancelFunc
	logger         zerolog.Logger
}

// NewCaptureService 创建捕获服务
func NewCaptureService(repo repository.CaptureRepository, outputDir string) *CaptureService {
	return &CaptureService{
		repo:      repo,
		outputDir: outputDir,
		logger:    config.GetServiceLogger("capture"),
	}
}

// StartCapture 启动流量捕获
func (s *CaptureService) StartCapture(ctx context.Context, req dto.StartCaptureRequest) (*dto.CaptureSessionResponse, error) {
	iface := req.Interface
	if iface == "" {
		iface = "eth0"
	}

	// 验证接口是否存在
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口列表失败: %w", err)
	}
	ifaceExists := false
	for _, dev := range devs {
		if dev.Name == iface {
			ifaceExists = true
			break
		}
	}
	if !ifaceExists {
		return nil, fmt.Errorf("网络接口 %s 不存在", iface)
	}

	// 打开实时捕获
	handle, err := pcap.OpenLive(iface, 65535, true, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("打开接口 %s 失败: %w", iface, err)
	}

	// 设置 BPF 过滤器
	if req.BPFFilter != "" {
		if err := handle.SetBPFFilter(req.BPFFilter); err != nil {
			handle.Close()
			return nil, fmt.Errorf("设置 BPF 过滤器失败: %w", err)
		}
	}

	sessionID := uuid.New().String()
	filePath := filepath.Join(s.outputDir, "capture_"+sessionID+".pcap")

	now := time.Now()

	// 持久化会话
	session := &repository.CaptureSession{
		ID:           sessionID,
		Interface:    iface,
		BPFFilter:    req.BPFFilter,
		Status:       "running",
		MaxPackets:   req.MaxPackets,
		DurationSecs: req.DurationSecs,
		Description:  req.Description,
		FilePath:     filePath,
		CreatedAt:    now,
		StartedAt:    &now,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		handle.Close()
		return nil, fmt.Errorf("创建捕获会话失败: %w", err)
	}

	// 创建 PCAP 文件
	f, err := os.Create(filePath)
	if err != nil {
		handle.Close()
		s.repo.UpdateSession(ctx, sessionID, map[string]interface{}{
			"status":    "error",
			"error_msg": err.Error(),
		})
		return nil, fmt.Errorf("创建 PCAP 文件失败: %w", err)
	}

	// 写入 PCAP 头
	writer := pcapgo.NewWriter(f)
	if err := writer.WriteFileHeader(65535, handle.LinkType()); err != nil {
		f.Close()
		handle.Close()
		s.repo.UpdateSession(ctx, sessionID, map[string]interface{}{
			"status":    "error",
			"error_msg": err.Error(),
		})
		return nil, fmt.Errorf("写入 PCAP 头失败: %w", err)
	}

	// 启动捕获 goroutine
	captureCtx, cancel := context.WithCancel(context.Background())
	s.activeCaptures.Store(sessionID, cancel)

	go func() {
		defer func() {
			handle.Close()
			f.Close()
			s.activeCaptures.Delete(sessionID)

			// 获取文件大小
			fi, _ := os.Stat(filePath)
			finalStatus := "completed"
			if captureCtx.Err() != nil {
				finalStatus = "stopped"
			}
			updates := map[string]interface{}{
				"status": finalStatus,
			}
			stoppedNow := time.Now()
			updates["stopped_at"] = &stoppedNow
			if fi != nil {
				updates["file_size"] = fi.Size()
			}
			_ = s.repo.UpdateSession(context.Background(), sessionID, updates)
			s.logger.Info().
				Str("session_id", sessionID).
				Str("status", finalStatus).
				Msg("Capture session ended")
		}()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		packetCount := 0
		var timeoutCh <-chan time.Time
		if req.DurationSecs > 0 {
			timeoutCh = time.After(time.Duration(req.DurationSecs) * time.Second)
		}

		for {
			select {
			case <-captureCtx.Done():
				return
			case <-timeoutCh:
				return
			case packet, ok := <-packetSource.Packets():
				if !ok {
					return
				}
				if err := writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
					s.logger.Warn().Err(err).Msg("写入数据包失败")
					continue
				}
				packetCount++
				_ = s.repo.UpdateSession(context.Background(), sessionID, map[string]interface{}{
					"packet_count": packetCount,
				})
				if req.MaxPackets > 0 && packetCount >= req.MaxPackets {
					return
				}
			}
		}
	}()

	return s.sessionToResponse(session), nil
}

// StopCapture 停止流量捕获
func (s *CaptureService) StopCapture(ctx context.Context, sessionID string) error {
	cancelVal, ok := s.activeCaptures.Load(sessionID)
	if !ok {
		return fmt.Errorf("捕获会话 %s 未在运行", sessionID)
	}
	cancel := cancelVal.(context.CancelFunc)
	cancel()
	return nil
}

// GetSession 获取捕获会话
func (s *CaptureService) GetSession(ctx context.Context, sessionID string) (*dto.CaptureSessionResponse, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return s.sessionToResponse(session), nil
}

// ListSessions 列出捕获会话
func (s *CaptureService) ListSessions(ctx context.Context, skip, limit int64) ([]dto.CaptureSessionResponse, int64, error) {
	sessions, total, err := s.repo.ListSessions(ctx, skip, limit)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.CaptureSessionResponse, 0, len(sessions))
	for i := range sessions {
		resp = append(resp, *s.sessionToResponse(&sessions[i]))
	}
	return resp, total, nil
}

// GetCaptureFile 获取捕获文件路径
func (s *CaptureService) GetCaptureFile(sessionID string) (string, error) {
	session, err := s.repo.GetSession(context.Background(), sessionID)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(session.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("PCAP 文件不存在: %s", session.FilePath)
	}
	return session.FilePath, nil
}

func (s *CaptureService) sessionToResponse(session *repository.CaptureSession) *dto.CaptureSessionResponse {
	resp := &dto.CaptureSessionResponse{
		ID:           session.ID,
		Interface:    session.Interface,
		BPFFilter:    session.BPFFilter,
		Status:       session.Status,
		MaxPackets:   session.MaxPackets,
		DurationSecs: session.DurationSecs,
		Description:  session.Description,
		PacketCount:  session.PacketCount,
		FileSize:     session.FileSize,
		FilePath:     session.FilePath,
		CreatedAt:    session.CreatedAt.Format(time.RFC3339),
	}
	if session.StartedAt != nil {
		resp.StartedAt = session.StartedAt.Format(time.RFC3339)
	}
	if session.StoppedAt != nil {
		resp.StoppedAt = session.StoppedAt.Format(time.RFC3339)
	}
	resp.ErrorMsg = session.ErrorMsg
	return resp
}
