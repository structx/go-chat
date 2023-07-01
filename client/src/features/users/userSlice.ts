
import { type PayloadAction, createSlice } from '@reduxjs/toolkit'
import { type UserState } from './types'
import { type RootState } from '../../app/store'

const initialState: UserState = {
  uid: '',
  email: '',
  username: '',
  created_at: '',
  updated_at: '',
  access_token: ''
}

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setId: (state, action: PayloadAction<string>) => {
      state.uid = action.payload
    },
    setAccessToken: (state, action: PayloadAction<string>) => {
      state.access_token = action.payload
    },
    setEmail: (state, action: PayloadAction<string>) => {
      state.email = action.payload
    },
    setUsername: (state, action: PayloadAction<string>) => {
      state.username = action.payload
    },
    setCreatedAt: (state, action: PayloadAction<string>) => {
      state.created_at = action.payload
    },
    setUpdatedAt: (state, action: PayloadAction<string>) => {
      state.updated_at = action.payload
    }
  }
})

export const { setId, setAccessToken, setEmail, setUsername, setCreatedAt, setUpdatedAt } = userSlice.actions

export const selectId = (state: RootState): string => state.user.uid
export const selectAccessToken = (state: RootState): string => state.user.access_token

export default userSlice.reducer
