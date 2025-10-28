import axios from './axios'

export const entitiesApi = {
  // Get all entities for a project
  getByProject: async (projectId) => {
    return axios.get(`/projects/${projectId}/entities`)
  },

  // Get entity by ID
  getById: async (id) => {
    return axios.get(`/entities/${id}`)
  },

  // Create entity
  create: async (projectId, data) => {
    return axios.post(`/projects/${projectId}/entities`, data)
  },

  // Update entity
  update: async (id, data) => {
    return axios.put(`/entities/${id}`, data)
  },

  // Delete entity
  delete: async (id) => {
    return axios.delete(`/entities/${id}`)
  },
}
