import React, { useState, useEffect } from 'react'
import { Plus, Trash2, X, Check, Database, Zap, Info, ToggleLeft, ToggleRight, GripVertical, AlertCircle, Link2 } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { entitiesApi } from '../../api/entities'

const FIELD_TYPES = [
  { value: 'string', label: 'String', icon: 'Aa', color: 'bg-blue-100 text-blue-700' },
  { value: 'text', label: 'Text', icon: 'Tt', color: 'bg-blue-100 text-blue-700' },
  { value: 'int', label: 'Integer', icon: '#', color: 'bg-green-100 text-green-700' },
  { value: 'float', label: 'Float', icon: '.0', color: 'bg-green-100 text-green-700' },
  { value: 'bool', label: 'Boolean', icon: '?', color: 'bg-purple-100 text-purple-700' },
  { value: 'date', label: 'Date', icon: 'D', color: 'bg-orange-100 text-orange-700' },
  { value: 'datetime', label: 'DateTime', icon: 'DT', color: 'bg-orange-100 text-orange-700' },
  { value: 'json', label: 'JSON', icon: '{}', color: 'bg-yellow-100 text-yellow-700' },
  { value: 'relation', label: 'Relation', icon: '🔗', color: 'bg-pink-100 text-pink-700' },
]

const RELATION_TYPES = [
  { value: 'belongsTo', label: 'Belongs To', description: 'This entity has FK to related entity (Many-to-One)', example: 'Post belongs to User' },
  { value: 'hasOne', label: 'Has One', description: 'Related entity has FK to this entity (One-to-One)', example: 'User has one Profile' },
  { value: 'hasMany', label: 'Has Many', description: 'Related entities have FK to this entity (One-to-Many)', example: 'User has many Posts' },
  { value: 'manyToMany', label: 'Many to Many', description: 'Junction table connects both entities', example: 'Post has many Tags' },
]

const ON_DELETE_OPTIONS = [
  { value: 'CASCADE', label: 'Cascade', description: 'Delete related records' },
  { value: 'SET NULL', label: 'Set Null', description: 'Set FK to null' },
  { value: 'RESTRICT', label: 'Restrict', description: 'Prevent deletion if related' },
]

const ENDPOINT_OPTIONS = [
  { key: 'list', name: 'List All', method: 'GET', path: '/{entities}', description: 'Get paginated list with filtering', icon: '📋' },
  { key: 'get', name: 'Get by ID', method: 'GET', path: '/{entities}/:id', description: 'Get single record by ID', icon: '🔍' },
  { key: 'create', name: 'Create', method: 'POST', path: '/{entities}', description: 'Create new record', icon: '➕' },
  { key: 'update', name: 'Update', method: 'PUT', path: '/{entities}/:id', description: 'Update existing record', icon: '✏️' },
  { key: 'delete', name: 'Delete', method: 'DELETE', path: '/{entities}/:id', description: 'Soft delete record', icon: '🗑️' },
]

const METHOD_COLORS = {
  GET: 'bg-blue-500',
  POST: 'bg-green-500',
  PUT: 'bg-yellow-500',
  DELETE: 'bg-red-500',
}

export const EntityForm = ({ onSubmit, onCancel, initialData = null, isLoading = false, projectId = null }) => {
  const [formData, setFormData] = useState({
    name: initialData?.name || '',
    table_name: initialData?.table_name || '',
    description: initialData?.description || '',
    fields: initialData?.fields || [
      { name: '', type: 'string', required: false, unique: false, length: null, description: '' }
    ],
    generate_endpoints: {
      list: true,
      get: true,
      create: true,
      update: true,
      delete: true,
    },
  })

  // Fetch existing entities for relation dropdown
  const { data: entitiesData } = useQuery({
    queryKey: ['entities', projectId],
    queryFn: () => entitiesApi.getByProject(projectId),
    enabled: !!projectId,
  })

  // Filter out current entity from relation options
  const availableEntities = (entitiesData?.data || []).filter(
    e => e.id !== initialData?.id && e.name !== formData.name
  )

  const [errors, setErrors] = useState({})
  const [activeFieldIndex, setActiveFieldIndex] = useState(null)

  // Auto-generate table_name from entity name
  useEffect(() => {
    if (!initialData && formData.name) {
      const tableName = formData.name
        .replace(/([A-Z])/g, '_$1')
        .toLowerCase()
        .replace(/^_/, '')
        .replace(/\s+/g, '_') + 's'
      setFormData(prev => ({ ...prev, table_name: tableName }))
    }
  }, [formData.name, initialData])

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: null }))
    }
  }

  const handleFieldChange = (index, field, value) => {
    const newFields = [...formData.fields]
    newFields[index] = { ...newFields[index], [field]: value }
    setFormData(prev => ({ ...prev, fields: newFields }))
  }

  const addField = () => {
    const newIndex = formData.fields.length
    setFormData(prev => ({
      ...prev,
      fields: [...prev.fields, { name: '', type: 'string', required: false, unique: false, length: null, description: '' }]
    }))
    setActiveFieldIndex(newIndex)
  }

  const removeField = (index) => {
    setFormData(prev => ({
      ...prev,
      fields: prev.fields.filter((_, i) => i !== index)
    }))
    if (activeFieldIndex === index) {
      setActiveFieldIndex(null)
    }
  }

  const handleEndpointChange = (endpoint, checked) => {
    setFormData(prev => ({
      ...prev,
      generate_endpoints: {
        ...prev.generate_endpoints,
        [endpoint]: checked,
      },
    }))
  }

  const toggleAllEndpoints = (enabled) => {
    setFormData(prev => ({
      ...prev,
      generate_endpoints: {
        list: enabled,
        get: enabled,
        create: enabled,
        update: enabled,
        delete: enabled,
      },
    }))
  }

  const allEndpointsEnabled = Object.values(formData.generate_endpoints).every(v => v)
  const someEndpointsEnabled = Object.values(formData.generate_endpoints).some(v => v)
  const enabledEndpointsCount = Object.values(formData.generate_endpoints).filter(v => v).length

  const validateForm = () => {
    const newErrors = {}
    if (!formData.name.trim()) newErrors.name = 'Entity name is required'
    if (!formData.table_name.trim()) newErrors.table_name = 'Table name is required'
    if (formData.fields.some(f => !f.name.trim())) newErrors.fields = 'All fields must have a name'

    // Validate relation fields
    const invalidRelation = formData.fields.find(f =>
      f.type === 'relation' && (!f.relation_type || !f.related_entity)
    )
    if (invalidRelation) {
      newErrors.fields = 'Relation fields must have relation type and related entity selected'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    if (validateForm()) {
      onSubmit(formData)
    }
  }

  const getFieldTypeInfo = (type) => {
    return FIELD_TYPES.find(t => t.value === type) || FIELD_TYPES[0]
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-8">
      {/* Basic Info Section */}
      <div className="space-y-4">
        <div className="flex items-center gap-2 mb-4">
          <div className="p-2 bg-indigo-100 rounded-lg">
            <Database className="w-5 h-5 text-indigo-600" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900">Entity Information</h3>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Entity Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              className={`input ${errors.name ? 'border-red-500 focus:ring-red-500' : ''}`}
              placeholder="e.g., Product, Customer, Order"
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="w-3 h-3" /> {errors.name}
              </p>
            )}
            <p className="mt-1 text-xs text-gray-500">Use PascalCase (e.g., ProductCategory)</p>
          </div>

          <div>
            <label htmlFor="table_name" className="block text-sm font-medium text-gray-700 mb-1">
              Table Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="table_name"
              name="table_name"
              value={formData.table_name}
              onChange={handleChange}
              className={`input ${errors.table_name ? 'border-red-500 focus:ring-red-500' : ''}`}
              placeholder="e.g., products, customers"
            />
            {errors.table_name && (
              <p className="mt-1 text-sm text-red-500 flex items-center gap-1">
                <AlertCircle className="w-3 h-3" /> {errors.table_name}
              </p>
            )}
            <p className="mt-1 text-xs text-gray-500">Database table name (auto-generated)</p>
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
            placeholder="Brief description of this entity..."
          />
        </div>
      </div>

      {/* Fields Section */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <div className="p-2 bg-green-100 rounded-lg">
            <GripVertical className="w-5 h-5 text-green-600" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900">Fields</h3>
            <p className="text-sm text-gray-500">{formData.fields.length} field(s) defined</p>
          </div>
        </div>

        {errors.fields && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600 flex items-center gap-2">
            <AlertCircle className="w-4 h-4" /> {errors.fields}
          </div>
        )}

        <div className="space-y-3">
          {formData.fields.map((field, index) => {
            const typeInfo = getFieldTypeInfo(field.type)
            const isActive = activeFieldIndex === index

            return (
              <div
                key={index}
                className={`border rounded-lg transition-all ${
                  isActive
                    ? 'border-indigo-300 ring-2 ring-indigo-100 shadow-sm'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                {/* Field Header */}
                <div
                  className="flex items-center justify-between p-3 cursor-pointer"
                  onClick={() => setActiveFieldIndex(isActive ? null : index)}
                >
                  <div className="flex items-center gap-3">
                    <span className={`w-8 h-8 flex items-center justify-center rounded-lg text-xs font-bold ${typeInfo.color}`}>
                      {typeInfo.icon}
                    </span>
                    <div>
                      <p className="font-medium text-gray-900">
                        {field.name || <span className="text-gray-400 italic">Unnamed field</span>}
                      </p>
                      <div className="flex items-center gap-2 text-xs text-gray-500">
                        <span>{typeInfo.label}</span>
                        {field.type === 'relation' && field.related_entity && (
                          <span className="px-1.5 py-0.5 bg-pink-100 text-pink-700 rounded">
                            → {field.related_entity}
                          </span>
                        )}
                        {field.required && (
                          <span className="px-1.5 py-0.5 bg-red-100 text-red-700 rounded">Required</span>
                        )}
                        {field.unique && (
                          <span className="px-1.5 py-0.5 bg-purple-100 text-purple-700 rounded">Unique</span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {formData.fields.length > 1 && (
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation()
                          removeField(index)
                        }}
                        className="p-1 text-red-500 hover:bg-red-50 rounded"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </div>

                {/* Field Details (Expandable) */}
                {isActive && (
                  <div className="border-t border-gray-100 p-4 bg-gray-50 space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-gray-600 mb-1">
                          Field Name <span className="text-red-500">*</span>
                        </label>
                        <input
                          type="text"
                          value={field.name}
                          onChange={(e) => handleFieldChange(index, 'name', e.target.value)}
                          className="input text-sm"
                          placeholder="e.g., email, price, status"
                        />
                      </div>

                      <div>
                        <label className="block text-xs font-medium text-gray-600 mb-1">
                          Data Type
                        </label>
                        <div className="grid grid-cols-4 gap-1">
                          {FIELD_TYPES.map(type => (
                            <button
                              key={type.value}
                              type="button"
                              onClick={() => handleFieldChange(index, 'type', type.value)}
                              className={`px-2 py-1.5 text-xs font-medium rounded border transition-all ${
                                field.type === type.value
                                  ? `${type.color} border-current`
                                  : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'
                              }`}
                            >
                              {type.label}
                            </button>
                          ))}
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      {field.type === 'string' && (
                        <div>
                          <label className="block text-xs font-medium text-gray-600 mb-1">
                            Max Length
                          </label>
                          <input
                            type="number"
                            value={field.length !== null && field.length !== undefined ? field.length : 255}
                            onChange={(e) => {
                              const value = e.target.value === '' ? null : parseInt(e.target.value)
                              handleFieldChange(index, 'length', value)
                            }}
                            className="input text-sm"
                            min="1"
                            max="65535"
                            placeholder="255"
                          />
                        </div>
                      )}

                      <div className="col-span-2">
                        <label className="block text-xs font-medium text-gray-600 mb-1">
                          Description
                        </label>
                        <input
                          type="text"
                          value={field.description || ''}
                          onChange={(e) => handleFieldChange(index, 'description', e.target.value)}
                          className="input text-sm"
                          placeholder="What is this field for?"
                        />
                      </div>
                    </div>

                    {/* Relation-specific options */}
                    {field.type === 'relation' && (
                      <div className="space-y-4 p-4 bg-pink-50 rounded-lg border border-pink-200">
                        <div className="flex items-center gap-2 text-pink-700 font-medium">
                          <Link2 className="w-4 h-4" />
                          Relation Configuration
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                          <div>
                            <label className="block text-xs font-medium text-gray-600 mb-1">
                              Relation Type <span className="text-red-500">*</span>
                            </label>
                            <select
                              value={field.relation_type || ''}
                              onChange={(e) => handleFieldChange(index, 'relation_type', e.target.value)}
                              className="input text-sm"
                            >
                              <option value="">Select relation type...</option>
                              {RELATION_TYPES.map(rt => (
                                <option key={rt.value} value={rt.value}>{rt.label}</option>
                              ))}
                            </select>
                            {field.relation_type && (
                              <p className="text-xs text-gray-500 mt-1">
                                {RELATION_TYPES.find(rt => rt.value === field.relation_type)?.description}
                              </p>
                            )}
                          </div>

                          <div>
                            <label className="block text-xs font-medium text-gray-600 mb-1">
                              Related Entity <span className="text-red-500">*</span>
                            </label>
                            <select
                              value={field.related_entity || ''}
                              onChange={(e) => handleFieldChange(index, 'related_entity', e.target.value)}
                              className="input text-sm"
                            >
                              <option value="">Select entity...</option>
                              {availableEntities.map(entity => (
                                <option key={entity.id} value={entity.name}>{entity.name}</option>
                              ))}
                            </select>
                            {availableEntities.length === 0 && (
                              <p className="text-xs text-amber-600 mt-1">
                                No other entities available. Create another entity first.
                              </p>
                            )}
                          </div>
                        </div>

                        {(field.relation_type === 'belongsTo' || field.relation_type === 'manyToMany') && (
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <label className="block text-xs font-medium text-gray-600 mb-1">
                                Foreign Key Column
                              </label>
                              <input
                                type="text"
                                value={field.foreign_key || ''}
                                onChange={(e) => handleFieldChange(index, 'foreign_key', e.target.value)}
                                className="input text-sm"
                                placeholder={field.related_entity ? `${field.related_entity.toLowerCase()}_id` : 'Auto-generated'}
                              />
                              <p className="text-xs text-gray-500 mt-1">
                                Leave empty for auto-generated name
                              </p>
                            </div>

                            <div>
                              <label className="block text-xs font-medium text-gray-600 mb-1">
                                On Delete
                              </label>
                              <select
                                value={field.on_delete || 'CASCADE'}
                                onChange={(e) => handleFieldChange(index, 'on_delete', e.target.value)}
                                className="input text-sm"
                              >
                                {ON_DELETE_OPTIONS.map(opt => (
                                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                                ))}
                              </select>
                            </div>
                          </div>
                        )}

                        {field.relation_type && field.related_entity && (
                          <div className="p-3 bg-white rounded border border-pink-200 text-sm">
                            <p className="font-medium text-gray-700">Preview:</p>
                            <p className="text-gray-600 mt-1">
                              {field.relation_type === 'belongsTo' && (
                                <><code className="bg-gray-100 px-1 rounded">{formData.name || 'This entity'}</code> belongs to <code className="bg-gray-100 px-1 rounded">{field.related_entity}</code> (FK: <code className="bg-gray-100 px-1 rounded">{field.foreign_key || `${field.related_entity.toLowerCase()}_id`}</code>)</>
                              )}
                              {field.relation_type === 'hasOne' && (
                                <><code className="bg-gray-100 px-1 rounded">{formData.name || 'This entity'}</code> has one <code className="bg-gray-100 px-1 rounded">{field.related_entity}</code></>
                              )}
                              {field.relation_type === 'hasMany' && (
                                <><code className="bg-gray-100 px-1 rounded">{formData.name || 'This entity'}</code> has many <code className="bg-gray-100 px-1 rounded">{field.related_entity}</code></>
                              )}
                              {field.relation_type === 'manyToMany' && (
                                <><code className="bg-gray-100 px-1 rounded">{formData.name || 'This entity'}</code> ↔ <code className="bg-gray-100 px-1 rounded">{field.related_entity}</code> (junction table)</>
                              )}
                            </p>
                          </div>
                        )}
                      </div>
                    )}

                    {/* Standard field options (hide for relation type) */}
                    {field.type !== 'relation' && (
                      <div className="flex gap-4 pt-2">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={field.required}
                            onChange={(e) => handleFieldChange(index, 'required', e.target.checked)}
                            className="w-4 h-4 text-indigo-600 rounded focus:ring-indigo-500"
                          />
                          <span className="text-sm text-gray-700">Required field</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={field.unique}
                            onChange={(e) => handleFieldChange(index, 'unique', e.target.checked)}
                            className="w-4 h-4 text-indigo-600 rounded focus:ring-indigo-500"
                          />
                          <span className="text-sm text-gray-700">Unique constraint</span>
                        </label>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>

        {/* Add Field Button - at bottom for better UX */}
        <button
          type="button"
          onClick={addField}
          className="w-full py-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-500 hover:border-indigo-400 hover:text-indigo-600 hover:bg-indigo-50 transition-colors flex items-center justify-center gap-2"
        >
          <Plus className="w-5 h-5" />
          Add Field
        </button>
      </div>

      {/* Generate Endpoints Section */}
      {!initialData && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Zap className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Auto-Generate Endpoints</h3>
                <p className="text-sm text-gray-500">
                  {enabledEndpointsCount} endpoint(s) will be created with request/response schemas
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => toggleAllEndpoints(!allEndpointsEnabled)}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                allEndpointsEnabled
                  ? 'bg-purple-100 text-purple-700 hover:bg-purple-200'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {allEndpointsEnabled ? (
                <>
                  <ToggleRight className="w-4 h-4" />
                  All Selected
                </>
              ) : (
                <>
                  <ToggleLeft className="w-4 h-4" />
                  Select All
                </>
              )}
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {ENDPOINT_OPTIONS.map((endpoint) => {
              const isEnabled = formData.generate_endpoints[endpoint.key]
              return (
                <label
                  key={endpoint.key}
                  className={`relative flex items-start p-4 rounded-xl border-2 cursor-pointer transition-all ${
                    isEnabled
                      ? 'border-purple-300 bg-purple-50'
                      : 'border-gray-200 bg-white hover:border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={isEnabled}
                    onChange={(e) => handleEndpointChange(endpoint.key, e.target.checked)}
                    className="sr-only"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-lg">{endpoint.icon}</span>
                      <span className={`px-2 py-0.5 text-xs font-bold text-white rounded ${METHOD_COLORS[endpoint.method]}`}>
                        {endpoint.method}
                      </span>
                      <span className="font-medium text-gray-900">{endpoint.name}</span>
                    </div>
                    <p className="text-xs text-gray-500 font-mono">{endpoint.path}</p>
                    <p className="text-xs text-gray-400 mt-1">{endpoint.description}</p>
                  </div>
                  <div className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ml-2 ${
                    isEnabled ? 'bg-purple-500' : 'bg-gray-200'
                  }`}>
                    {isEnabled && <Check className="w-3 h-3 text-white" />}
                  </div>
                </label>
              )
            })}
          </div>

          {someEndpointsEnabled && (
            <div className="p-4 bg-blue-50 border border-blue-200 rounded-xl">
              <div className="flex gap-3">
                <Info className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-blue-800">
                  <p className="font-medium mb-1">Generated endpoints will include:</p>
                  <ul className="list-disc list-inside space-y-0.5 text-blue-700">
                    <li>Request body JSON Schema with validation rules</li>
                    <li>Response JSON Schema with example values</li>
                    <li>Auto-generated descriptions</li>
                    <li>Pagination for list endpoints</li>
                  </ul>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center justify-between pt-6 border-t border-gray-200">
        <p className="text-sm text-gray-500">
          {formData.fields.filter(f => f.name.trim()).length} fields configured
        </p>
        <div className="flex gap-3">
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
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                Saving...
              </>
            ) : (
              <>
                <Check className="w-4 h-4 mr-1" />
                {initialData ? 'Update Entity' : 'Create Entity'}
              </>
            )}
          </button>
        </div>
      </div>
    </form>
  )
}
