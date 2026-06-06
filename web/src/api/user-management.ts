import { del, get, post, put } from './index'
import type { CreateUserRequest, ManagedUser, UpdateUserRequest } from '@/types/user-management'

export const userManagementApi = {
    getUsers: (): Promise<ManagedUser[]> => {
        return get<ManagedUser[]>('/users')
    },

    createUser: (payload: CreateUserRequest): Promise<ManagedUser> => {
        return post<ManagedUser>('/users', payload)
    },

    updateUser: (id: string, payload: UpdateUserRequest): Promise<ManagedUser> => {
        return put<ManagedUser>(`/users/${id}`, payload)
    },

    deleteUser: (id: string): Promise<void> => {
        return del<void>(`/users/${id}`)
    },
}
