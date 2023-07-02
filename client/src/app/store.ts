
import { type Store, type PreloadedState, combineReducers, configureStore } from '@reduxjs/toolkit'
import { persistReducer } from 'redux-persist'
import AsyncStorage from '@react-native-async-storage/async-storage'

import userReducer from '../features/users/userSlice'
import { userApi } from '../features/api/user'

import { REACT_APP_REDUX_STORE_KEY } from '../env'

const rootReducer = combineReducers({
  user: userReducer,
  [userApi.reducerPath]: userApi.reducer
})

const persistConfig = {
  // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
  key: `${REACT_APP_REDUX_STORE_KEY}`,
  storage: AsyncStorage
}

const persistedReducer = persistReducer(persistConfig, rootReducer)

export function setupStore (preloadedState?: PreloadedState<RootState>): Store {
  return configureStore({
    reducer: persistedReducer,
    preloadedState,
    middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(userApi.middleware)
  })
}

export type RootState = ReturnType<typeof rootReducer>
export type AppStore = ReturnType<typeof setupStore>
export type AppDispatch = AppStore['dispatch']
