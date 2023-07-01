import React from 'react'

import Grid from '@mui/material/Unstable_Grid2'
import Container from '@mui/material/Container'

import { SignInPage } from './features/users/SiginIn'

export const App = (): JSX.Element => {
  return (
    <div className="App">
      <Grid container >
        <Grid xs display="flex" justifyContent="center" alignItems="center">
          <Container>
            <SignInPage />
          </Container>
        </Grid>
      </Grid>

    </div>
  )
}

export default App
