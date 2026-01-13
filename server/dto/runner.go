package dto

// RunnerControlRequest 运行器控制请求
type RunnerControlRequest struct {
	Action string `json:"action" binding:"required,oneof=start stop restart force_stop reload"` // 控制动作
}

// RunnerControlResponse 运行器控制响应
type RunnerControlResponse struct {
	Success bool   `json:"success" example:"true"`     // 操作是否成功
	Action  string `json:"action" example:"start"`     // 执行的动作
	Message string `json:"message" example:"运行器已成功启动"` // 操作消息
	State   string `json:"state" example:"running"`    // 操作后的状态
}

// RunnerStatusResponse 运行器状态响应
type RunnerStatusResponse struct {
	State     string `json:"state" example:"running"`  // 状态：running, stopped, error
	IsRunning bool   `json:"isRunning" example:"true"` // 是否正在运行
}
