import React, { useState } from 'react'
import { Plus, Trash2, X } from 'lucide-react'

const FIELD_TYPES = [
  { value: 'string', label: 'String' },
  { value: 'int', label: 'Integer' },
  { value: 'float', label: 'Float' },
  { value: 'bool', label: 'Boolean' },
  { value: 'date', label: 'Date' },
  { value: 'datetime', label: 'DateTime' },
  { value: 'json', label: 'JSON' },
]

export const EntityForm = ({ onSubmit, onCancel, initialData = null, isLoading = false }) => {
  const [formData, setFormData] = useState({
    name: initialData?.name || '',
    table_name: initialData?.table_name || '',
    description: initialData?.description || '',
    fields: initialData?.fields || [
      { name: '', type: 'string', required: false, unique: false, length: 255, description: '' }
    ],
  })

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  const handleFieldChange = (index, field, value) => {
    const newFields = [...formData.fields]
    newFields[index] = { ...newFields[index], [field]: value }
    setFormData(prev => ({ ...prev, fields: newFields }))
  }

  const addField = () => {
    setFormData(prev => ({
      ...prev,
      fields: [...prev.fields, { name: '', type: 'string', required: false, unique: false, length: 255, description: '' }]
    }))
  }

  const removeField = (index) => {
    setFormData(prev => ({
      ...prev,
      fields: prev.fields.filter((_, i) => i !== index)
    }))
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit(formData)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Info */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-gray-900">Basic Information</h3>

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
            required
            className="input"
            placeholder="e.g., User, Product, Order"
          />
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
            required
            className="input"
            placeholder="e.g., users, products, orders"
          />
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
            rows={3}
            className="input"
            placeholder="Describe this entity..."
          />
        </div>
      </div>

      {/* Fields */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Fields</h3>
          <button
            type="button"
            onClick={addField}
            className="btn btn-secondary text-sm"
          >
            <Plus className="w-4 h-4 mr-1" />
            Add Field
          </button>
        </div>

        <div className="space-y-3">
          {formData.fields.map((field, index) => (
            <div key={index} className="border border-gray-200 rounded-lg p-4 space-y-3">
              <div className="flex items-start justify-between">
                <h4 className="text-sm font-medium text-gray-700">Field {index + 1}</h4>
                {formData.fields.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeField(index)}
                    className="text-red-600 hover:text-red-700"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">
                    Field Name
                  </label>
                  <input
                    type="text"
                    value={field.name}
                    onChange={(e) => handleFieldChange(index, 'name', e.target.value)}
                    className="input text-sm"
                    placeholder="e.g., email"
                    required
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">
                    Type
                  </label>
                  <select
                    value={field.type}
                    onChange={(e) => handleFieldChange(index, 'type', e.target.value)}
                    className="input text-sm"
                  >
                    {FIELD_TYPES.map(type => (
                      <option key={type.value} value={type.value}>{type.label}</option>
                    ))}
                  </select>
                </div>

                {(field.type === 'string') && (
                  <div>
                    <label className="block text-xs font-medium text-gray-600 mb-1">
                      Max Length
                    </label>
                    <input
                      type="number"
                      value={field.length || 255}
                      onChange={(e) => handleFieldChange(index, 'length', parseInt(e.target.value))}
                      className="input text-sm"
                      min="1"
                    />
                  </div>
                )}

                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">
                    Description
                  </label>
                  <input
                    type="text"
                    value={field.description || ''}
                    onChange={(e) => handleFieldChange(index, 'description', e.target.value)}
                    className="input text-sm"
                    placeholder="Field description"
                  />
                </div>
              </div>

              <div className="flex gap-4">
                <label className="flex items-center text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={field.required}
                    onChange={(e) => handleFieldChange(index, 'required', e.target.checked)}
                    className="mr-2"
                  />
                  Required
                </label>
                <label className="flex items-center text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={field.unique}
                    onChange={(e) => handleFieldChange(index, 'unique', e.target.checked)}
                    className="mr-2"
                  />
                  Unique
                </label>
              </div>
            </div>
          ))}
        </div>
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
          disabled={isLoading}
        >
          {isLoading ? 'Saving...' : (initialData ? 'Update Entity' : 'Create Entity')}
        </button>
      </div>
    </form>
  )
}
