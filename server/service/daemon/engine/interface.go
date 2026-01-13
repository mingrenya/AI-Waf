package engine

import (
	"fmt"

	"github.com/HUAHUAI23/RuiQi/coraza-spoa/pkg/server"
	"github.com/rs/zerolog"
)

type EngineService interface {
	Start() error
	Restart() error
	Stop() error
	Reload() error
}

// NewEngineService 创建一个新的引擎服务实例
func NewEngineService(
	logger zerolog.Logger,
	mongoURI string,
) (EngineService, error) {
	// 创建并返回引擎服务
	agent, err := server.NewAgentServer(
		logger,
		mongoURI,
	)

	if err != nil {
		return nil, fmt.Errorf("创建AgentServer失败: %w", err)
	}

	engineService := &EngineServiceImpl{
		agent: agent,
	}

	return engineService, nil
}

type EngineServiceImpl struct {
	agent server.AgentServer
}

func (s *EngineServiceImpl) Start() error {
	return s.agent.Start()
}

func (s *EngineServiceImpl) Restart() error {
	return s.agent.Restart()
}

func (s *EngineServiceImpl) Stop() error {
	return s.agent.Stop()
}

func (s *EngineServiceImpl) Reload() error {
	return s.agent.UpdateApplications()
}
