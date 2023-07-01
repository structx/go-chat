
export interface UserState {
  uid: string
  email: string
  username: string
  created_at: string
  updated_at: string
  access_token: string
}

export interface NewUser {
  email: string
  username: string
  password: string
}

export interface User {
  uid: string
  email: string
  username: string
  created_at: string
}

export interface UpdateUser {
  uid: string
  email: string
  username: string
}
