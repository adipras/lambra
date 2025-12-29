import axios from './axios'

export const deploymentApi = {
  // Deploy a project (generate code + start Docker container)
  deploy: async (projectId) => {
    return axios.post(`/projects/${projectId}/deploy`)
  },

  // Start a deployed service
  start: async (projectId) => {
    return axios.post(`/projects/${projectId}/start`)
  },

  // Stop a running service
  stop: async (projectId) => {
    return axios.post(`/projects/${projectId}/stop`)
  },

  // Redeploy a service (down + up for cache clearing)
  redeploy: async (projectId) => {
    return axios.post(`/projects/${projectId}/redeploy`)
  },

  // Destroy a service completely (stop containers, remove volumes, delete workspace)
  destroy: async (projectId) => {
    return axios.delete(`/projects/${projectId}/destroy`)
  },

  // Get service deployment status
  getStatus: async (projectId) => {
    return axios.get(`/projects/${projectId}/status`)
  },
}
