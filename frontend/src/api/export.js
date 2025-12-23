import axios from './axios'

export const exportApi = {
  // Get OpenAPI spec as JSON
  getOpenAPI: async (projectId) => {
    return axios.get(`/projects/${projectId}/export/openapi`)
  },

  // Download OpenAPI spec as file
  downloadOpenAPI: async (projectId) => {
    const response = await axios.get(`/projects/${projectId}/export/openapi?download=true`, {
      responseType: 'blob',
    })
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `openapi-${projectId.slice(0, 8)}.json`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  },

  // Get Postman collection as JSON
  getPostman: async (projectId) => {
    return axios.get(`/projects/${projectId}/export/postman`)
  },

  // Download Postman collection as file
  downloadPostman: async (projectId) => {
    const response = await axios.get(`/projects/${projectId}/export/postman?download=true`, {
      responseType: 'blob',
    })
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `postman-${projectId.slice(0, 8)}.json`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  },
}
