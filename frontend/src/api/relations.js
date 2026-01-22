import axios from './axios';

/**
 * Relations API Client
 * Handles all relation-related API calls
 */

// Create a new relation
export const createRelation = async (data) => {
  const response = await axios.post('/relations', data);
  return response.data;
};

// Get relation by ID
export const getRelationById = async (id) => {
  const response = await axios.get(`/relations/${id}`);
  return response.data;
};

// Get all relations for an entity
export const getRelationsByEntity = async (entityId) => {
  const response = await axios.get(`/entities/${entityId}/relations`);
  return response.data;
};

// Update a relation
export const updateRelation = async (id, data) => {
  const response = await axios.put(`/relations/${id}`, data);
  return response.data;
};

// Delete a relation
export const deleteRelation = async (id) => {
  const response = await axios.delete(`/relations/${id}`);
  return response.data;
};

// Get all relations for a project (for diagram view)
export const getRelationsByProject = async (projectId) => {
  const response = await axios.get(`/projects/${projectId}/relations`);
  return response.data;
};
