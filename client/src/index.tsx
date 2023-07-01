import React from 'react'
import ReactDOM from 'react-dom/client'
import {
  RouterProvider
} from 'react-router-dom'
import router from './routes'

import { CssBaseline } from '@mui/material'

import { Provider } from 'react-redux'
import { PersistGate } from 'redux-persist/integration/react'
import { setupStore } from './app/store'
import { persistStore } from 'redux-persist'

const store = setupStore()

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
)

root.render(
  <React.StrictMode>
    <Provider store={store}>
      <PersistGate loading={null} persistor={ persistStore(store) }>
        <CssBaseline>
          <RouterProvider router={router} />
        </CssBaseline>
      </PersistGate>
    </Provider>
  </React.StrictMode >
)
