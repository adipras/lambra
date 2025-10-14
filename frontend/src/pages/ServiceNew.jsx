import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCreateProject } from '../hooks/useProjects'
import { ErrorAlert } from '../components/shared/ErrorAlert'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'

export const ServiceNew = () => {
  const navigate = useNavigate()
  const createProject = useCreateProject()

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    namespace: '',
  })
  const [error, setError] = useState(null)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)

    try {
      await createProject.mutateAsync(formData)
      navigate('/services')
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to create service')
    }
  }

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  return (
    <div className="max-w-2xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Create New Service</h1>
        <p className="mt-1 text-gray-600">
          Define your microservice configuration
        </p>
      </div>

      {error && <ErrorAlert message={error} onClose={() => setError(null)} />}

      <form onSubmit={handleSubmit} className="card">
        <div className="space-y-6">
          {/* Service Name */}
          <div>
            <label htmlFor="name" className="label">
              Service Name *
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              className="input"
              placeholder="e.g., User Service, Payment Service"
              required
              minLength={3}
              maxLength={100}
            />
            <p className="mt-1 text-sm text-gray-500">
              A descriptive name for your microservice
            </p>
          </div>

          {/* Namespace */}
          <div>
            <label htmlFor="namespace" className="label">
              Namespace *
            </label>
            <input
              type="text"
              id="namespace"
              name="namespace"
              value={formData.namespace}
              onChange={handleChange}
              className="input"
              placeholder="e.g., user-ns, payment-ns"
              required
              minLength={3}
              maxLength={50}
              pattern="[a-z0-9-]+"
            />
            <p className="mt-1 text-sm text-gray-500">
              Kubernetes namespace (lowercase, numbers, and hyphens only)
            </p>
          </div>

          {/* Description */}
          <div>
            <label htmlFor="description" className="label">
              Description
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              className="input"
              rows={4}
              placeholder="Brief description of what this service does..."
              maxLength={500}
            />
            <p className="mt-1 text-sm text-gray-500">
              Optional description (max 500 characters)
            </p>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-4 mt-8">
          <button
            type="submit"
            disabled={createProject.isPending}
            className="btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {createProject.isPending ? (
              <>
                <LoadingSpinner size="sm" className="inline mr-2" />
                Creating...
              </>
            ) : (
              'Create Service'
            )}
          </button>

          <button
            type="button"
            onClick={() => navigate('/services')}
            className="btn btn-secondary"
            disabled={createProject.isPending}
          >
            Cancel
          </button>
        </div>
      </form>

      {/* Info Box */}
      <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <h3 className="text-sm font-medium text-blue-900 mb-2">
          What happens next?
        </h3>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>• Service project will be created in the database</li>
          <li>• You can then define entities and endpoints</li>
          <li>• Generate the actual microservice code</li>
          <li>• Deploy to Docker (local) or Kubernetes</li>
        </ul>
      </div>
    </div>
  )
}
