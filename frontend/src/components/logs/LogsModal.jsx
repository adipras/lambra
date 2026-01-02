import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import LogViewer from './LogViewer';
import { getLatestDeployment, getContainerLogs, createContainerLogStream, createDeploymentLogStream } from '../../api/deployments';

export default function LogsModal({ projectId, isOpen, onClose }) {
  const [activeTab, setActiveTab] = useState('deployment');
  const [containerLogs, setContainerLogs] = useState([]);
  const [deploymentLogs, setDeploymentLogs] = useState([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamError, setStreamError] = useState(null);

  // Get latest deployment
  const { data: latestDeployment, refetch: refetchDeployment } = useQuery({
    queryKey: ['latestDeployment', projectId],
    queryFn: () => getLatestDeployment(projectId),
    enabled: isOpen && !!projectId,
  });

  // Get container logs (non-streaming)
  const { data: containerLogsData, refetch: refetchContainerLogs } = useQuery({
    queryKey: ['containerLogs', projectId],
    queryFn: () => getContainerLogs(projectId, 200),
    enabled: isOpen && !!projectId && activeTab === 'container' && !isStreaming,
    refetchInterval: isStreaming ? false : 5000,
  });

  // Initialize logs from queries
  useEffect(() => {
    if (latestDeployment?.data?.logs) {
      setDeploymentLogs(latestDeployment.data.logs);
    }
  }, [latestDeployment]);

  useEffect(() => {
    if (containerLogsData?.data?.logs) {
      // Parse container logs string into array
      const logsString = containerLogsData.data.logs;
      const lines = logsString.split('\n').filter(line => line.trim());
      setContainerLogs(lines.map((line, index) => ({
        id: index,
        message: line,
        stream: 'stdout',
        timestamp: new Date().toISOString(),
      })));
    }
  }, [containerLogsData]);

  // Stream logs
  useEffect(() => {
    if (!isOpen || !isStreaming) return;

    let eventSource;

    if (activeTab === 'container' && projectId) {
      eventSource = createContainerLogStream(
        projectId,
        (log) => {
          setContainerLogs(prev => [...prev, log]);
        },
        (error) => {
          setStreamError('Stream connection lost');
          setIsStreaming(false);
        }
      );
    } else if (activeTab === 'deployment' && latestDeployment?.data?.id) {
      eventSource = createDeploymentLogStream(
        latestDeployment.data.id,
        (log) => {
          setDeploymentLogs(prev => {
            // Avoid duplicates
            if (prev.find(l => l.id === log.id)) return prev;
            return [...prev, log];
          });
        },
        (error) => {
          setStreamError('Stream connection lost');
          setIsStreaming(false);
        }
      );
    }

    return () => {
      if (eventSource) {
        eventSource.close();
      }
    };
  }, [isOpen, isStreaming, activeTab, projectId, latestDeployment]);

  // Cleanup on close
  useEffect(() => {
    if (!isOpen) {
      setIsStreaming(false);
      setStreamError(null);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleRefresh = () => {
    if (activeTab === 'deployment') {
      refetchDeployment();
    } else {
      refetchContainerLogs();
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:p-0">
        {/* Backdrop */}
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={onClose}></div>

        {/* Modal */}
        <div className="relative bg-white rounded-lg shadow-xl transform transition-all sm:max-w-4xl sm:w-full">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-semibold text-gray-900">Service Logs</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-500">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Tabs */}
          <div className="flex border-b border-gray-200">
            <button
              className={`px-6 py-3 text-sm font-medium ${activeTab === 'deployment' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-gray-700'}`}
              onClick={() => setActiveTab('deployment')}
            >
              Deployment Logs
            </button>
            <button
              className={`px-6 py-3 text-sm font-medium ${activeTab === 'container' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-gray-700'}`}
              onClick={() => setActiveTab('container')}
            >
              Container Logs
            </button>
          </div>

          {/* Controls */}
          <div className="flex items-center justify-between px-6 py-3 bg-gray-50">
            <div className="flex items-center gap-2">
              <button
                onClick={() => setIsStreaming(!isStreaming)}
                className={`px-3 py-1.5 text-sm font-medium rounded ${isStreaming ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'}`}
              >
                {isStreaming ? (
                  <>
                    <span className="w-2 h-2 bg-red-500 rounded-full inline-block mr-2 animate-pulse"></span>
                    Stop Streaming
                  </>
                ) : (
                  <>
                    <span className="w-2 h-2 bg-green-500 rounded-full inline-block mr-2"></span>
                    Start Streaming
                  </>
                )}
              </button>
              <button
                onClick={handleRefresh}
                disabled={isStreaming}
                className="px-3 py-1.5 text-sm font-medium bg-gray-100 text-gray-700 rounded hover:bg-gray-200 disabled:opacity-50"
              >
                Refresh
              </button>
            </div>

            {streamError && (
              <span className="text-sm text-red-500">{streamError}</span>
            )}
          </div>

          {/* Content */}
          <div className="p-6">
            {activeTab === 'deployment' ? (
              latestDeployment?.data ? (
                <div className="space-y-4">
                  <div className="flex items-center justify-between text-sm text-gray-500">
                    <span>
                      Deployment: {latestDeployment.data.version} ({latestDeployment.data.status})
                    </span>
                    <span>
                      Started: {new Date(latestDeployment.data.started_at).toLocaleString()}
                    </span>
                  </div>
                  <LogViewer
                    logs={deploymentLogs}
                    isStreaming={isStreaming}
                    title="Deployment Logs"
                    type="deployment"
                    maxHeight="500px"
                  />
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  No deployment found. Deploy your service first.
                </div>
              )
            ) : (
              <LogViewer
                logs={containerLogs}
                isStreaming={isStreaming}
                title="Container Logs"
                type="container"
                maxHeight="500px"
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
