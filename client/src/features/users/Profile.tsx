
import React, { type ChangeEvent, useState, type FormEvent } from 'react'

import { useAppDispatch, useAppSelector } from '../../app/hooks'

import { useFetchUserQuery, useUpdateUserMutation } from '../api/user'
import { Box, Button, TextField } from '@mui/material'
import Grid from '@mui/material/Unstable_Grid2/Grid2'
import { setEmail, setUsername } from './userSlice'

interface ProfileState {
  uid: string
  email: string
  username: string
  created_at: string
  updated_at: string
}

export const Profile = (): JSX.Element => {
  const dispatch = useAppDispatch()

  const uid = useAppSelector((state) => state.user.uid)
  const accessToken = useAppSelector((state) => state.user.access_token)

  const [profileState, setState] = useState<ProfileState>({
    uid,
    username: '',
    email: '',
    created_at: '',
    updated_at: ''
  })

  const { data, isFetching, isSuccess, isError, error } = useFetchUserQuery({ uid, access_token: accessToken })
  const [updateUser, { isLoading: updating, isError: updateError }] = useUpdateUserMutation()

  let content

  const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
    setState({ ...profileState, [event.target.id]: event.target.value })
  }

  const handleSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault()

    updateUser({
      update_user: {
        uid,
        email: profileState.email,
        username: profileState.username
      },
      access_token: accessToken
    }).unwrap()
      .then((data) => {
        dispatch(setEmail(data.user.email))
        dispatch(setUsername(data.user.username))
        setState({ ...profileState, email: data.user.email })
        setState({ ...profileState, username: data.user.username })
      })
      .catch((err) => {
        return err
      })
  }

  if (isSuccess) {
    content = (
      <Grid container spacing={2} smOffset='auto' mdOffset='auto'>
        <Grid xs={4} md={6}>
          <Box component='form' onSubmit={handleSubmit} sx={{ mt: 1 }}>
            <TextField
              margin='normal'
              required
              fullWidth
              id='email'
              label='Email Address'
              variant='filled'
              defaultValue={data.user.email}
              name='email'
              autoComplete='email'
              onChange={handleChange}
            />
            <TextField
              margin='normal'
              required
              fullWidth
              id='username'
              defaultValue={data.user.username}
              name='username'
              label='Username'
              variant='filled'
              onChange={handleChange}
            />
            <Button
              type='submit'
              variant='contained'
            >Submit</Button>
            <Button
              variant='outlined'
            >Cancel</Button>
          </Box>
        </Grid>
      </Grid>
    )
  } else if (isFetching) {
    content = <p>loading</p>
  } else if (isError) {
    if ('status' in error) {
      const errMsg = 'error' in error ? error.error : JSON.stringify(error.data)
      content = <p>{errMsg}</p>
    } else {
      content = <p>{error.message}</p>
    }
  } else if (updating) {
    content = <p>updating user</p>
  } else if (updateError) {
    content = <p>error occured updating user</p>
  }

  return (
    <div>
      {content}
    </div>
  )
}
