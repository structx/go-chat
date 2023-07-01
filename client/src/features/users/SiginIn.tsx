
import React, { useState, type ChangeEvent, type FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'

import { useAppDispatch } from '../../app/hooks'
import { useUserLoginMutation } from '../api/user'
import { setId, setAccessToken } from './userSlice'

import Container from '@mui/material/Container'
import Box from '@mui/material/Box'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'

interface SigninPageState {
  email: string
  password: string
}

export const SignInPage = (): JSX.Element => {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()

  const [state, setState] = useState<SigninPageState>({
    email: '',
    password: ''
  })

  const [userLogin, { isLoading, isError }] = useUserLoginMutation()

  const handleFormChange = (event: ChangeEvent<HTMLInputElement>): void => {
    setState({ ...state, [event.target.id]: event.target.value })
  }

  const handleSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault()

    userLogin({ username: state.email, password: state.password }).unwrap()
      .then((data) => {
        dispatch(setId(data.user_id))
        dispatch(setAccessToken(data.access_token))
        navigate('/profile')
      })
      .catch((err) => {
        return err
      })
  }

  if (isLoading) {
    return (
      <div>
        <p>attempting user authentication</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div>
        <p>error occured during user authentication</p>
      </div>
    )
  }

  return (
    <div>
      <Container>
        <Box component="form" onSubmit={handleSubmit}>
          <TextField
            id="email"
            name="email"
            label="Email Address"
            onChange={handleFormChange}
            required
            fullWidth
            autoFocus
          />
          <TextField
            id="password"
            name="password"
            type="password"
            label="Password"
            onChange={handleFormChange}
            required
            fullWidth
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
          >Sign In</Button>
        </Box>
        <br></br>
        <Link to='/register'>Create</Link>
      </Container>
    </div>
  )
}
