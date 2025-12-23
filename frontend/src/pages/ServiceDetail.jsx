import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, Code, Play, Eye, X, FileCode, Check, AlertCircle, Square, Rocket, ExternalLink, RefreshCw, ChevronRight, Download, FileJson, ChevronDown } from 'lucide-react'
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
import { EntityForm } from '../components/forms/EntityForm'
import { EndpointForm } from '../components/forms/EndpointForm'
import { CodeEditor } from '../components/code/CodeEditor'

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

  // Create entity mutation
  const createEntityMutation = useMutation({
    mutationFn: (data) => entitiesApi.create(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['entities', id])
      setShowEntityModal(false)
    },
  })

  // Create endpoint mutation
  const createEndpointMutation = useMutation({
    mutationFn: (data) => endpointsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries(['entity-endpoints'])
      setShowEndpointModal(false)
      setSelectedEntity(null)
    },
  })

  // Delete entity mutation
  const deleteEntityMutation = useMutation({
    mutationFn: (entityId) => entitiesApi.delete(entityId),
    onSuccess: () => {
      queryClient.invalidateQueries(['entities', id])
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
      // Poll every 5 seconds if service is running
      return data?.data?.status === 'running' ? 5000 : false
    },
  })

  const deploymentStatus = statusData?.data

  // Deploy mutation
  const deployMutation = useMutation({
    mutationFn: () => deploymentApi.deploy(id),
    onSuccess: (response) => {
      showNotification('success', `Service deployed! URL: ${response.data?.data?.url || 'N/A'}`)
      refetchStatus()
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.message || 'Failed to deploy service')
    },
  })

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
      showNotification('success', `Service redeployed! URL: ${response.data?.data?.url || 'N/A'}`)
      refetchStatus()
    },
    onError: (error) => {
      showNotification('error', error.response?.data?.message || 'Failed to redeploy service')
    },
  })

  const isDeploying = deployMutation.isPending
  const isStarting = startMutation.isPending
  const isStopping = stopMutation.isPending
  const isRedeploying = redeployMutation.isPending

  const [showExportMenu, setShowExportMenu] = useState(false)
  const [isExporting, setIsExporting] = useState(false)

  const handleExportOpenAPI = async () => {
    setIsExporting(true)
    try {
      await exportApi.downloadOpenAPI(id)
      showNotification('success', 'OpenAPI specification downloaded!')
    } catch (error) {
      showNotification('error', 'Failed to export OpenAPI specification')
    } finally {
      setIsExporting(false)
      setShowExportMenu(false)
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
      setShowExportMenu(false)
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

  const handleDeleteEntity = (entityId) => {
    if (window.confirm('Are you sure you want to delete this entity?')) {
      deleteEntityMutation.mutate(entityId)
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
        <div className={`fixed top-4 right-4 z-50 flex items-center gap-2 px-4 py-3 rounded-lg shadow-lg ${
          notification.type === 'success'
            ? 'bg-green-50 text-green-800 border border-green-200'
            : 'bg-red-50 text-red-800 border border-red-200'
        }`}>
          {notification.type === 'success' ? (
            <Check className="w-5 h-5" />
          ) : (
            <AlertCircle className="w-5 h-5" />
          )}
          <span>{notification.message}</span>
          <button onClick={() => setNotification(null)} className="ml-2">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/services')}
            className="btn btn-secondary"
          >
            <ArrowLeft className="w-4 h-4" />
          </button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{project?.name}</h1>
            <p className="text-gray-600 mt-1">{project?.description || 'No description'}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <StatusBadge status={project?.status} />

          {/* Deployment Status & Controls */}
          <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg">
            <div className={`w-2 h-2 rounded-full ${
              deploymentStatus?.status === 'running' ? 'bg-green-500 animate-pulse' :
              deploymentStatus?.status === 'stopped' ? 'bg-yellow-500' : 'bg-gray-400'
            }`} />
            <span className="text-sm font-medium text-gray-700 capitalize">
              {deploymentStatus?.status || 'not deployed'}
            </span>
            {deploymentStatus?.url && deploymentStatus?.status === 'running' && (
              <a
                href={deploymentStatus.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:text-blue-700"
              >
                <ExternalLink className="w-4 h-4" />
              </a>
            )}
          </div>

          {/* Action Buttons */}
          {deploymentStatus?.status === 'running' ? (
            <>
              <button
                onClick={() => stopMutation.mutate()}
                disabled={isStopping || isRedeploying}
                className="btn bg-red-600 hover:bg-red-700 text-white flex items-center gap-2"
              >
                {isStopping ? (
                  <>
                    <LoadingSpinner size="sm" />
                    Stopping...
                  </>
                ) : (
                  <>
                    <Square className="w-4 h-4" />
                    Stop
                  </>
                )}
              </button>
              <button
                onClick={() => redeployMutation.mutate()}
                disabled={isRedeploying || isStopping}
                className="btn bg-orange-600 hover:bg-orange-700 text-white flex items-center gap-2"
                title="Redeploy (clear cache)"
              >
                {isRedeploying ? (
                  <>
                    <LoadingSpinner size="sm" />
                    Redeploying...
                  </>
                ) : (
                  <>
                    <RefreshCw className="w-4 h-4" />
                    Redeploy
                  </>
                )}
              </button>
            </>
          ) : deploymentStatus?.status === 'stopped' ? (
            <>
              <button
                onClick={() => startMutation.mutate()}
                disabled={isStarting || isRedeploying}
                className="btn bg-green-600 hover:bg-green-700 text-white flex items-center gap-2"
              >
                {isStarting ? (
                  <>
                    <LoadingSpinner size="sm" />
                    Starting...
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4" />
                    Start
                  </>
                )}
              </button>
              <button
                onClick={() => redeployMutation.mutate()}
                disabled={isRedeploying || isStarting}
                className="btn bg-orange-600 hover:bg-orange-700 text-white flex items-center gap-2"
                title="Redeploy (clear cache)"
              >
                {isRedeploying ? (
                  <>
                    <LoadingSpinner size="sm" />
                    Redeploying...
                  </>
                ) : (
                  <>
                    <RefreshCw className="w-4 h-4" />
                    Redeploy
                  </>
                )}
              </button>
            </>
          ) : (
            <button
              onClick={() => deployMutation.mutate()}
              disabled={isDeploying || entities.length === 0}
              className="btn bg-indigo-600 hover:bg-indigo-700 text-white flex items-center gap-2"
            >
              {isDeploying ? (
                <>
                  <LoadingSpinner size="sm" />
                  Deploying...
                </>
              ) : (
                <>
                  <Rocket className="w-4 h-4" />
                  Deploy
                </>
              )}
            </button>
          )}

          <button
            onClick={() => refetchStatus()}
            className="btn btn-secondary p-2"
            title="Refresh Status"
          >
            <RefreshCw className="w-4 h-4" />
          </button>

          {/* Export Dropdown */}
          <div className="relative">
            <button
              onClick={() => setShowExportMenu(!showExportMenu)}
              disabled={isExporting || entities.length === 0}
              className="btn btn-secondary flex items-center gap-2"
            >
              {isExporting ? (
                <LoadingSpinner size="sm" />
              ) : (
                <Download className="w-4 h-4" />
              )}
              Export
              <ChevronDown className="w-4 h-4" />
            </button>
            {showExportMenu && (
              <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-10">
                <button
                  onClick={handleExportOpenAPI}
                  className="w-full px-4 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2"
                >
                  <FileJson className="w-4 h-4 text-blue-600" />
                  OpenAPI Spec
                </button>
                <button
                  onClick={handleExportPostman}
                  className="w-full px-4 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2"
                >
                  <FileJson className="w-4 h-4 text-orange-600" />
                  Postman Collection
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Project Info */}
      <div className="card grid grid-cols-4 gap-6">
        <div>
          <p className="text-sm text-gray-600">Namespace</p>
          <p className="font-medium text-gray-900">{project?.namespace}</p>
        </div>
        <div>
          <p className="text-sm text-gray-600">Entities</p>
          <p className="font-medium text-gray-900">{entities.length}</p>
        </div>
        <div>
          <p className="text-sm text-gray-600">Created</p>
          <p className="font-medium text-gray-900">
            {new Date(project?.created_at).toLocaleDateString()}
          </p>
        </div>
        <div>
          <p className="text-sm text-gray-600">Service URL</p>
          {deploymentStatus?.url && deploymentStatus?.status === 'running' ? (
            <a
              href={deploymentStatus.url}
              target="_blank"
              rel="noopener noreferrer"
              className="font-medium text-blue-600 hover:text-blue-700 flex items-center gap-1"
            >
              {deploymentStatus.url}
              <ExternalLink className="w-3 h-3" />
            </a>
          ) : (
            <p className="font-medium text-gray-400">Not running</p>
          )}
        </div>
      </div>

      {/* Entities Section */}
      <div className="card">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-bold text-gray-900">Entities</h2>
          <button
            onClick={() => setShowEntityModal(true)}
            className="btn btn-primary"
          >
            <Plus className="w-4 h-4 mr-1" />
            Add Entity
          </button>
        </div>

        {entitiesLoading ? (
          <LoadingSpinner />
        ) : entities.length === 0 ? (
          <div className="text-center py-12">
            <Code className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No entities yet</p>
            <button
              onClick={() => setShowEntityModal(true)}
              className="btn btn-primary mt-4"
            >
              Create your first entity
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            {entities.map((entity) => (
              <EntityCard
                key={entity.id}
                entity={entity}
                onAddEndpoint={() => {
                  setSelectedEntity(entity)
                  setShowEndpointModal(true)
                }}
                onDelete={() => handleDeleteEntity(entity.id)}
                onDeleteEndpoint={handleDeleteEndpoint}
                onPreview={() => handlePreview(entity)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Entity Modal */}
      {showEntityModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-3xl w-full max-h-[90vh] overflow-y-auto p-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">Create Entity</h2>
            <EntityForm
              onSubmit={(data) => createEntityMutation.mutate(data)}
              onCancel={() => setShowEntityModal(false)}
              isLoading={createEntityMutation.isPending}
            />
          </div>
        </div>
      )}

      {/* Endpoint Modal */}
      {showEndpointModal && selectedEntity && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-3xl w-full max-h-[90vh] overflow-y-auto p-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Create Endpoint</h2>
            <p className="text-gray-600 mb-6">for entity: {selectedEntity.name}</p>
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
    </div>
  )
}

// EntityCard component
const EntityCard = ({ entity, onAddEndpoint, onDelete, onDeleteEndpoint, onPreview }) => {
  const { data: endpointsData } = useQuery({
    queryKey: ['entity-endpoints', entity.id],
    queryFn: () => endpointsApi.getByEntity(entity.id),
  })

  const endpoints = endpointsData?.data || []

  const getMethodBadgeColor = (method) => {
    const colors = {
      GET: 'bg-blue-100 text-blue-800 border-blue-200',
      POST: 'bg-green-100 text-green-800 border-green-200',
      PUT: 'bg-yellow-100 text-yellow-800 border-yellow-200',
      DELETE: 'bg-red-100 text-red-800 border-red-200',
      PATCH: 'bg-purple-100 text-purple-800 border-purple-200',
    }
    return colors[method] || 'bg-gray-100 text-gray-800 border-gray-200'
  }

  return (
    <div className="border border-gray-200 rounded-lg p-4">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="font-semibold text-gray-900">{entity.name}</h3>
          <p className="text-sm text-gray-600">Table: {entity.table_name}</p>
          {entity.description && (
            <p className="text-sm text-gray-500 mt-1">{entity.description}</p>
          )}
        </div>
        <div className="flex gap-2">
          <button
            onClick={onPreview}
            className="text-indigo-600 hover:text-indigo-700"
            title="Preview Generated Code"
          >
            <Eye className="w-4 h-4" />
          </button>
          <button
            onClick={onAddEndpoint}
            className="text-blue-600 hover:text-blue-700"
            title="Add Endpoint"
          >
            <Plus className="w-4 h-4" />
          </button>
          <button
            onClick={onDelete}
            className="text-red-600 hover:text-red-700"
            title="Delete Entity"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Fields */}
      <div className="mb-3">
        <p className="text-xs font-medium text-gray-600 mb-2">Fields:</p>
        <div className="flex flex-wrap gap-2">
          {entity.fields?.map((field, idx) => (
            <span
              key={idx}
              className="inline-flex items-center px-2 py-1 rounded text-xs bg-gray-100 text-gray-700"
            >
              {field.name} <span className="text-gray-500 ml-1">({field.type})</span>
            </span>
          ))}
        </div>
      </div>

      {/* Endpoints */}
      {endpoints.length > 0 && (
        <div>
          <p className="text-xs font-medium text-gray-600 mb-2">Endpoints:</p>
          <div className="space-y-2">
            {endpoints.map((endpoint) => (
              <div
                key={endpoint.id}
                className="flex items-center justify-between bg-gray-50 rounded p-2 group hover:bg-gray-100 transition-colors"
              >
                <Link
                  to={`/endpoints/${endpoint.id}`}
                  className="flex items-center gap-2 flex-1"
                >
                  <span
                    className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${getMethodBadgeColor(endpoint.method)}`}
                  >
                    {endpoint.method}
                  </span>
                  <span className="text-sm font-medium text-gray-900">
                    {endpoint.name}
                  </span>
                  <span className="text-xs text-gray-500">
                    {endpoint.path}
                  </span>
                  <ChevronRight className="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity ml-auto" />
                </Link>
                <button
                  onClick={(e) => {
                    e.preventDefault()
                    onDeleteEndpoint(endpoint.id)
                  }}
                  className="text-red-600 hover:text-red-700 ml-2"
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// CodePreviewModal component
const CodePreviewModal = ({ entity, files, isLoading, onClose, onGenerate, isGenerating }) => {
  const [selectedFile, setSelectedFile] = useState(0)

  const getLayerColor = (layer) => {
    const colors = {
      model: 'bg-blue-100 text-blue-800',
      repository: 'bg-green-100 text-green-800',
      service: 'bg-purple-100 text-purple-800',
      handler: 'bg-orange-100 text-orange-800',
      dto: 'bg-pink-100 text-pink-800',
      migration: 'bg-yellow-100 text-yellow-800',
    }
    return colors[layer] || 'bg-gray-100 text-gray-800'
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg w-full max-w-6xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Generated Code Preview</h2>
            <p className="text-sm text-gray-600">Entity: {entity.name}</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={onGenerate}
              disabled={isGenerating || isLoading}
              className="btn btn-primary flex items-center gap-2"
            >
              {isGenerating ? (
                <>
                  <LoadingSpinner size="sm" />
                  Generating...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" />
                  Generate Files
                </>
              )}
            </button>
            <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
              <X className="w-6 h-6" />
            </button>
          </div>
        </div>

        {isLoading ? (
          <div className="flex-1 flex items-center justify-center">
            <LoadingSpinner size="lg" />
          </div>
        ) : files.length === 0 ? (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            No files to preview
          </div>
        ) : (
          <div className="flex-1 flex overflow-hidden">
            {/* File List Sidebar */}
            <div className="w-64 border-r bg-gray-50 overflow-y-auto">
              <div className="p-2">
                <p className="text-xs font-medium text-gray-500 uppercase px-2 py-1">
                  Files ({files.length})
                </p>
                {files.map((file, idx) => (
                  <button
                    key={idx}
                    onClick={() => setSelectedFile(idx)}
                    className={`w-full text-left px-3 py-2 rounded-lg mb-1 transition-colors ${
                      selectedFile === idx
                        ? 'bg-indigo-100 text-indigo-900'
                        : 'hover:bg-gray-100 text-gray-700'
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <FileCode className="w-4 h-4 flex-shrink-0" />
                      <div className="min-w-0">
                        <p className="text-sm font-medium truncate">
                          {file.path.split('/').pop()}
                        </p>
                        <span className={`inline-block text-xs px-1.5 py-0.5 rounded ${getLayerColor(file.layer)}`}>
                          {file.layer}
                        </span>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            </div>

            {/* Code Preview */}
            <div className="flex-1 overflow-hidden flex flex-col">
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
