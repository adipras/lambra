import axios from './axios'

export const projectsApi = {
  // Get all projects
  getAll: async (params = {}) => {
    const { page = 1, limit = 20 } = params
    return axios.get('/projects', { params: { page, limit } })
  },

  // Get project by ID
  getById: async (id) => {
    return axios.get(`/projects/${id}`)
  },

  // Create project
  create: async (data) => {
    return axios.post('/projects', data)
  },

  // Update project
  update: async (id, data) => {
    return axios.put(`/projects/${id}`, data)
  },

  // Delete project
  delete: async (id) => {
    return axios.delete(`/projects/${id}`)
  },

  // Generate service
  generate: async (id) => {
    return axios.post(`/projects/${id}/generate`)
  },

  // Regenerate service
  regenerate: async (id) => {
    return axios.post(`/projects/${id}/regenerate`)
  },
}
