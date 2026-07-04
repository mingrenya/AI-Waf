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
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/rs/zerolog"
)

// ForensicsCapture 取证式流量捕获
// 根据攻击事件自动捕获相关流量（源IP + 目标 + 时间窗口），
// 为安全分析提供完整的攻击数据包以供下载分析。
type ForensicsCapture struct {
	repo      repository.CaptureRepository
	outputDir string
	logger    zerolog.Logger
	ctx       context.Context
	cancel    context.CancelFunc

	mu      sync.Mutex
	enabled bool
	handle  *pcap.Handle

	// 自动捕获配置
	autoCaptureEnabled bool
	captureWindow      time.Duration // 攻击事件前后时间窗口
	maxPacketsPerEvent int           // 每个事件最大抓包数
}

// ForensicsCaptureConfig 取证捕获配置
type ForensicsCaptureConfig struct {
	OutputDir          string
	AutoCaptureEnabled bool
	CaptureWindow      time.Duration
	MaxPacketsPerEvent int
}

// DefaultForensicsConfig 默认配置
func DefaultForensicsConfig() ForensicsCaptureConfig {
	return ForensicsCaptureConfig{
		OutputDir:          "",
		AutoCaptureEnabled: true,
		CaptureWindow:      30 * time.Second,
		MaxPacketsPerEvent: 5000,
	}
}

// NewForensicsCapture 创建取证捕获器
func NewForensicsCapture(cfg ForensicsCaptureConfig, repo repository.CaptureRepository) *ForensicsCapture {
	ctx, cancel := context.WithCancel(context.Background())

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir, _ = ResolveCapturePath()
	}

	return &ForensicsCapture{
		repo:               repo,
		outputDir:          outputDir,
		logger:             config.GetServiceLogger("forensics-capture"),
		ctx:                ctx,
		cancel:             cancel,
		autoCaptureEnabled: cfg.AutoCaptureEnabled,
		captureWindow:      cfg.CaptureWindow,
		maxPacketsPerEvent: cfg.MaxPacketsPerEvent,
	}
}

// Start 启动后台监听（常驻 pcap 流）
func (f *ForensicsCapture) Start(iface string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.handle != nil {
		return fmt.Errorf("forensics capture already running on interface %s", iface)
	}

	handle, err := pcap.OpenLive(iface, 65535, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("打开接口 %s 失败: %w", iface, err)
	}

	f.handle = handle
	f.enabled = true
	f.logger.Info().
		Str("interface", iface).
		Bool("auto_capture", f.autoCaptureEnabled).
		Msg("取证流量捕获已启动（后台常驻监听）")

	return nil
}

// Stop 停止后台监听
func (f *ForensicsCapture) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.cancel()
	if f.handle != nil {
		f.handle.Close()
		f.handle = nil
	}
	f.enabled = false
	f.logger.Info().Msg("取证流量捕获已停止")
}

// CaptureAttackTraffic 根据攻击事件自动捕获相关流量
// 输入：攻击事件核心信息（源IP、目标域名/端口、事件发生时间）
// 输出：自动创建 PCAP 会话并填充数据包，返回 session ID
func (f *ForensicsCapture) CaptureAttackTraffic(event AttackCaptureEvent) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.enabled || f.handle == nil {
		return "", fmt.Errorf("forensics capture not running")
	}

	// 构建 BPF 过滤器（捕获攻击源IP相关的所有流量）
	bpfFilter := fmt.Sprintf("host %s", event.SourceIP)
	if event.TargetPort > 0 {
		bpfFilter = fmt.Sprintf("host %s and port %d", event.SourceIP, event.TargetPort)
	}

	sessionID := uuid.New().String()
	fileName := fmt.Sprintf("forensics_%s_%s.pcap", event.SourceIP, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(f.outputDir, fileName)

	// 创建 PCAP 文件
	fh, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建PCAP文件失败: %w", err)
	}

	writer := pcapgo.NewWriter(fh)
	if err := writer.WriteFileHeader(65535, f.handle.LinkType()); err != nil {
		fh.Close()
		return "", fmt.Errorf("写入PCAP头失败: %w", err)
	}

	// 持久化会话记录
	now := time.Now()
	session := &repository.CaptureSession{
		ID:           sessionID,
		Interface:    event.Interface,
		BPFFilter:    bpfFilter,
		Status:       "capturing",
		MaxPackets:   f.maxPacketsPerEvent,
		DurationSecs: int(f.captureWindow.Seconds()),
		Description:  fmt.Sprintf("自动取证捕获 - 攻击源 %s (%s)", event.SourceIP, event.AttackType),
		FilePath:     filePath,
		CreatedAt:    now,
		StartedAt:    &now,
	}
	if err := f.repo.CreateSession(f.ctx, session); err != nil {
		fh.Close()
		return "", fmt.Errorf("创建捕获记录失败: %w", err)
	}

	// 异步捕获（后台 goroutine）
	go func() {
		defer fh.Close()

		ctx, cancel := context.WithTimeout(context.Background(), f.captureWindow)
		defer cancel()

		packetSource := gopacket.NewPacketSource(f.handle, f.handle.LinkType())
		packetCount := 0

		for {
			select {
			case <-ctx.Done():
				// 时间窗口结束
				finalizeSession(f.repo, sessionID, filePath, packetCount, "completed")
				f.logger.Info().
					Str("session_id", sessionID).
					Str("source_ip", event.SourceIP).
					Str("attack_type", event.AttackType).
					Int("packets", packetCount).
					Msg("取证捕获完成")
				return
			case packet, ok := <-packetSource.Packets():
				if !ok {
					finalizeSession(f.repo, sessionID, filePath, packetCount, "error")
					return
				}
				if err := writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
					f.logger.Warn().Err(err).Msg("写入数据包失败")
					continue
				}
				packetCount++

				// 每 100 个包更新一次统计
				if packetCount%100 == 0 {
					_ = f.repo.UpdateSession(f.ctx, sessionID, map[string]interface{}{
						"packet_count": packetCount,
					})
				}

				if f.maxPacketsPerEvent > 0 && packetCount >= f.maxPacketsPerEvent {
					finalizeSession(f.repo, sessionID, filePath, packetCount, "completed")
					return
				}
			}
		}
	}()

	return sessionID, nil
}

// AttackCaptureEvent 触发自动捕获的攻击事件
type AttackCaptureEvent struct {
	SourceIP   string `json:"source_ip"`
	TargetHost string `json:"target_host"`
	TargetPort int    `json:"target_port"`
	AttackType string `json:"attack_type"`
	Severity   string `json:"severity"`
	RuleID     int    `json:"rule_id"`
	RequestID  string `json:"request_id"`
	Interface  string `json:"interface"` // 网卡接口，默认 "eth0"
}

// IsEnabled 检查是否启用自动捕获
func (f *ForensicsCapture) IsEnabled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.enabled && f.autoCaptureEnabled
}

// EnableAutoCapture 启用/禁用自动捕获
func (f *ForensicsCapture) EnableAutoCapture(enable bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.autoCaptureEnabled = enable
	f.logger.Info().Bool("auto_capture", enable).Msg("自动取证捕获开关变更")
}

// Stats 返回当前统计信息
func (f *ForensicsCapture) Stats() map[string]interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	return map[string]interface{}{
		"enabled":      f.enabled,
		"auto_capture": f.autoCaptureEnabled,
		"output_dir":   f.outputDir,
	}
}

// finalizeSession 标记会话结束
func finalizeSession(repo repository.CaptureRepository, sessionID, filePath string, packetCount int, status string) {
	stoppedNow := time.Now()
	var fileSize int64
	if fi, err := os.Stat(filePath); err == nil {
		fileSize = fi.Size()
	}
	_ = repo.UpdateSession(context.Background(), sessionID, map[string]interface{}{
		"status":       status,
		"packet_count": packetCount,
		"file_size":    fileSize,
		"stopped_at":   &stoppedNow,
	})
}
