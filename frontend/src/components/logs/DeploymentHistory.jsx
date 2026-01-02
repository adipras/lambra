import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getProjectDeployments, getDeployment } from '../../api/deployments';
import LogViewer from './LogViewer';

const StatusBadge = ({ status }) => {
  const styles = {
    pending: 'bg-gray-100 text-gray-800',
    deploying: 'bg-blue-100 text-blue-800',
    success: 'bg-green-100 text-green-800',
    failed: 'bg-red-100 text-red-800',
  };

  return (
    <span className={`px-2 py-1 text-xs font-medium rounded-full ${styles[status] || styles.pending}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
};

const formatDuration = (startedAt, completedAt) => {
  if (!completedAt) return 'In progress...';

  const start = new Date(startedAt);
  const end = new Date(completedAt);
  const diff = Math.floor((end - start) / 1000);

  if (diff < 60) return `${diff}s`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ${diff % 60}s`;
  return `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`;
};

const DeploymentRow = ({ deployment, isExpanded, onToggle }) => {
  const { data: details, isLoading } = useQuery({
    queryKey: ['deployment', deployment.id],
    queryFn: () => getDeployment(deployment.id),
    enabled: isExpanded,
  });

  const startTime = new Date(deployment.started_at).toLocaleString();
  const duration = formatDuration(deployment.started_at, deployment.completed_at);

  return (
    <div className="border-b border-gray-200 last:border-b-0">
      {/* Row header */}
      <div
        className="flex items-center justify-between p-4 hover:bg-gray-50 cursor-pointer"
        onClick={onToggle}
      >
        <div className="flex items-center gap-4">
          <button className="text-gray-400">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className={`h-5 w-5 transform transition-transform ${isExpanded ? 'rotate-90' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
          <div>
            <p className="font-medium text-gray-900">{deployment.version}</p>
            <p className="text-sm text-gray-500">{startTime}</p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-500">{duration}</span>
          <StatusBadge status={deployment.status} />
        </div>
      </div>

      {/* Expanded content */}
      {isExpanded && (
        <div className="px-4 pb-4 bg-gray-50">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <svg className="animate-spin h-6 w-6 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            </div>
          ) : details?.data ? (
            <div className="space-y-4">
              {/* Deployment info */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-gray-500">Environment</p>
                  <p className="font-medium text-gray-900">{deployment.environment}</p>
                </div>
                <div>
                  <p className="text-gray-500">Deployed By</p>
                  <p className="font-medium text-gray-900">{deployment.deployed_by || 'System'}</p>
                </div>
                {deployment.deployment_url && (
                  <div>
                    <p className="text-gray-500">URL</p>
                    <a
                      href={deployment.deployment_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-medium text-blue-600 hover:underline"
                    >
                      {deployment.deployment_url}
                    </a>
                  </div>
                )}
                {deployment.error_message && (
                  <div className="col-span-2 md:col-span-4">
                    <p className="text-gray-500">Error</p>
                    <p className="font-medium text-red-600">{deployment.error_message}</p>
                  </div>
                )}
              </div>

              {/* Deployment logs */}
              {details.data.logs && details.data.logs.length > 0 && (
                <LogViewer
                  logs={details.data.logs}
                  title="Deployment Logs"
                  type="deployment"
                  maxHeight="300px"
                />
              )}
            </div>
          ) : (
            <p className="text-gray-500 text-center py-4">No details available</p>
          )}
        </div>
      )}
    </div>
  );
};

export default function DeploymentHistory({ projectId }) {
  const [expandedId, setExpandedId] = useState(null);
  const [page, setPage] = useState(0);
  const limit = 10;

  const { data, isLoading, error } = useQuery({
    queryKey: ['deployments', projectId, page],
    queryFn: () => getProjectDeployments(projectId, limit, page * limit),
    enabled: !!projectId,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <svg className="animate-spin h-6 w-6 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8 text-red-500">
        Failed to load deployment history: {error.message}
      </div>
    );
  }

  const deployments = data?.data || [];
  const total = data?.meta?.total || 0;
  const totalPages = Math.ceil(total / limit);

  if (deployments.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        No deployments yet. Click "Deploy" to create your first deployment.
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
        <h3 className="font-semibold text-gray-900">Deployment History</h3>
        <p className="text-sm text-gray-500">{total} total deployments</p>
      </div>

      <div className="divide-y divide-gray-200">
        {deployments.map((deployment) => (
          <DeploymentRow
            key={deployment.id}
            deployment={deployment}
            isExpanded={expandedId === deployment.id}
            onToggle={() => setExpandedId(expandedId === deployment.id ? null : deployment.id)}
          />
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between px-4 py-3 bg-gray-50 border-t border-gray-200">
          <button
            onClick={() => setPage(p => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1 text-sm bg-white border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="text-sm text-gray-500">
            Page {page + 1} of {totalPages}
          </span>
          <button
            onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1 text-sm bg-white border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}
