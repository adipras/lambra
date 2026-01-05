import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ArrowLeft, Plus, Trash2, Code, Play, Eye, X, FileCode, Check, AlertCircle,
  Square, Rocket, ExternalLink, RefreshCw, ChevronRight, FileJson,
  ChevronDown, ChevronUp, Database, Zap, Settings, MoreVertical, Copy, Terminal, FileText
} from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { projectsApi } from '../api/projects'
import { entitiesApi } from '../api/entities'
import { endpointsApi } from '../api/endpoints'
import { generatorApi } from '../api/generator'
import { deploymentApi } from '../api/deployment'
import { exportApi } from '../api/export'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'
import { ErrorAlert } from '../components/shared/ErrorAlert'
import { StatusBadge } from '../components/shared/StatusBadge'
import { StatusIndicator } from '../components/shared/StatusIndicator'
import { EntityCardSkeleton, StatsCardSkeleton } from '../components/shared/Skeleton'
import { EntityForm } from '../components/forms/EntityForm'
import { EndpointForm } from '../components/forms/EndpointForm'
import { CodeEditor } from '../components/code/CodeEditor'
import { ExampleBody } from '../components/shared/ExampleBody'
import { TestEndpointModal } from '../components/shared/TestEndpointModal'
import { DeleteServiceModal } from '../components/shared/DeleteServiceModal'
import { DeleteEntityModal } from '../components/shared/DeleteEntityModal'
import { DeployProgressModal } from '../components/deployment/DeployProgressModal'
import { SnapshotList } from '../components/deployment/SnapshotList'
import LogsModal from '../components/logs/LogsModal'
import DeploymentHistory from '../components/logs/DeploymentHistory'

export const ServiceDetail = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [showEntityModal, setShowEntityModal] = useState(false)
  const [showEndpointModal, setShowEndpointModal] = useState(false)
  const [showPreviewModal, setShowPreviewModal] = useState(false)
  const [selectedEntity, setSelectedEntity] = useState(null)
  const [previewEntity, setPreviewEntity] = useState(null)
  const [notification, setNotification] = useState(null)

  // Fetch project
  const { data: projectData, isLoading: projectLoading, error: projectError } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectsApi.getById(id),
  })

  // Fetch entities
  const { data: entitiesData, isLoading: entitiesLoading } = useQuery({
    queryKey: ['entities', id],
    queryFn: () => entitiesApi.getByProject(id),
  })

  // Create entity mutation with optimistic update
  const createEntityMutation = useMutation({
    mutationFn: (data) => entitiesApi.create(id, data),
    onMutate: async (newEntity) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries(['entities', id])

      // Snapshot the previous value
      const previousEntities = queryClient.getQueryData(['entities', id])

      // Optimistically update to the new value
      queryClient.setQueryData(['entities', id], (old) => {
        const optimisticEntity = {
          id: 'temp-' + Date.now(),
          ...newEntity,
          endpoints_count: newEntity.generate_endpoints ? 5 : 0,
          _isOptimistic: true,
        }
        return {
          ...old,
          data: [...(old?.data || []), optimisticEntity],
        }
      })

      // Return context with the snapshotted value
      return { previousEntities }
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['entities', id])
      setShowEntityModal(false)
      showNotification('success', 'Entity created successfully!')
    },
    onError: (error, newEntity, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      queryClient.setQueryData(['entities', id], context.previousEntities)
      showNotification('error', error.response?.data?.message || 'Failed to create entity')
    },
  })

  // Create endpoint mutation
  const createEndpointMutation = useMutation({
    mutationFn: (data) => endpointsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries(['entity-endpoints'])
      setShowEndpointModal(false)
      setSelectedEntity(null)
      showNotification('success', 'Endpoint created successfully!')
    },
  })

  // Delete entity mutation with optimistic update
  const deleteEntityMutation = useMutation({
    mutationFn: (entityId) => entitiesApi.delete(entityId),
    onMutate: async (entityId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries(['entities', id])

      // Snapshot the previous value
      const previousEntities = queryClient.getQueryData(['entities', id])

      // Optimistically remove from the list
      queryClient.setQueryData(['entities', id], (old) => ({
        ...old,
        data: (old?.data || []).filter((e) => e.id !== entityId),
      }))

      return { previousEntities }
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['entities', id])
      showNotification('success', 'Entity deleted successfully!')
    },
    onError: (error, entityId, context) => {
      // Roll back on error
      queryClient.setQueryData(['entities', id], context.previousEntities)
      showNotification('error', error.response?.data?.message || 'Failed to delete entity')
    },
  })

  // Delete endpoint mutation
  const deleteEndpointMutation = useMutation({
    mutationFn: (endpointId) => endpointsApi.delete(endpointId),
    onSuccess: () => {
      queryClient.invalidateQueries(['entity-endpoints'])
    },
  })

  // Preview code query
  const { data: previewData, isLoading: previewLoading, refetch: refetchPreview } = useQuery({
    queryKey: ['preview', previewEntity?.id],
    queryFn: () => generatorApi.previewEntity(previewEntity.id),
    enabled: !!previewEntity,
  })

  // Generate project mutation
  const generateProjectMutation = useMutation({
    mutationFn: () => generatorApi.generateProject(id),
    onSuccess: (response) => {
      showNotification('success', `Generated ${response.data?.files?.length || 0} files successfully!`)
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.error || 'Failed to generate code')
    },
  })

  // Generate entity mutation
  const generateEntityMutation = useMutation({
    mutationFn: (entityId) => generatorApi.generateEntity(entityId),
    onSuccess: (response) => {
      showNotification('success', `Generated ${response.data?.files?.length || 0} files for entity!`)
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.error || 'Failed to generate code')
    },
  })

  const showNotification = (type, message) => {
    setNotification({ type, message })
    setTimeout(() => setNotification(null), 5000)
  }

  const handlePreview = (entity) => {
    setPreviewEntity(entity)
    setShowPreviewModal(true)
  }

  // Fetch deployment status
  const { data: statusData, refetch: refetchStatus } = useQuery({
    queryKey: ['deployment-status', id],
    queryFn: () => deploymentApi.getStatus(id),
    refetchInterval: (data) => {
      return data?.data?.status === 'running' ? 5000 : false
    },
  })

  const deploymentStatus = statusData?.data

  // Deploy mutation with progress modal
  const deployMutation = useMutation({
    mutationFn: (options = {}) => deploymentApi.deploy(id, options),
    onMutate: () => {
      // Show progress modal immediately
      setShowDeployProgress(true)
      setShowDeployOptions(false)
    },
    onSuccess: (response) => {
      const deploymentId = response.data?.deployment_id
      if (deploymentId) {
        setCurrentDeploymentId(deploymentId)
      }
      showNotification('success', `Service deployed! URL: ${response.data?.url || 'N/A'}`)
      refetchStatus()
      setResetDatabase(false) // Reset the checkbox
    },
    onError: (error) => {
      setShowDeployProgress(false)
      showNotification('error', error.response?.data?.message || 'Failed to deploy service')
    },
  })

  // Handle deploy with options
  const handleDeploy = () => {
    deployMutation.mutate({ reset_database: resetDatabase })
  }

  // Handle deploy progress completion
  const handleDeployComplete = useCallback(() => {
    refetchStatus()
    queryClient.invalidateQueries(['deployments', id])
    setTimeout(() => {
      setShowDeployProgress(false)
      setCurrentDeploymentId(null)
    }, 1500)
  }, [refetchStatus, queryClient, id])

  // Start mutation
  const startMutation = useMutation({
    mutationFn: () => deploymentApi.start(id),
    onSuccess: () => {
      showNotification('success', 'Service started successfully')
      refetchStatus()
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.message || 'Failed to start service')
    },
  })

  // Stop mutation
  const stopMutation = useMutation({
    mutationFn: () => deploymentApi.stop(id),
    onSuccess: () => {
      showNotification('success', 'Service stopped successfully')
      refetchStatus()
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.message || 'Failed to stop service')
    },
  })

  // Redeploy mutation
  const redeployMutation = useMutation({
    mutationFn: () => deploymentApi.redeploy(id),
    onSuccess: (response) => {
      showNotification('success', `Service redeployed! URL: ${response.data?.url || 'N/A'}`)
      refetchStatus()
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.message || 'Failed to redeploy service')
    },
  })

  // Destroy mutation
  const destroyMutation = useMutation({
    mutationFn: () => deploymentApi.destroy(id),
    onSuccess: () => {
      // Invalidate projects query cache
      queryClient.invalidateQueries(['projects'])
      // Navigate to services list since project is deleted
      navigate('/services', { replace: true })
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.message || 'Failed to destroy service')
    },
  })

  const isDeploying = deployMutation.isPending
  const isStarting = startMutation.isPending
  const isStopping = stopMutation.isPending
  const isRedeploying = redeployMutation.isPending

  const [isExporting, setIsExporting] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showDeployOptions, setShowDeployOptions] = useState(false)
  const [resetDatabase, setResetDatabase] = useState(false)
  const [showActionsMenu, setShowActionsMenu] = useState(false)
  const [showLogsModal, setShowLogsModal] = useState(false)
  const [showDeployProgress, setShowDeployProgress] = useState(false)
  const [currentDeploymentId, setCurrentDeploymentId] = useState(null)
  const [entityToDelete, setEntityToDelete] = useState(null)

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (showDeployOptions && !e.target.closest('.deploy-options-container')) {
        setShowDeployOptions(false)
      }
    }
    document.addEventListener('click', handleClickOutside)
    return () => document.removeEventListener('click', handleClickOutside)
  }, [showDeployOptions])

  const handleExportOpenAPI = async () => {
    setIsExporting(true)
    try {
      await exportApi.downloadOpenAPI(id)
      showNotification('success', 'OpenAPI specification downloaded!')
    } catch (error) {
      showNotification('error', 'Failed to export OpenAPI specification')
    } finally {
      setIsExporting(false)
    }
  }

  const handleExportPostman = async () => {
    setIsExporting(true)
    try {
      await exportApi.downloadPostman(id)
      showNotification('success', 'Postman collection downloaded!')
    } catch (error) {
      showNotification('error', 'Failed to export Postman collection')
    } finally {
      setIsExporting(false)
    }
  }

  if (projectLoading) {
    return <LoadingSpinner size="lg" className="mt-20" />
  }

  if (projectError) {
    return <ErrorAlert message="Failed to load project details" />
  }

  const project = projectData?.data
  const entities = entitiesData?.data || []

  const handleDeleteEntity = (entity) => {
    setEntityToDelete(entity)
  }

  const confirmDeleteEntity = () => {
    if (entityToDelete) {
      deleteEntityMutation.mutate(entityToDelete.id, {
        onSuccess: () => {
          setEntityToDelete(null)
        },
        onError: () => {
          setEntityToDelete(null)
        }
      })
    }
  }

  const handleDeleteEndpoint = (endpointId) => {
    if (window.confirm('Are you sure you want to delete this endpoint?')) {
      deleteEndpointMutation.mutate(endpointId)
    }
  }

  return (
    <div className="space-y-6">
      {/* Notification */}
      {notification && (
        <div className={`fixed top-4 right-4 z-50 flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg animate-slide-in ${
          notification.type === 'success'
            ? 'bg-green-50 text-green-800 border border-green-200'
            : 'bg-red-50 text-red-800 border border-red-200'
        }`}>
          {notification.type === 'success' ? (
            <Check className="w-5 h-5 text-green-600" />
          ) : (
            <AlertCircle className="w-5 h-5 text-red-600" />
          )}
          <span className="font-medium">{notification.message}</span>
          <button onClick={() => setNotification(null)} className="ml-2 hover:opacity-70">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <button
            onClick={() => navigate('/services')}
            className="btn btn-secondary p-2 mt-1"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold text-gray-900">{project?.name}</h1>
              <StatusBadge status={project?.status} />
            </div>
            <p className="text-gray-500 mt-1">{project?.description || 'No description'}</p>
          </div>
        </div>

        {/* Action Buttons - Clean Layout */}
        <div className="flex items-center gap-2">
          {/* Status Indicator - Enhanced */}
          <StatusIndicator
            status={isDeploying ? 'deploying' : (deploymentStatus?.status || 'not_deployed')}
            size="md"
            pulse={true}
          />

          {/* Divider */}
          <div className="h-8 w-px bg-gray-200" />

          {/* Button Group */}
          <div className="flex items-center bg-gray-100 rounded-lg p-1 gap-1">
            {/* Primary Action */}
            {deploymentStatus?.status === 'running' ? (
              <button
                onClick={() => stopMutation.mutate()}
                disabled={isStopping || isRedeploying}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-red-500 hover:bg-red-600 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
              >
                {isStopping ? <LoadingSpinner size="sm" /> : <Square className="w-3.5 h-3.5" />}
                <span>Stop</span>
              </button>
            ) : deploymentStatus?.status === 'stopped' ? (
              <button
                onClick={() => startMutation.mutate()}
                disabled={isStarting || isRedeploying}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-green-500 hover:bg-green-600 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
              >
                {isStarting ? <LoadingSpinner size="sm" /> : <Play className="w-3.5 h-3.5" />}
                <span>Start</span>
              </button>
            ) : (
              <div className="relative deploy-options-container">
                <div className="flex">
                  <button
                    onClick={handleDeploy}
                    disabled={isDeploying || entities.length === 0}
                    className="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-l-md text-sm font-medium transition-colors disabled:opacity-50"
                    title={entities.length === 0 ? 'Add at least one entity first' : 'Deploy to Docker'}
                  >
                    {isDeploying ? <LoadingSpinner size="sm" /> : <Rocket className="w-3.5 h-3.5" />}
                    <span>Deploy</span>
                  </button>
                  <button
                    onClick={() => setShowDeployOptions(!showDeployOptions)}
                    disabled={isDeploying || entities.length === 0}
                    className="flex items-center px-2 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-r-md text-sm font-medium transition-colors disabled:opacity-50 border-l border-indigo-500"
                    title="Deploy options"
                  >
                    <ChevronDown className="w-3.5 h-3.5" />
                  </button>
                </div>
                {/* Deploy Options Dropdown */}
                {showDeployOptions && (
                  <div className="absolute top-full left-0 mt-1 w-56 bg-white rounded-lg shadow-lg border border-gray-200 py-2 z-50">
                    <div className="px-3 py-2 border-b border-gray-100">
                      <label className="flex items-center gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={resetDatabase}
                          onChange={(e) => setResetDatabase(e.target.checked)}
                          className="w-4 h-4 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500"
                        />
                        <div>
                          <span className="text-sm font-medium text-gray-700">Reset Database</span>
                          <p className="text-xs text-gray-500">Drop all tables before deploy</p>
                        </div>
                      </label>
                    </div>
                    <div className="px-3 py-2">
                      <button
                        onClick={handleDeploy}
                        className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-md text-sm font-medium transition-colors"
                      >
                        <Rocket className="w-4 h-4" />
                        {resetDatabase ? 'Deploy with Reset' : 'Deploy'}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Open URL - only when running */}
            {deploymentStatus?.status === 'running' && deploymentStatus?.url && (
              <a
                href={deploymentStatus.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 px-3 py-1.5 bg-white hover:bg-gray-50 text-gray-700 rounded-md text-sm font-medium transition-colors border border-gray-200"
                title="Open service URL"
              >
                <ExternalLink className="w-3.5 h-3.5" />
                <span>Open</span>
              </a>
            )}

            {/* Redeploy - only when deployed */}
            {(deploymentStatus?.status === 'running' || deploymentStatus?.status === 'stopped') && (
              <button
                onClick={() => redeployMutation.mutate()}
                disabled={isRedeploying}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-white hover:bg-gray-50 text-gray-700 rounded-md text-sm font-medium transition-colors border border-gray-200 disabled:opacity-50"
                title="Regenerate code and rebuild"
              >
                {isRedeploying ? <LoadingSpinner size="sm" /> : <RefreshCw className="w-3.5 h-3.5" />}
                <span>Redeploy</span>
              </button>
            )}
          </div>

          {/* More Actions */}
          <div className="relative">
            <button
              onClick={() => setShowActionsMenu(!showActionsMenu)}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              title="More actions"
            >
              <MoreVertical className="w-5 h-5 text-gray-500" />
            </button>
            {showActionsMenu && (
              <>
                <div
                  className="fixed inset-0 z-10"
                  onClick={() => setShowActionsMenu(false)}
                />
                <div className="absolute right-0 mt-2 w-52 bg-white rounded-lg shadow-lg border border-gray-200 py-1.5 z-20">
                  {/* Refresh Status */}
                  <button
                    onClick={() => { refetchStatus(); setShowActionsMenu(false) }}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2.5"
                  >
                    <RefreshCw className="w-4 h-4 text-gray-400" />
                    <span>Refresh Status</span>
                  </button>

                  {/* View Logs */}
                  <button
                    onClick={() => { setShowLogsModal(true); setShowActionsMenu(false) }}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2.5"
                  >
                    <FileText className="w-4 h-4 text-gray-400" />
                    <span>View Logs</span>
                  </button>

                  <div className="border-t border-gray-100 my-1.5" />

                  {/* Export Options */}
                  <div className="px-3 py-1.5">
                    <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Export</span>
                  </div>
                  <button
                    onClick={() => { handleExportOpenAPI(); setShowActionsMenu(false) }}
                    disabled={isExporting || entities.length === 0}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2.5 disabled:opacity-50"
                  >
                    <FileJson className="w-4 h-4 text-blue-500" />
                    <span>OpenAPI Spec</span>
                  </button>
                  <button
                    onClick={() => { handleExportPostman(); setShowActionsMenu(false) }}
                    disabled={isExporting || entities.length === 0}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2.5 disabled:opacity-50"
                  >
                    <FileJson className="w-4 h-4 text-orange-500" />
                    <span>Postman Collection</span>
                  </button>

                  {/* Delete - always available */}
                  <div className="border-t border-gray-100 my-1.5" />
                  <button
                    onClick={() => { setShowDeleteModal(true); setShowActionsMenu(false) }}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-red-50 flex items-center gap-2.5 text-red-600"
                  >
                    <Trash2 className="w-4 h-4" />
                    <span>Delete Project</span>
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4">
        {entitiesLoading ? (
          <>
            <StatsCardSkeleton />
            <StatsCardSkeleton />
            <StatsCardSkeleton />
            <StatsCardSkeleton />
          </>
        ) : (
          <>
            <div className="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition-shadow">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-indigo-100 rounded-lg">
                  <Database className="w-5 h-5 text-indigo-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-900">{entities.length}</p>
                  <p className="text-sm text-gray-500">Entities</p>
                </div>
              </div>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition-shadow">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-green-100 rounded-lg">
                  <Zap className="w-5 h-5 text-green-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-900">
                    {entities.reduce((acc, e) => acc + (e.endpoints_count || 0), 0) || '-'}
                  </p>
                  <p className="text-sm text-gray-500">Endpoints</p>
                </div>
              </div>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition-shadow">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-purple-100 rounded-lg">
                  <Terminal className="w-5 h-5 text-purple-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-900 font-mono text-lg">
                    {deploymentStatus?.port || '-'}
                  </p>
                  <p className="text-sm text-gray-500">Port</p>
                </div>
              </div>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition-shadow">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-orange-100 rounded-lg">
                  <Settings className="w-5 h-5 text-orange-600" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900 truncate max-w-[150px]" title={project?.namespace}>
                    {project?.namespace}
                  </p>
                  <p className="text-sm text-gray-500">Namespace</p>
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Entities Section */}
      <div className="bg-white rounded-xl border border-gray-200">
        <div className="flex items-center justify-between p-5 border-b border-gray-100">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-indigo-100 rounded-lg">
              <Database className="w-5 h-5 text-indigo-600" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Entities & Endpoints</h2>
              <p className="text-sm text-gray-500">Define your data models and API endpoints</p>
            </div>
          </div>
          <button
            onClick={() => setShowEntityModal(true)}
            className="btn btn-primary"
          >
            <Plus className="w-4 h-4 mr-1.5" />
            Add Entity
          </button>
        </div>

        <div className="p-5">
          {entitiesLoading ? (
            <div className="space-y-4">
              <EntityCardSkeleton />
              <EntityCardSkeleton />
            </div>
          ) : entities.length === 0 ? (
            <div className="text-center py-16">
              <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Database className="w-8 h-8 text-gray-400" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-1">No entities yet</h3>
              <p className="text-gray-500 mb-6">Create your first entity to define your data model</p>
              <button
                onClick={() => setShowEntityModal(true)}
                className="btn btn-primary"
              >
                <Plus className="w-4 h-4 mr-1.5" />
                Create First Entity
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {entities.map((entity) => (
                <EntityCard
                  key={entity.id}
                  entity={entity}
                  serviceUrl={deploymentStatus?.url}
                  onAddEndpoint={() => {
                    setSelectedEntity(entity)
                    setShowEndpointModal(true)
                  }}
                  onDelete={() => handleDeleteEntity(entity)}
                  onDeleteEndpoint={handleDeleteEndpoint}
                  onPreview={() => handlePreview(entity)}
                />
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Snapshots Section */}
      <div className="mt-8">
        <SnapshotList projectId={id} onNotification={showNotification} />
      </div>

      {/* Deployment History Section */}
      <div className="mt-8">
        <DeploymentHistory projectId={id} />
      </div>

      {/* Entity Modal */}
      {showEntityModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
              <div>
                <h2 className="text-xl font-bold text-gray-900">Create New Entity</h2>
                <p className="text-sm text-gray-500">Define your data model with fields and auto-generated endpoints</p>
              </div>
              <button onClick={() => setShowEntityModal(false)} className="p-2 hover:bg-gray-100 rounded-lg">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6">
              <EntityForm
                onSubmit={(data) => createEntityMutation.mutate(data)}
                onCancel={() => setShowEntityModal(false)}
                isLoading={createEntityMutation.isPending}
                projectId={id}
              />
            </div>
          </div>
        </div>
      )}

      {/* Endpoint Modal */}
      {showEndpointModal && selectedEntity && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-2xl max-w-3xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
              <div>
                <h2 className="text-xl font-bold text-gray-900">Create Endpoint</h2>
                <p className="text-sm text-gray-500">for entity: <span className="font-medium text-indigo-600">{selectedEntity.name}</span></p>
              </div>
              <button onClick={() => { setShowEndpointModal(false); setSelectedEntity(null) }} className="p-2 hover:bg-gray-100 rounded-lg">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6">
              <EndpointForm
                entityId={selectedEntity.id}
                onSubmit={(data) => createEndpointMutation.mutate(data)}
                onCancel={() => {
                  setShowEndpointModal(false)
                  setSelectedEntity(null)
                }}
                isLoading={createEndpointMutation.isPending}
              />
            </div>
          </div>
        </div>
      )}

      {/* Code Preview Modal */}
      {showPreviewModal && previewEntity && (
        <CodePreviewModal
          entity={previewEntity}
          files={previewData?.data?.files || []}
          isLoading={previewLoading}
          onClose={() => {
            setShowPreviewModal(false)
            setPreviewEntity(null)
          }}
          onGenerate={() => {
            generateEntityMutation.mutate(previewEntity.id)
            setShowPreviewModal(false)
            setPreviewEntity(null)
          }}
          isGenerating={generateEntityMutation.isPending}
        />
      )}

      {/* Delete Service Modal */}
      {showDeleteModal && (
        <DeleteServiceModal
          serviceName={project?.name || 'Unknown Service'}
          onConfirm={() => destroyMutation.mutate()}
          onCancel={() => setShowDeleteModal(false)}
          isDeleting={destroyMutation.isPending}
        />
      )}

      {/* Delete Entity Modal */}
      {entityToDelete && (
        <DeleteEntityModal
          entityName={entityToDelete.name}
          tableName={entityToDelete.table_name}
          endpointsCount={entityToDelete.endpoints_count || 0}
          isDeployed={deploymentStatus?.status === 'running' || deploymentStatus?.status === 'stopped'}
          onConfirm={confirmDeleteEntity}
          onCancel={() => setEntityToDelete(null)}
          isDeleting={deleteEntityMutation.isPending}
        />
      )}

      {/* Logs Modal */}
      <LogsModal
        projectId={id}
        isOpen={showLogsModal}
        onClose={() => setShowLogsModal(false)}
      />

      {/* Deploy Progress Modal */}
      <DeployProgressModal
        isOpen={showDeployProgress}
        onClose={() => {
          setShowDeployProgress(false)
          setCurrentDeploymentId(null)
        }}
        projectId={id}
        deploymentId={currentDeploymentId}
        onComplete={handleDeployComplete}
      />
    </div>
  )
}

// EntityCard component with improved UI
const EntityCard = ({ entity, onAddEndpoint, onDelete, onDeleteEndpoint, onPreview, serviceUrl }) => {
  const [isExpanded, setIsExpanded] = useState(true)
  const [selectedEndpoint, setSelectedEndpoint] = useState(null)
  const [testEndpoint, setTestEndpoint] = useState(null)

  const isOptimistic = entity._isOptimistic

  const { data: endpointsData, isLoading: endpointsLoading } = useQuery({
    queryKey: ['entity-endpoints', entity.id],
    queryFn: () => endpointsApi.getByEntity(entity.id),
    enabled: !isOptimistic, // Don't fetch endpoints for optimistic entities
  })

  const endpoints = endpointsData?.data || []

  const getMethodBadge = (method) => {
    const styles = {
      GET: 'bg-blue-500 text-white',
      POST: 'bg-green-500 text-white',
      PUT: 'bg-yellow-500 text-white',
      DELETE: 'bg-red-500 text-white',
      PATCH: 'bg-purple-500 text-white',
    }
    return styles[method] || 'bg-gray-500 text-white'
  }

  const copyEndpointUrl = (endpoint) => {
    const baseUrl = serviceUrl || 'http://localhost:9850'
    navigator.clipboard.writeText(`${baseUrl}${endpoint.path}`)
  }

  return (
    <>
      <div className={`border rounded-xl overflow-hidden transition-all duration-300 ${
        isOptimistic
          ? 'border-indigo-300 bg-indigo-50/50 animate-pulse'
          : 'border-gray-200 hover:border-gray-300'
      }`}>
        {/* Entity Header */}
        <div
          className={`flex items-center justify-between p-4 cursor-pointer ${isOptimistic ? 'bg-indigo-50' : 'bg-gray-50'}`}
          onClick={() => !isOptimistic && setIsExpanded(!isExpanded)}
        >
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-lg border ${isOptimistic ? 'bg-indigo-100 border-indigo-200' : 'bg-white border-gray-200'}`}>
              {isOptimistic ? (
                <LoadingSpinner size="sm" />
              ) : (
                <Database className="w-4 h-4 text-indigo-600" />
              )}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="font-semibold text-gray-900">{entity.name}</h3>
                {isOptimistic ? (
                  <span className="text-xs px-2 py-0.5 bg-indigo-200 text-indigo-700 rounded-full animate-pulse">
                    Saving...
                  </span>
                ) : (
                  <span className="text-xs px-2 py-0.5 bg-gray-200 text-gray-600 rounded-full">
                    {endpoints.length} endpoints
                  </span>
                )}
              </div>
              <p className="text-sm text-gray-500">
                Table: <span className="font-mono text-xs">{entity.table_name}</span>
              </p>
            </div>
          </div>
          {!isOptimistic && (
            <div className="flex items-center gap-2">
              <button
                onClick={(e) => { e.stopPropagation(); onPreview() }}
                className="p-2 text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                title="Preview Code"
              >
                <Eye className="w-4 h-4" />
              </button>
              <button
                onClick={(e) => { e.stopPropagation(); onAddEndpoint() }}
                className="p-2 text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                title="Add Endpoint"
              >
                <Plus className="w-4 h-4" />
              </button>
              <button
                onClick={(e) => { e.stopPropagation(); onDelete() }}
                className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                title="Delete Entity"
              >
                <Trash2 className="w-4 h-4" />
              </button>
              {isExpanded ? <ChevronUp className="w-4 h-4 text-gray-400" /> : <ChevronDown className="w-4 h-4 text-gray-400" />}
            </div>
          )}
        </div>

        {/* Expanded Content */}
        {isExpanded && !isOptimistic && (
          <div className="p-4 space-y-4">
            {/* Fields */}
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">Fields</p>
              <div className="flex flex-wrap gap-2">
                {entity.fields?.map((field, idx) => (
                  <div
                    key={idx}
                    className="inline-flex items-center gap-1.5 px-2.5 py-1.5 bg-gray-100 rounded-lg"
                  >
                    <span className="font-mono text-sm text-gray-700">{field.name}</span>
                    <span className="text-xs text-gray-400">({field.type})</span>
                    {field.required && <span className="w-1 h-1 bg-red-500 rounded-full" title="Required" />}
                  </div>
                ))}
              </div>
            </div>

            {/* Endpoints */}
            {endpoints.length > 0 && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">Endpoints</p>
                <div className="space-y-2">
                  {endpoints.map((endpoint) => (
                    <div key={endpoint.id}>
                      <div
                        className={`flex items-center justify-between p-3 rounded-lg border transition-all cursor-pointer ${
                          selectedEndpoint?.id === endpoint.id
                            ? 'border-indigo-300 bg-indigo-50'
                            : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                        }`}
                        onClick={() => setSelectedEndpoint(selectedEndpoint?.id === endpoint.id ? null : endpoint)}
                      >
                        <div className="flex items-center gap-3">
                          <span className={`px-2.5 py-1 text-xs font-bold rounded ${getMethodBadge(endpoint.method)}`}>
                            {endpoint.method}
                          </span>
                          <div>
                            <p className="font-medium text-gray-900">{endpoint.name}</p>
                            <p className="text-xs font-mono text-gray-500">{endpoint.path}</p>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          {/* Test Button */}
                          <button
                            onClick={(e) => { e.stopPropagation(); setTestEndpoint(endpoint) }}
                            className="flex items-center gap-1 px-2.5 py-1.5 text-xs font-medium text-green-700 bg-green-100 hover:bg-green-200 rounded-lg transition-colors"
                            title="Test Endpoint"
                          >
                            <Play className="w-3.5 h-3.5" />
                            Test
                          </button>
                          <button
                            onClick={(e) => { e.stopPropagation(); copyEndpointUrl(endpoint) }}
                            className="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded"
                            title="Copy URL"
                          >
                            <Copy className="w-3.5 h-3.5" />
                          </button>
                          <button
                            onClick={(e) => { e.stopPropagation(); onDeleteEndpoint(endpoint.id) }}
                            className="p-1.5 text-red-400 hover:text-red-600 hover:bg-red-50 rounded"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                          {selectedEndpoint?.id === endpoint.id ? (
                            <ChevronUp className="w-4 h-4 text-gray-400" />
                          ) : (
                            <ChevronDown className="w-4 h-4 text-gray-400" />
                          )}
                        </div>
                      </div>

                      {/* Example Body Preview */}
                      {selectedEndpoint?.id === endpoint.id && (
                        <div className="mt-2 grid grid-cols-2 gap-3 p-3 bg-gray-50 rounded-lg border border-gray-200">
                          <ExampleBody
                            schema={endpoint.request_schema}
                            title="Example Request"
                            type="request"
                          />
                          <ExampleBody
                            schema={endpoint.response_schema}
                            title="Example Response"
                            type="response"
                          />
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Test Endpoint Modal */}
      {testEndpoint && (
        <TestEndpointModal
          endpoint={testEndpoint}
          serviceUrl={serviceUrl}
          onClose={() => setTestEndpoint(null)}
        />
      )}
    </>
  )
}

// CodePreviewModal component
const CodePreviewModal = ({ entity, files, isLoading, onClose, onGenerate, isGenerating }) => {
  const [selectedFile, setSelectedFile] = useState(0)

  const getLayerColor = (layer) => {
    const colors = {
      model: 'bg-blue-100 text-blue-700',
      repository: 'bg-green-100 text-green-700',
      service: 'bg-purple-100 text-purple-700',
      handler: 'bg-orange-100 text-orange-700',
      dto: 'bg-pink-100 text-pink-700',
      migration: 'bg-yellow-100 text-yellow-700',
    }
    return colors[layer] || 'bg-gray-100 text-gray-700'
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl w-full max-w-6xl max-h-[90vh] flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Generated Code Preview</h2>
            <p className="text-sm text-gray-500">Entity: <span className="font-medium text-indigo-600">{entity.name}</span></p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={onGenerate}
              disabled={isGenerating || isLoading}
              className="btn btn-primary"
            >
              {isGenerating ? (
                <>
                  <LoadingSpinner size="sm" />
                  Generating...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4 mr-1.5" />
                  Generate Files
                </>
              )}
            </button>
            <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-lg">
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {isLoading ? (
          <div className="flex-1 flex items-center justify-center">
            <LoadingSpinner size="lg" />
          </div>
        ) : files.length === 0 ? (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            <div className="text-center">
              <Code className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>No files to preview</p>
            </div>
          </div>
        ) : (
          <div className="flex-1 flex overflow-hidden">
            {/* File List Sidebar */}
            <div className="w-64 border-r bg-gray-50 overflow-y-auto">
              <div className="p-3">
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">
                  Files ({files.length})
                </p>
                <div className="space-y-1">
                  {files.map((file, idx) => (
                    <button
                      key={idx}
                      onClick={() => setSelectedFile(idx)}
                      className={`w-full text-left px-3 py-2.5 rounded-lg transition-colors ${
                        selectedFile === idx
                          ? 'bg-indigo-100 text-indigo-900'
                          : 'hover:bg-gray-100 text-gray-700'
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        <FileCode className="w-4 h-4 flex-shrink-0 opacity-60" />
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium truncate">
                            {file.path.split('/').pop()}
                          </p>
                          <span className={`inline-block text-[10px] px-1.5 py-0.5 rounded mt-0.5 ${getLayerColor(file.layer)}`}>
                            {file.layer}
                          </span>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Code Preview */}
            <div className="flex-1 overflow-hidden flex flex-col">
              <div className="px-4 py-2 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
                <span className="text-sm font-mono text-gray-600">{files[selectedFile]?.path}</span>
                <button
                  onClick={() => navigator.clipboard.writeText(files[selectedFile]?.content || '')}
                  className="text-xs px-2 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded"
                >
                  <Copy className="w-3.5 h-3.5 inline mr-1" />
                  Copy
                </button>
              </div>
              <div className="flex-1 overflow-auto">
                <CodeEditor
                  code={files[selectedFile]?.content || ''}
                  filename={files[selectedFile]?.path || ''}
                  showLineNumbers={true}
                  className="h-full border-0 rounded-none"
                />
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
