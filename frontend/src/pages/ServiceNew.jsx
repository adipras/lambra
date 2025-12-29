import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCreateProject } from '../hooks/useProjects'
import { ErrorAlert } from '../components/shared/ErrorAlert'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'
import { Database, AlertCircle, CheckCircle } from 'lucide-react'

export const ServiceNew = () => {
  const navigate = useNavigate()
  const createProject = useCreateProject()

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    namespace: '',
    // Database configuration
    db_host: 'host.docker.internal',
    db_port: 3306,
    db_user: 'root',
    db_password: '',
    db_name: '',
  })
  const [error, setError] = useState(null)
  const [dbValidated, setDbValidated] = useState(false)

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
    const { name, value, type } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value
    }))
    // Reset validation when db config changes
    if (name.startsWith('db_')) {
      setDbValidated(false)
    }
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

          {/* Database Configuration Section */}
          <div className="border-t border-gray-200 pt-6">
            <div className="flex items-center gap-2 mb-4">
              <div className="p-2 bg-indigo-100 rounded-lg">
                <Database className="w-5 h-5 text-indigo-600" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Database Configuration</h3>
                <p className="text-sm text-gray-500">
                  MySQL database for the generated service (connection will be validated)
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* DB Host */}
              <div>
                <label htmlFor="db_host" className="label">
                  Database Host *
                </label>
                <input
                  type="text"
                  id="db_host"
                  name="db_host"
                  value={formData.db_host}
                  onChange={handleChange}
                  className="input"
                  placeholder="localhost or host.docker.internal"
                  required
                />
                <p className="mt-1 text-xs text-gray-500">
                  Use "host.docker.internal" for local MySQL
                </p>
              </div>

              {/* DB Port */}
              <div>
                <label htmlFor="db_port" className="label">
                  Port *
                </label>
                <input
                  type="number"
                  id="db_port"
                  name="db_port"
                  value={formData.db_port}
                  onChange={handleChange}
                  className="input"
                  placeholder="3306"
                  required
                  min={1}
                  max={65535}
                />
              </div>

              {/* DB User */}
              <div>
                <label htmlFor="db_user" className="label">
                  Username *
                </label>
                <input
                  type="text"
                  id="db_user"
                  name="db_user"
                  value={formData.db_user}
                  onChange={handleChange}
                  className="input"
                  placeholder="root"
                  required
                />
              </div>

              {/* DB Password */}
              <div>
                <label htmlFor="db_password" className="label">
                  Password *
                </label>
                <input
                  type="password"
                  id="db_password"
                  name="db_password"
                  value={formData.db_password}
                  onChange={handleChange}
                  className="input"
                  placeholder="••••••••"
                  required
                />
              </div>

              {/* DB Name */}
              <div className="col-span-2">
                <label htmlFor="db_name" className="label">
                  Database Name *
                </label>
                <input
                  type="text"
                  id="db_name"
                  name="db_name"
                  value={formData.db_name}
                  onChange={handleChange}
                  className="input"
                  placeholder="my_service_db"
                  required
                  pattern="[a-zA-Z0-9_]+"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Database will be created if it doesn't exist
                </p>
              </div>
            </div>

            {/* Database validation info */}
            <div className={`mt-4 p-3 rounded-lg flex items-start gap-2 ${
              dbValidated
                ? 'bg-green-50 border border-green-200'
                : 'bg-yellow-50 border border-yellow-200'
            }`}>
              {dbValidated ? (
                <>
                  <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0" />
                  <p className="text-sm text-green-700">Database connection validated</p>
                </>
              ) : (
                <>
                  <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0" />
                  <p className="text-sm text-yellow-700">
                    Database connection will be validated when you create the service
                  </p>
                </>
              )}
            </div>
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
          <li>• Database connection will be validated</li>
          <li>• Database will be created if it doesn't exist</li>
          <li>• Service project will be created with your configuration</li>
          <li>• You can then define entities and endpoints</li>
          <li>• Generated service will auto-migrate tables on startup</li>
        </ul>
      </div>
    </div>
  )
}
