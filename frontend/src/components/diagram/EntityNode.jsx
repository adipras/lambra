import React, { memo } from 'react'
import { Handle, Position } from 'reactflow'
import { Database, Key, Link2 } from 'lucide-react'

const TYPE_ICONS = {
  string: { icon: 'Aa', color: 'text-blue-600' },
  text: { icon: 'Tt', color: 'text-blue-600' },
  int: { icon: '#', color: 'text-green-600' },
  float: { icon: '.0', color: 'text-green-600' },
  bool: { icon: '?', color: 'text-purple-600' },
  date: { icon: 'D', color: 'text-orange-600' },
  datetime: { icon: 'DT', color: 'text-orange-600' },
  json: { icon: '{}', color: 'text-yellow-600' },
  relation: { icon: '🔗', color: 'text-pink-600' },
}

const EntityNode = ({ data, isConnectable }) => {
  const { entity, onFieldClick, selectedField } = data

  return (
    <div className="bg-white rounded-lg shadow-lg border-2 border-gray-200 min-w-[250px] max-w-[350px]">
      {/* Entity Header */}
      <div className="bg-indigo-600 text-white px-4 py-3 rounded-t-lg flex items-center gap-2">
        <Database className="w-5 h-5" />
        <div className="flex-1">
          <div className="font-bold text-lg">{entity.name}</div>
          <div className="text-xs text-indigo-200">{entity.table_name}</div>
        </div>
      </div>

      {/* Fields List */}
      <div className="divide-y divide-gray-100">
        {/* Primary Key */}
        <div className="px-4 py-2 bg-yellow-50 flex items-center gap-2">
          <Key className="w-4 h-4 text-yellow-600" />
          <span className="text-sm font-semibold text-gray-700">id</span>
          <span className="text-xs text-gray-500 ml-auto">UUID</span>
        </div>

        {/* Entity Fields */}
        {entity.fields?.map((field, index) => {
          const typeInfo = TYPE_ICONS[field.type] || TYPE_ICONS.string
          const isSelected = selectedField?.entityId === entity.id && selectedField?.fieldName === field.name
          const isRelation = field.type === 'relation'

          return (
            <div
              key={index}
              className={`px-4 py-2 flex items-center gap-2 hover:bg-gray-50 cursor-pointer transition-colors relative ${
                isSelected ? 'bg-indigo-50 border-l-4 border-indigo-500' : ''
              } ${isRelation ? 'bg-pink-50' : ''}`}
              onClick={() => onFieldClick(entity.id, field)}
            >
              {/* Connection Handle for Relations */}
              {isRelation && (
                <>
                  <Handle
                    type="source"
                    position={Position.Right}
                    id={`${entity.id}-${field.name}-source`}
                    className="w-3 h-3 bg-pink-500 border-2 border-white"
                    isConnectable={isConnectable}
                  />
                  <Handle
                    type="target"
                    position={Position.Left}
                    id={`${entity.id}-${field.name}-target`}
                    className="w-3 h-3 bg-pink-500 border-2 border-white"
                    isConnectable={isConnectable}
                  />
                </>
              )}

              {/* Field Icon */}
              <span className={`text-xs font-bold ${typeInfo.color} w-6 text-center`}>
                {typeInfo.icon}
              </span>

              {/* Field Name */}
              <span className={`text-sm flex-1 ${field.required ? 'font-semibold' : ''}`}>
                {field.name}
              </span>

              {/* Field Type */}
              <span className="text-xs text-gray-500">{field.type}</span>

              {/* Field Badges */}
              <div className="flex gap-1">
                {field.required && (
                  <span className="px-1 py-0.5 bg-red-100 text-red-600 text-xs rounded">*</span>
                )}
                {field.unique && (
                  <span className="px-1 py-0.5 bg-purple-100 text-purple-600 text-xs rounded">U</span>
                )}
                {isRelation && field.related_entity && (
                  <span className="px-1 py-0.5 bg-pink-100 text-pink-600 text-xs rounded flex items-center gap-1">
                    <Link2 className="w-3 h-3" />
                    {field.related_entity}
                  </span>
                )}
              </div>
            </div>
          )
        })}

        {/* Timestamps */}
        <div className="px-4 py-2 bg-gray-50">
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <span>created_at</span>
            <span className="ml-auto">datetime</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-gray-500 mt-1">
            <span>updated_at</span>
            <span className="ml-auto">datetime</span>
          </div>
        </div>
      </div>

      {/* Default Handles for Non-Relation Connections */}
      <Handle
        type="source"
        position={Position.Right}
        id={`${entity.id}-default-source`}
        className="w-3 h-3 bg-indigo-500 border-2 border-white"
        isConnectable={isConnectable}
      />
      <Handle
        type="target"
        position={Position.Left}
        id={`${entity.id}-default-target`}
        className="w-3 h-3 bg-indigo-500 border-2 border-white"
        isConnectable={isConnectable}
      />
    </div>
  )
}

export default memo(EntityNode)
