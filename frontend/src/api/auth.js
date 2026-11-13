import axios from './axios'

export const authApi = {
  login: (data) => axios.post('/auth/login', data),
  me: () => axios.get('/auth/me'),
}
