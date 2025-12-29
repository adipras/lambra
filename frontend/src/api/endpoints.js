import axios from './axios'

export const endpointsApi = {
  // Get all endpoints for a project
  getByProject: async (projectId) => {
    return axios.get(`/projects/${projectId}/endpoints`)
  },

  // Get all endpoints for an entity
  getByEntity: async (entityId) => {
    return axios.get(`/entities/${entityId}/endpoints`)
  },

  // Get endpoint by ID
  getById: async (id) => {
    return axios.get(`/endpoints/${id}`)
  },

  // Create endpoint
  create: async (data) => {
    return axios.post('/endpoints', data)
  },

  // Update endpoint
  update: async (id, data) => {
    return axios.put(`/endpoints/${id}`, data)
  },

  // Delete endpoint
  delete: async (id) => {
    return axios.delete(`/endpoints/${id}`)
  },

  // Test endpoint - sends a request to the deployed service
  test: async (id, { headers = {}, body = null, params = {} } = {}) => {
    return axios.post(`/endpoints/${id}/test`, {
      headers,
      body,
      params,
    })
  },
}
