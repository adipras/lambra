import axios from './axios'

export const snapshotsApi = {
  // Get all snapshots for a project
  getByProject: async (projectId, params = {}) => {
    const { page = 1, limit = 20 } = params
    return axios.get(`/projects/${projectId}/snapshots`, { params: { page, limit } })
  },

  // Create a new snapshot for a project
  create: async (projectId) => {
    return axios.post(`/projects/${projectId}/snapshots`)
  },

  // Get snapshot by ID
  getById: async (id) => {
    return axios.get(`/snapshots/${id}`)
  },

  // Get snapshot metadata
  getMetadata: async (id) => {
    return axios.get(`/snapshots/${id}/metadata`)
  },

  // Rollback to a snapshot
  rollback: async (id) => {
    return axios.post(`/snapshots/${id}/rollback`)
  },

  // Delete a snapshot
  delete: async (id) => {
    return axios.delete(`/snapshots/${id}`)
  },
}
