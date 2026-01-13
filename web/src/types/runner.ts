export type RunnerAction = 'start' | 'stop' | 'restart' | 'force_stop' | 'reload'

export interface RunnerControlRequest {
    action: RunnerAction
}

export interface RunnerControlResponse {
    action: string
    message: string
    state: string
    success: boolean
}

export interface RunnerStatusResponse {
    isRunning: boolean
    state: string
}