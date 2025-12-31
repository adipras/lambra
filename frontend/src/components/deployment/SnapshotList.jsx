import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { History, RotateCcw, Trash2, ChevronDown, ChevronUp, Check, Clock, AlertCircle, Loader2 } from 'lucide-react'
import { snapshotsApi } from '../../api/snapshots'

export const SnapshotList = ({ projectId, onNotification }) => {
  const queryClient = useQueryClient()
  const [expandedSnapshot, setExpandedSnapshot] = useState(null)
  const [confirmRollback, setConfirmRollback] = useState(null)

  // Fetch snapshots
  const { data: snapshotsData, isLoading } = useQuery({
    queryKey: ['snapshots', projectId],
    queryFn: () => snapshotsApi.getByProject(projectId),
    enabled: !!projectId,
  })

  // Rollback mutation
  const rollbackMutation = useMutation({
    mutationFn: (snapshotId) => snapshotsApi.rollback(snapshotId),
    onSuccess: () => {
      queryClient.invalidateQueries(['snapshots', projectId])
      queryClient.invalidateQueries(['entities', projectId])
      queryClient.invalidateQueries(['project', projectId])
      setConfirmRollback(null)
      onNotification?.('success', 'Rollback completed successfully! Service is being redeployed.')
    },
    onError: (error) => {
      onNotification?.('error', error.response?.data?.message || 'Failed to rollback')
      setConfirmRollback(null)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (snapshotId) => snapshotsApi.delete(snapshotId),
    onSuccess: () => {
      queryClient.invalidateQueries(['snapshots', projectId])
      onNotification?.('success', 'Snapshot deleted')
    },
  })

  // Create snapshot mutation
  const createMutation = useMutation({
    mutationFn: () => snapshotsApi.create(projectId),
    onSuccess: () => {
      queryClient.invalidateQueries(['snapshots', projectId])
      onNotification?.('success', 'Snapshot created successfully!')
    },
    onError: (error) => {
      onNotification?.('error', error.response?.data?.message || 'Failed to create snapshot')
    },
  })

  const snapshots = snapshotsData?.data || []

  const getStatusIcon = (status) => {
    switch (status) {
      case 'active':
        return <Check className="w-4 h-4 text-green-500" />
      case 'rolled_back':
        return <RotateCcw className="w-4 h-4 text-yellow-500" />
      default:
        return <Clock className="w-4 h-4 text-gray-500" />
    }
  }

  const getStatusBadge = (status) => {
    const styles = {
      active: 'bg-green-100 text-green-800',
      rolled_back: 'bg-yellow-100 text-yellow-800',
      created: 'bg-gray-100 text-gray-800',
    }
    return (
      <span className={`px-2 py-1 rounded-full text-xs font-medium ${styles[status] || styles.created}`}>
        {status === 'rolled_back' ? 'Rolled Back' : status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    )
  }

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
      </div>
    )
  }

  return (
    <div className="bg-white rounded-lg shadow">
      <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <History className="w-5 h-5 text-gray-500" />
          <h3 className="text-lg font-medium text-gray-900">Snapshots</h3>
          <span className="text-sm text-gray-500">({snapshots.length})</span>
        </div>
        <button
          onClick={() => createMutation.mutate()}
          disabled={createMutation.isPending}
          className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center gap-1"
        >
          {createMutation.isPending ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <History className="w-4 h-4" />
          )}
          Create Snapshot
        </button>
      </div>

      {snapshots.length === 0 ? (
        <div className="px-4 py-8 text-center text-gray-500">
          <History className="w-12 h-12 mx-auto mb-3 text-gray-300" />
          <p>No snapshots yet</p>
          <p className="text-sm mt-1">Snapshots are created automatically when you deploy</p>
        </div>
      ) : (
        <div className="divide-y divide-gray-200">
          {snapshots.map((snapshot) => (
            <div key={snapshot.id} className="px-4 py-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {getStatusIcon(snapshot.status)}
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-gray-900">{snapshot.version}</span>
                      {getStatusBadge(snapshot.status)}
                    </div>
                    <div className="text-sm text-gray-500">
                      {formatDate(snapshot.created_at)}
                      {snapshot.created_by?.String && (
                        <span className="ml-2">by {snapshot.created_by.String}</span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {snapshot.status !== 'active' && (
                    <>
                      {confirmRollback === snapshot.id ? (
                        <div className="flex items-center gap-2">
                          <span className="text-sm text-orange-600">Confirm rollback?</span>
                          <button
                            onClick={() => rollbackMutation.mutate(snapshot.id)}
                            disabled={rollbackMutation.isPending}
                            className="px-2 py-1 text-sm bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
                          >
                            {rollbackMutation.isPending ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              'Yes'
                            )}
                          </button>
                          <button
                            onClick={() => setConfirmRollback(null)}
                            className="px-2 py-1 text-sm bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
                          >
                            No
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => setConfirmRollback(snapshot.id)}
                          className="p-1.5 text-orange-600 hover:bg-orange-50 rounded"
                          title="Rollback to this snapshot"
                        >
                          <RotateCcw className="w-4 h-4" />
                        </button>
                      )}
                    </>
                  )}

                  <button
                    onClick={() => setExpandedSnapshot(expandedSnapshot === snapshot.id ? null : snapshot.id)}
                    className="p-1.5 text-gray-500 hover:bg-gray-100 rounded"
                  >
                    {expandedSnapshot === snapshot.id ? (
                      <ChevronUp className="w-4 h-4" />
                    ) : (
                      <ChevronDown className="w-4 h-4" />
                    )}
                  </button>

                  {snapshot.status !== 'active' && (
                    <button
                      onClick={() => deleteMutation.mutate(snapshot.id)}
                      disabled={deleteMutation.isPending}
                      className="p-1.5 text-red-500 hover:bg-red-50 rounded"
                      title="Delete snapshot"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </div>
              </div>

              {expandedSnapshot === snapshot.id && (
                <div className="mt-3 p-3 bg-gray-50 rounded-lg text-sm">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <span className="text-gray-500">Commit Hash:</span>
                      <span className="ml-2 font-mono text-gray-700">{snapshot.git_commit_hash}</span>
                    </div>
                    {snapshot.git_tag?.String && (
                      <div>
                        <span className="text-gray-500">Git Tag:</span>
                        <span className="ml-2 font-mono text-gray-700">{snapshot.git_tag.String}</span>
                      </div>
                    )}
                  </div>
                  {snapshot.metadata && (
                    <div className="mt-3">
                      <SnapshotMetadataSummary metadata={snapshot.metadata} />
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// Helper component to show metadata summary
const SnapshotMetadataSummary = ({ metadata }) => {
  try {
    const data = typeof metadata === 'string' ? JSON.parse(metadata) : metadata
    const entityCount = data.entities?.length || 0
    const endpointCount = data.endpoints?.length || 0

    return (
      <div className="flex items-center gap-4 text-gray-600">
        <span>{entityCount} {entityCount === 1 ? 'entity' : 'entities'}</span>
        <span>{endpointCount} {endpointCount === 1 ? 'endpoint' : 'endpoints'}</span>
      </div>
    )
  } catch {
    return null
  }
}
