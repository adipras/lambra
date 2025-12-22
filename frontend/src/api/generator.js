import axios from './axios'

export const generatorApi = {
  // Preview generated code for an entity
  previewEntity: async (entityId) => {
    return axios.get(`/generate/preview/${entityId}`)
  },

  // Get list of files to be generated
  getFilesList: async (entityId) => {
    return axios.get(`/generate/files/${entityId}`)
  },

  // Generate code for an entity
  generateEntity: async (entityId, outputDir = './generated') => {
    return axios.post('/generate/entity', {
      entity_id: entityId,
      output_dir: outputDir,
    })
  },

  // Generate code for entire project
  generateProject: async (projectId, outputDir = './generated') => {
    return axios.post('/generate/project', {
      project_id: projectId,
      output_dir: outputDir,
    })
  },
}
