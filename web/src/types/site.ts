// 站点相关类型定义
export interface Site {
    id: string
    name: string
    domain: string
    listenPort: number
    enableHTTPS: boolean
    activeStatus: boolean
    wafEnabled: boolean
    wafMode: WAFMode
    backend: Backend
    certificate?: Certificate
    createdAt: string
    updatedAt: string
}

export enum WAFMode {
    Protection = "protection",
    Observation = "observation"
}

export interface Backend {
    servers: Server[]
}

export interface Server {
    host: string
    port: number
    isSSL: boolean
}

export interface Certificate {
    certName: string
    expireDate: string
    fingerPrint: string
    issuerName: string
    privateKey: string
    publicKey: string
}

// 请求和响应类型
export interface SiteListResponse {
    items: Site[]
    total: number
}

export interface CreateSiteRequest {
    name: string
    domain: string
    listenPort: number
    enableHTTPS: boolean
    activeStatus: boolean
    wafEnabled: boolean
    wafMode: WAFMode
    backend: Backend
    certificate?: Certificate
}

export interface UpdateSiteRequest {
    name?: string
    domain?: string
    listenPort?: number
    enableHTTPS?: boolean
    activeStatus?: boolean
    wafEnabled?: boolean
    wafMode?: WAFMode
    backend?: Backend
    certificate?: Certificate
} 