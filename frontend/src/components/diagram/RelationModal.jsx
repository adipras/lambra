import React, { useState, useEffect } from 'react';
import { X } from 'lucide-react';

const RELATION_TYPES = [
  {
    value: 'belongsTo',
    label: 'Belongs To',
    description: 'Source entity will have a foreign key to target entity',
    example: 'Post belongs to User (post.user_id → user.id)',
    color: 'pink'
  },
  {
    value: 'hasOne',
    label: 'Has One',
    description: 'Target entity will have a foreign key to source entity',
    example: 'User has one Profile (profile.user_id → user.id)',
    color: 'purple'
  },
  {
    value: 'hasMany',
    label: 'Has Many',
    description: 'Target entities will have foreign keys to source entity',
    example: 'User has many Posts (post.user_id → user.id)',
    color: 'blue'
  },
  {
    value: 'manyToMany',
    label: 'Many to Many',
    description: 'Junction table will connect both entities',
    example: 'Post has many Tags (via post_tags junction table)',
    color: 'orange'
  }
];

const ON_DELETE_OPTIONS = [
  { value: 'CASCADE', label: 'CASCADE', description: 'Delete related records automatically' },
  { value: 'SET NULL', label: 'SET NULL', description: 'Set foreign key to NULL' },
  { value: 'RESTRICT', label: 'RESTRICT', description: 'Prevent deletion if related records exist' },
  { value: 'NO ACTION', label: 'NO ACTION', description: 'Same as RESTRICT (SQL standard)' }
];

const ON_UPDATE_OPTIONS = [
  { value: 'CASCADE', label: 'CASCADE', description: 'Update foreign key automatically' },
  { value: 'SET NULL', label: 'SET NULL', description: 'Set foreign key to NULL' },
  { value: 'RESTRICT', label: 'RESTRICT', description: 'Prevent update if related records exist' },
  { value: 'NO ACTION', label: 'NO ACTION', description: 'Same as RESTRICT (SQL standard)' }
];

export default function RelationModal({ isOpen, onClose, onSubmit, sourceEntity, targetEntity }) {
  const [formData, setFormData] = useState({
    source_entity_id: '',
    target_entity_id: '',
    field_name: '',
    relation_type: 'belongsTo',
    on_delete: 'RESTRICT',
    on_update: 'CASCADE',
    junction_table: '',
    description: '',
    required: false
  });

  const [errors, setErrors] = useState({});
  const [availableFields, setAvailableFields] = useState([]);

  // Get available fields from the entity that will have the FK
  const getAvailableFields = (relationType) => {
    if (!sourceEntity || !targetEntity) return [];

    let entityWithFK = null;
    let referencedEntity = null;

    switch (relationType) {
      case 'belongsTo':
        entityWithFK = sourceEntity;
        referencedEntity = targetEntity;
        break;
      case 'hasOne':
      case 'hasMany':
        entityWithFK = targetEntity;
        referencedEntity = sourceEntity;
        break;
      default:
        return [];
    }

    // Parse fields
    let fields = [];
    try {
      fields = JSON.parse(entityWithFK.fields || '[]');
    } catch (e) {
      fields = [];
    }

    // Filter for int/int64 fields or suggest new field
    const intFields = fields
      .filter(f => f.type === 'int' || f.type === 'int64')
      .map(f => ({
        value: toSnakeCase(f.name),
        label: toSnakeCase(f.name),
        type: f.type,
        isExisting: true,
        description: `Existing ${f.type} field (will be BIGINT FK)`
      }));

    // Add suggested new field
    const suggestedName = `${toSnakeCase(referencedEntity.name)}_id`;
    const suggestedField = {
      value: suggestedName,
      label: suggestedName,
      type: 'bigint',
      isExisting: false,
      description: 'New FK field (recommended)'
    };

    // Check if suggested name already exists
    const existingSuggested = intFields.find(f => f.value === suggestedName);
    if (existingSuggested) {
      // Mark existing field as recommended
      existingSuggested.description = 'Existing field (recommended, will be FK)';
      return intFields;
    }

    return [suggestedField, ...intFields];
  };

  // Update available fields when relation type or entities change
  useEffect(() => {
    const fields = getAvailableFields(formData.relation_type);
    setAvailableFields(fields);
    
    // Auto-select first field (recommended)
    if (fields.length > 0 && !formData.field_name) {
      setFormData(prev => ({
        ...prev,
        field_name: fields[0].value
      }));
    }
  }, [formData.relation_type, sourceEntity, targetEntity]);

  // Initialize form data when entities are provided
  useEffect(() => {
    if (sourceEntity && targetEntity) {
      const suggestedFieldName = generateFieldName(sourceEntity, targetEntity, formData.relation_type);
      const suggestedJunctionTable = generateJunctionTableName(sourceEntity, targetEntity);

      setFormData(prev => ({
        ...prev,
        source_entity_id: sourceEntity.id,
        target_entity_id: targetEntity.id,
        field_name: suggestedFieldName,
        junction_table: suggestedJunctionTable
      }));
    }
  }, [sourceEntity, targetEntity]);

  // Update field name when relation type changes
  useEffect(() => {
    if (sourceEntity && targetEntity) {
      const suggestedFieldName = generateFieldName(sourceEntity, targetEntity, formData.relation_type);
      setFormData(prev => ({
        ...prev,
        field_name: suggestedFieldName
      }));
    }
  }, [formData.relation_type, sourceEntity, targetEntity]);

  const generateFieldName = (source, target, relationType) => {
    if (!source || !target) return '';
    
    const targetNameSnake = toSnakeCase(target.name);
    
    if (relationType === 'belongsTo') {
      return `${targetNameSnake}_id`;
    } else if (relationType === 'hasOne') {
      return `${toSnakeCase(source.name)}_id`;
    } else if (relationType === 'hasMany') {
      return `${toSnakeCase(source.name)}_id`;
    } else if (relationType === 'manyToMany') {
      return ''; // No field name for many-to-many
    }
    return '';
  };

  const generateJunctionTableName = (source, target) => {
    if (!source || !target) return '';
    
    const sourceTable = source.table_name;
    const targetTable = target.table_name;
    
    // Alphabetical order for consistency
    if (sourceTable < targetTable) {
      return `${sourceTable}_${targetTable}`;
    } else {
      return `${targetTable}_${sourceTable}`;
    }
  };

  const toSnakeCase = (str) => {
    return str
      .replace(/([A-Z])/g, '_$1')
      .toLowerCase()
      .replace(/^_/, '');
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
    
    // Clear error for this field
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: null }));
    }
  };

  const validate = () => {
    const newErrors = {};

    if (!formData.source_entity_id) {
      newErrors.source_entity_id = 'Source entity is required';
    }
    if (!formData.target_entity_id) {
      newErrors.target_entity_id = 'Target entity is required';
    }
    if (!formData.relation_type) {
      newErrors.relation_type = 'Relation type is required';
    }
    if (formData.relation_type !== 'manyToMany' && !formData.field_name) {
      newErrors.field_name = 'Field name is required';
    }
    if (formData.relation_type === 'manyToMany' && !formData.junction_table) {
      newErrors.junction_table = 'Junction table name is required for many-to-many relations';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    
    if (!validate()) {
      return;
    }

    onSubmit(formData);
  };

  const handleClose = () => {
    setFormData({
      source_entity_id: '',
      target_entity_id: '',
      field_name: '',
      relation_type: 'belongsTo',
      on_delete: 'RESTRICT',
      on_update: 'CASCADE',
      junction_table: '',
      description: '',
      required: false
    });
    setErrors({});
    onClose();
  };

  if (!isOpen) return null;

  const selectedRelationType = RELATION_TYPES.find(rt => rt.value === formData.relation_type);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-white border-b px-6 py-4 flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Create Relation</h2>
            <p className="text-sm text-gray-500 mt-1">
              Define a relationship between two entities
            </p>
          </div>
          <button
            onClick={handleClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="px-6 py-4 space-y-6">
            {/* Entity Info */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <p className="text-sm font-medium text-blue-900">Source Entity</p>
                  <p className="text-lg font-semibold text-blue-700">{sourceEntity?.name}</p>
                  <p className="text-xs text-blue-600">Table: {sourceEntity?.table_name}</p>
                </div>
                <div className="mx-4">
                  <svg className="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                  </svg>
                </div>
                <div className="flex-1 text-right">
                  <p className="text-sm font-medium text-blue-900">Target Entity</p>
                  <p className="text-lg font-semibold text-blue-700">{targetEntity?.name}</p>
                  <p className="text-xs text-blue-600">Table: {targetEntity?.table_name}</p>
                </div>
              </div>
            </div>

            {/* Relation Type */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Relation Type *
              </label>
              <div className="grid grid-cols-2 gap-3">
                {RELATION_TYPES.map(type => (
                  <label
                    key={type.value}
                    className={`
                      relative flex flex-col p-4 border-2 rounded-lg cursor-pointer transition-all
                      ${formData.relation_type === type.value 
                        ? `border-${type.color}-500 bg-${type.color}-50` 
                        : 'border-gray-200 hover:border-gray-300'
                      }
                    `}
                  >
                    <input
                      type="radio"
                      name="relation_type"
                      value={type.value}
                      checked={formData.relation_type === type.value}
                      onChange={handleChange}
                      className="sr-only"
                    />
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-gray-900">{type.label}</span>
                      <div className={`w-3 h-3 rounded-full bg-${type.color}-500`}></div>
                    </div>
                    <p className="text-xs text-gray-600">{type.description}</p>
                    <p className="text-xs text-gray-500 mt-1 italic">{type.example}</p>
                  </label>
                ))}
              </div>
              {errors.relation_type && (
                <p className="text-red-500 text-sm mt-1">{errors.relation_type}</p>
              )}
            </div>

            {/* Field Name (not for manyToMany) */}
            {formData.relation_type !== 'manyToMany' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Foreign Key Field *
                  <span className="text-xs font-normal text-gray-500 ml-2">
                    (Select existing int/int64 field or use suggested)
                  </span>
                </label>
                <select
                  name="field_name"
                  value={formData.field_name}
                  onChange={handleChange}
                  className={`
                    w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent
                    ${errors.field_name ? 'border-red-500' : 'border-gray-300'}
                  `}
                >
                  <option value="">-- Select FK field --</option>
                  {availableFields.map(field => (
                    <option key={field.value} value={field.value}>
                      {field.label} {field.isExisting ? '(existing)' : '(new)'}
                    </option>
                  ))}
                </select>

                {/* Show description of selected field */}
                {formData.field_name && availableFields.length > 0 && (
                  <div className="mt-2 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                    <p className="text-xs text-blue-800">
                      {availableFields.find(f => f.value === formData.field_name)?.description}
                    </p>
                    <p className="text-xs text-blue-600 mt-1">
                      {formData.relation_type === 'belongsTo' && `Column will be in ${sourceEntity?.name} table (source)`}
                      {(formData.relation_type === 'hasOne' || formData.relation_type === 'hasMany') && 
                        `Column will be in ${targetEntity?.name} table (target)`}
                    </p>
                  </div>
                )}

                {/* Option to add custom field name */}
                <details className="mt-2">
                  <summary className="text-xs text-gray-600 cursor-pointer hover:text-gray-800">
                    Or enter custom field name
                  </summary>
                  <input
                    type="text"
                    placeholder="e.g., author_id, owner_id, parent_id"
                    value={formData.field_name}
                    onChange={handleChange}
                    name="field_name"
                    className="mt-2 w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Custom names allowed. Will be created as BIGINT NOT NULL.
                  </p>
                </details>

                {errors.field_name && (
                  <p className="text-red-500 text-sm mt-1">{errors.field_name}</p>
                )}
              </div>
            )}

            {/* Junction Table (only for manyToMany) */}
            {formData.relation_type === 'manyToMany' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Junction Table Name *
                </label>
                <input
                  type="text"
                  name="junction_table"
                  value={formData.junction_table}
                  onChange={handleChange}
                  placeholder="e.g., posts_tags"
                  className={`
                    w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent
                    ${errors.junction_table ? 'border-red-500' : 'border-gray-300'}
                  `}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Junction table will connect {sourceEntity?.name} and {targetEntity?.name}
                </p>
                {errors.junction_table && (
                  <p className="text-red-500 text-sm mt-1">{errors.junction_table}</p>
                )}
              </div>
            )}

            {/* Required Checkbox */}
            {formData.relation_type === 'belongsTo' && (
              <div className="flex items-center">
                <input
                  type="checkbox"
                  name="required"
                  id="required"
                  checked={formData.required}
                  onChange={handleChange}
                  className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                />
                <label htmlFor="required" className="ml-2 text-sm text-gray-700">
                  Required (NOT NULL constraint)
                </label>
              </div>
            )}

            {/* ON DELETE */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                ON DELETE Behavior
              </label>
              <select
                name="on_delete"
                value={formData.on_delete}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                {ON_DELETE_OPTIONS.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label} - {option.description}
                  </option>
                ))}
              </select>
              <p className="text-xs text-gray-500 mt-1">
                What happens when the referenced record is deleted
              </p>
            </div>

            {/* ON UPDATE */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                ON UPDATE Behavior
              </label>
              <select
                name="on_update"
                value={formData.on_update}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                {ON_UPDATE_OPTIONS.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label} - {option.description}
                  </option>
                ))}
              </select>
              <p className="text-xs text-gray-500 mt-1">
                What happens when the referenced record's ID is updated
              </p>
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Description (Optional)
              </label>
              <textarea
                name="description"
                value={formData.description}
                onChange={handleChange}
                rows={3}
                placeholder="Add a description for this relation..."
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* Footer */}
          <div className="sticky bottom-0 bg-gray-50 border-t px-6 py-4 flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-gray-700 hover:text-gray-900 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
            >
              Create Relation
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
