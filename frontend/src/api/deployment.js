import axios from './axios'

export const deploymentApi = {
  // Deploy a project (generate code + start Docker container)
  // Returns { deployment_id, url, port, ... }
  // Options: { reset_database: boolean } - if true, drops all tables before creating
  // Note: Deploy can take a long time (building Docker images), so we use extended timeout
  deploy: async (projectId, options = {}) => {
    return axios.post(`/projects/${projectId}/deploy`, options, {
      timeout: 300000, // 5 minutes timeout for deploy operations
    })
  },

  // Create SSE connection for deployment progress streaming
  createDeployProgressStream: (deploymentId, onMessage, onError, onComplete) => {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
    const eventSource = new EventSource(`${baseUrl}/deployments/${deploymentId}/logs/stream`)

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        onMessage?.(data)

        // Check for completion
        if (data.step === 'complete' && data.level === 'info') {
          onComplete?.()
        }
      } catch (e) {
        console.error('Failed to parse deployment log:', e)
      }
    }

    eventSource.onerror = (error) => {
      onError?.(error)
      eventSource.close()
    }

    return eventSource
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
    return axios.post(`/projects/${projectId}/redeploy`, {}, {
      timeout: 300000, // 5 minutes timeout for redeploy operations
    })
  },

  // Destroy a service completely (stop containers, remove volumes, delete workspace)
  destroy: async (projectId) => {
    return axios.delete(`/projects/${projectId}/destroy`, {
      timeout: 120000, // 2 minutes timeout for destroy operations
    })
  },

  // Get service deployment status
  getStatus: async (projectId) => {
    return axios.get(`/projects/${projectId}/status`)
  },
}
