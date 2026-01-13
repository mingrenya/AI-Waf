// Request types
export interface UserLoginRequest {
    username: string
    password: string
}

export interface UserPasswordResetRequest {
    oldPassword: string
    newPassword: string
}

// Response types
export interface User {
    id: string
    username: string
    role: string
    needReset: boolean
    createdAt?: string
    updatedAt?: string
    lastLogin?: string
    permissions?: string[]
}

export interface LoginResponseData {
    token: string
    user: User
}

export interface GetUserInfoResponseData {
    id: string
    username: string
    role: string
    needReset: boolean
}

// Auth state
export interface AuthState {
    user: User | null
    token: string | null
    isAuthenticated: boolean
    needPasswordReset: boolean
    login: (token: string, user: User) => void
    logout: () => void
    setUser: (user: User) => void
} 