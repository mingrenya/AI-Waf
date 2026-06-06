export interface ManagedUser {
    id: string
    username: string
    role: 'admin' | 'auditor' | 'configurator' | 'user'
    needReset: boolean
    createdAt?: string
    updatedAt?: string
    lastLogin?: string
    permissions?: string[]
}

export interface CreateUserRequest {
    username: string
    password: string
    role: 'admin' | 'auditor' | 'configurator' | 'user'
}

export interface UpdateUserRequest {
    username?: string
    password?: string
    role?: 'admin' | 'auditor' | 'configurator' | 'user'
    needReset?: boolean
}
