import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/useAuthStore'

// RequireAuth redirects unauthenticated visitors to /login,
// remembering where they were headed
export const RequireAuth = ({ children }) => {
  const token = useAuthStore((s) => s.token)
  const location = useLocation()

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }
  return children
}
