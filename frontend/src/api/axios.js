import axios from 'axios'

// In production, use relative URL (via nginx proxy)
// In development, use absolute URL to backend
const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const axiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
axiosInstance.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
axiosInstance.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    const message = error.response?.data?.message || error.message || 'Something went wrong'
    console.error('API Error:', message)

    // On 401: clear session and redirect to login — unless the failing
    // request is the login call itself (wrong password shows inline error)
    const isLoginRequest = error.config?.url?.includes('/auth/login')
    const status = error.response?.status
    if (status === 401 && !isLoginRequest && window.location.pathname !== '/login') {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }

    return Promise.reject(error)
  }
)

export default axiosInstance
