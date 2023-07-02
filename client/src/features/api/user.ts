
import { REACT_APP_CHAT_SERVICE_URL } from '../../env'

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

import type { UpdateUser, NewUser, User } from '../users/types'

export interface CreateUserRequest {
  new_user: NewUser
}

export interface CreateUserResponse {
  user: User
  access_token: string
}

export interface UserLoginRequest {
  username: string
  password: string
}

export interface UserLoginResponse {
  user_id: string
  access_token: string
}

export interface FetchUserRequest {
  uid: string
  access_token: string
}

export interface FetchUserResponse {
  user: User
}

export interface UpdateUserRequest {
  update_user: UpdateUser
  access_token: string
}

export interface UpdateUserResponse {
  user: User
}

export interface DeleteUserRequest {
  uid: string
}

export interface DeleteUserResponse {
  message: string
}

export const userApi = createApi({
  baseQuery: fetchBaseQuery({
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    baseUrl: `${REACT_APP_CHAT_SERVICE_URL}/api/v1`,
    mode: 'cors',
    prepareHeaders: (headers) => {
      headers.set('Content-Type', 'application/json')
      return headers
    }
  }),
  endpoints: (builder) => ({

    createUser: builder.mutation<CreateUserResponse, CreateUserRequest>({
      query: (request) => ({
        url: '/user',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: request
      }),
      transformResponse: (value: CreateUserResponse, meta, arg) => value
    }),

    userLogin: builder.mutation<UserLoginResponse, UserLoginRequest>({
      query: (request) => ({
        url: '/user/login',
        method: 'POST',
        headers: {
          'content-type': 'application/json'
        },
        body: JSON.stringify(request)
      }),
      transformResponse: (value: UserLoginResponse, meta, arg) => value
    }),

    fetchUser: builder.query<FetchUserResponse, FetchUserRequest>({
      query: (fetchUser) => ({
        url: `/user/${fetchUser.uid}`,
        method: 'GET',
        headers: {
          Authorization: `Bearer: ${fetchUser.access_token}`,
          'Content-Type': 'application/json'
        }
      })
    }),

    updateUser: builder.mutation<UpdateUserResponse, UpdateUserRequest>({
      query: (updateUser) => ({
        url: '/user',
        method: 'PUT',
        headers: {
          Authorization: `Bearer: ${updateUser.access_token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(updateUser)
      })
    }),

    deleteUser: builder.query<DeleteUserResponse, DeleteUserRequest>({
      query: (req) => ({
        url: `/user/${req.uid}`,
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json'
        }
      })
    })
  })
})

export const {
  useCreateUserMutation,
  useFetchUserQuery,
  useUpdateUserMutation,
  useDeleteUserQuery,
  useUserLoginMutation
} = userApi
