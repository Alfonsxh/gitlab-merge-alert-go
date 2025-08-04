export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: number
  user: AccountResponse
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface AccountResponse {
  id: number
  username: string
  email: string
  role: string
  is_active: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface CreateAccountRequest {
  username: string
  password: string
  email: string
  role?: string
}

export interface UpdateAccountRequest {
  email?: string
  role?: string
  is_active?: boolean
}