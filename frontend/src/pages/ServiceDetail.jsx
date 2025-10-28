import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, Code } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { projectsApi } from '../api/projects'
import { entitiesApi } from '../api/entities'
import { endpointsApi } from '../api/endpoints'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'
import { ErrorAlert } from '../components/shared/ErrorAlert'
import { StatusBadge } from '../components/shared/StatusBadge'
import { EntityForm } from '../components/forms/EntityForm'
import { EndpointForm } from '../components/forms/EndpointForm'

export const ServiceDetail = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [showEntityModal, setShowEntityModal] = useState(false)
  const [showEndpointModal, setShowEndpointModal] = useState(false)
  const [selectedEntity, setSelectedEntity] = useState(null)

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
        <StatusBadge status={project?.status} />
      </div>

      {/* Project Info */}
      <div className="card grid grid-cols-3 gap-6">
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
    </div>
  )
}

// EntityCard component
const EntityCard = ({ entity, onAddEndpoint, onDelete, onDeleteEndpoint }) => {
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
                className="flex items-center justify-between bg-gray-50 rounded p-2"
              >
                <div className="flex items-center gap-2">
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
                </div>
                <button
                  onClick={() => onDeleteEndpoint(endpoint.id)}
                  className="text-red-600 hover:text-red-700"
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
