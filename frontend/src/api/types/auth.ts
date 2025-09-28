export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: number
  user: AccountResponse
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  gitlab_personal_access_token: string
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
  avatar?: string
  is_active: boolean
  force_password_reset: boolean
  password_initialized_at?: string
  last_login_at?: string
  created_at: string
  updated_at: string
  has_gitlab_personal_access_token: boolean
}

export interface CreateAccountRequest {
  username: string
  password: string
  email: string
  role?: string
  gitlab_personal_access_token?: string
}

export interface UpdateAccountRequest {
  email?: string
  role?: string
  is_active?: boolean
  gitlab_personal_access_token?: string
}
