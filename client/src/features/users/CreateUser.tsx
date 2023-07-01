
import React, { useState, type ChangeEvent, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch } from '../../app/hooks'
import { type NewUser } from './types'
import { useCreateUserMutation } from '../api/user'
import { setId, setAccessToken } from './userSlice'

export const CreateUserPage = (): JSX.Element => {
  const dispatch = useAppDispatch()

  const [state, setState] = useState<NewUser>({
    email: '',
    username: '',
    password: ''
  })

  const navigate = useNavigate()
  const [createUser, { isLoading, isError }] = useCreateUserMutation()

  function handleChange (event: ChangeEvent<HTMLInputElement>): void {
    setState({ ...state, [event.target.id]: event.target.value })
  }

  const handleSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault()

    const newUser: NewUser = {
      email: state.email,
      username: state.username,
      password: state.password
    }

    createUser({ new_user: newUser }).unwrap()
      .then((data) => {
        // dispatch events to persist user id and access token
        dispatch(setId(data.user.uid))
        dispatch(setAccessToken(data.access_token))

        // navigate to dashboard
        navigate('/dashboard')
      })
      .catch((err) => {
        return err
      })
  }

  if (isLoading) {
    return (
      <div>
        <p>creating new user</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div>
        <p>an error occured</p>
      </div>
    )
  }

  return (
    <div>
      <form onSubmit={handleSubmit}>
        <input type='text' id='username' name='username' onChange={handleChange} placeholder='Username' />
        <br></br>
        <input type='email' id='email' name='email' onChange={handleChange} placeholder='Email Address' />
        <br></br>
        <input type='password' id='password' name='password' onChange={handleChange} placeholder='Password' />
        <br></br>
        <input type='password' id='confirm_password' name='password' onChange={handleChange} placeholder='Confirm Password' />
        <br></br>
        <input type='submit' value='Register' />
      </form>
    </div>
  )
}
