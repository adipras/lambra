import axios from './axios';

// Get all deployments for a project
export const getProjectDeployments = async (projectId, limit = 20, offset = 0) => {
  const response = await axios.get(`/projects/${projectId}/deployments`, {
    params: { limit, offset }
  });
  return response;
};

// Get latest deployment for a project
export const getLatestDeployment = async (projectId) => {
  const response = await axios.get(`/projects/${projectId}/deployments/latest`);
  return response;
};

// Get a single deployment by ID
export const getDeployment = async (deploymentId) => {
  const response = await axios.get(`/deployments/${deploymentId}`);
  return response;
};

// Get deployment logs
export const getDeploymentLogs = async (deploymentId, limit = 100, offset = 0) => {
  const response = await axios.get(`/deployments/${deploymentId}/logs`, {
    params: { limit, offset }
  });
  return response;
};

// Get container logs
export const getContainerLogs = async (projectId, tail = 100) => {
  const response = await axios.get(`/projects/${projectId}/container-logs`, {
    params: { tail }
  });
  return response;
};

// Create SSE connection for deployment logs streaming
export const createDeploymentLogStream = (deploymentId, onMessage, onError) => {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
  const eventSource = new EventSource(`${baseUrl}/deployments/${deploymentId}/logs/stream`);

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (e) {
      console.error('Failed to parse log message:', e);
    }
  };

  eventSource.onerror = (error) => {
    if (onError) onError(error);
    eventSource.close();
  };

  return eventSource;
};

// Create SSE connection for container logs streaming
export const createContainerLogStream = (projectId, onMessage, onError) => {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
  const eventSource = new EventSource(`${baseUrl}/projects/${projectId}/container-logs/stream`);

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (e) {
      console.error('Failed to parse log message:', e);
    }
  };

  eventSource.onerror = (error) => {
    if (onError) onError(error);
    eventSource.close();
  };

  return eventSource;
};
