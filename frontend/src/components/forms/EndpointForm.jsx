import React, { useState } from 'react'
import { X } from 'lucide-react'

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']

export const EndpointForm = ({
  entityId,
  onSubmit,
  onCancel,
  initialData = null,
  isLoading = false
}) => {
  const [formData, setFormData] = useState({
    entity_id: entityId || initialData?.entity_id || '',
    name: initialData?.name || '',
    path: initialData?.path || '',
    method: initialData?.method || 'GET',
    description: initialData?.description || '',
    require_auth: initialData?.require_auth ?? true,
    request_schema: initialData?.request_schema ? JSON.stringify(initialData.request_schema, null, 2) : '{}',
    response_schema: initialData?.response_schema ? JSON.stringify(initialData.response_schema, null, 2) : '{}',
  })

  const [schemaError, setSchemaError] = useState({ request: '', response: '' })

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
  }

  const handleSchemaChange = (field, value) => {
    setFormData(prev => ({ ...prev, [field]: value }))

    // Validate JSON
    try {
      if (value.trim()) {
        JSON.parse(value)
      }
      setSchemaError(prev => ({ ...prev, [field === 'request_schema' ? 'request' : 'response']: '' }))
    } catch (err) {
      setSchemaError(prev => ({
        ...prev,
        [field === 'request_schema' ? 'request' : 'response']: 'Invalid JSON format'
      }))
    }
  }

  const handleSubmit = (e) => {
    e.preventDefault()

    // Validate schemas
    if (schemaError.request || schemaError.response) {
      return
    }

    // Parse JSON schemas
    const submitData = {
      ...formData,
      request_schema: formData.request_schema.trim() ? JSON.parse(formData.request_schema) : {},
      response_schema: formData.response_schema.trim() ? JSON.parse(formData.response_schema) : {},
    }

    onSubmit(submitData)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Info */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-gray-900">Endpoint Information</h3>

        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
            Endpoint Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            required
            className="input"
            placeholder="e.g., CreateUser, GetUserById"
          />
          <p className="text-xs text-gray-500 mt-1">A descriptive name for this endpoint</p>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="method" className="block text-sm font-medium text-gray-700 mb-1">
              HTTP Method <span className="text-red-500">*</span>
            </label>
            <select
              id="method"
              name="method"
              value={formData.method}
              onChange={handleChange}
              required
              className="input"
            >
              {HTTP_METHODS.map(method => (
                <option key={method} value={method}>{method}</option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="path" className="block text-sm font-medium text-gray-700 mb-1">
              Path <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="path"
              name="path"
              value={formData.path}
              onChange={handleChange}
              required
              className="input"
              placeholder="/users/:id"
            />
          </div>
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows={2}
            className="input"
            placeholder="Describe what this endpoint does..."
          />
        </div>

        <div>
          <label className="flex items-center text-sm text-gray-700">
            <input
              type="checkbox"
              name="require_auth"
              checked={formData.require_auth}
              onChange={handleChange}
              className="mr-2"
            />
            Require Authentication
          </label>
        </div>
      </div>

      {/* Request Schema */}
      <div className="space-y-2">
        <label htmlFor="request_schema" className="block text-sm font-medium text-gray-700">
          Request Schema (JSON)
        </label>
        <textarea
          id="request_schema"
          value={formData.request_schema}
          onChange={(e) => handleSchemaChange('request_schema', e.target.value)}
          rows={6}
          className={`input font-mono text-sm ${schemaError.request ? 'border-red-500' : ''}`}
          placeholder='{"email": "string", "password": "string"}'
        />
        {schemaError.request && (
          <p className="text-xs text-red-600">{schemaError.request}</p>
        )}
        <p className="text-xs text-gray-500">Define the expected request body structure</p>
      </div>

      {/* Response Schema */}
      <div className="space-y-2">
        <label htmlFor="response_schema" className="block text-sm font-medium text-gray-700">
          Response Schema (JSON)
        </label>
        <textarea
          id="response_schema"
          value={formData.response_schema}
          onChange={(e) => handleSchemaChange('response_schema', e.target.value)}
          rows={6}
          className={`input font-mono text-sm ${schemaError.response ? 'border-red-500' : ''}`}
          placeholder='{"id": "string", "email": "string", "created_at": "datetime"}'
        />
        {schemaError.response && (
          <p className="text-xs text-red-600">{schemaError.response}</p>
        )}
        <p className="text-xs text-gray-500">Define the expected response structure</p>
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-4 border-t">
        <button
          type="button"
          onClick={onCancel}
          className="btn btn-secondary"
          disabled={isLoading}
        >
          <X className="w-4 h-4 mr-1" />
          Cancel
        </button>
        <button
          type="submit"
          className="btn btn-primary"
          disabled={isLoading || schemaError.request || schemaError.response}
        >
          {isLoading ? 'Saving...' : (initialData ? 'Update Endpoint' : 'Create Endpoint')}
        </button>
      </div>
    </form>
  )
}
