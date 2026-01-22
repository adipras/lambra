import React, { useState } from 'react'
import { BaseEdge, EdgeLabelRenderer, getBezierPath } from 'reactflow'
import { Trash2, Info } from 'lucide-react'

const RELATION_STYLES = {
  belongsTo: { color: '#ec4899', label: 'belongs to', arrow: '→' },
  hasOne: { color: '#8b5cf6', label: 'has one', arrow: '1:1' },
  hasMany: { color: '#3b82f6', label: 'has many', arrow: '1:N' },
  manyToMany: { color: '#f59e0b', label: 'many to many', arrow: 'N:M' },
}

const RelationEdge = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  data,
  markerEnd,
}) => {
  const [showActions, setShowActions] = useState(false)
  const [showTooltip, setShowTooltip] = useState(false)

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  })

  const relationType = data?.relationType || 'belongsTo'
  const relationStyle = RELATION_STYLES[relationType] || RELATION_STYLES.belongsTo
  const isLegacy = data?.isLegacy || false

  const handleDeleteClick = (e) => {
    e.stopPropagation()
    if (data?.onDelete && data?.relationId) {
      if (confirm('Are you sure you want to delete this relation?')) {
        data.onDelete(data.relationId)
      }
    }
  }

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          ...style,
          stroke: relationStyle.color,
          strokeWidth: showActions ? 3 : 2,
          opacity: isLegacy ? 0.5 : 1,
          strokeDasharray: isLegacy ? '5,5' : 'none',
        }}
      />
      <EdgeLabelRenderer>
        <div
          style={{
            position: 'absolute',
            transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
            pointerEvents: 'all',
          }}
          className="nodrag nopan"
          onMouseEnter={() => {
            setShowActions(true)
            setShowTooltip(true)
          }}
          onMouseLeave={() => {
            setShowActions(false)
            setShowTooltip(false)
          }}
        >
          {/* Main Label */}
          <div className="relative">
            <div
              className={`
                px-2 py-1 rounded text-xs font-medium text-white shadow-lg
                transition-all duration-200 cursor-pointer
                ${showActions ? 'scale-110' : 'scale-100'}
              `}
              style={{ backgroundColor: relationStyle.color }}
            >
              {relationStyle.arrow} {relationStyle.label}
              {data?.fieldName && ` (${data.fieldName})`}
            </div>

            {/* Action Buttons - Show on Hover */}
            {showActions && !isLegacy && (
              <div className="absolute top-full left-1/2 transform -translate-x-1/2 mt-1 flex gap-1 bg-white rounded shadow-lg p-1">
                <button
                  onClick={handleDeleteClick}
                  className="p-1 hover:bg-red-50 rounded text-red-600 transition-colors"
                  title="Delete relation"
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              </div>
            )}

            {/* Tooltip - Show on Hover */}
            {showTooltip && (
              <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 w-64 bg-gray-900 text-white text-xs rounded-lg p-3 shadow-xl z-50">
                <div className="space-y-1">
                  <div className="font-semibold border-b border-gray-700 pb-1">
                    {relationStyle.label.toUpperCase()}
                  </div>
                  {data?.fieldName && (
                    <div className="flex justify-between">
                      <span className="text-gray-400">Field:</span>
                      <span className="font-mono">{data.fieldName}</span>
                    </div>
                  )}
                  {data?.onDeleteBehavior && (
                    <div className="flex justify-between">
                      <span className="text-gray-400">ON DELETE:</span>
                      <span className={`
                        font-medium
                        ${data.onDeleteBehavior === 'CASCADE' ? 'text-red-400' : 
                          data.onDeleteBehavior === 'SET NULL' ? 'text-yellow-400' : 
                          'text-gray-300'}
                      `}>
                        {data.onDeleteBehavior}
                      </span>
                    </div>
                  )}
                  {isLegacy && (
                    <div className="mt-2 text-yellow-400 text-xs border-t border-gray-700 pt-1">
                      ⚠️ Legacy relation (field-based)
                    </div>
                  )}
                  {!isLegacy && (
                    <div className="mt-2 text-xs text-gray-500">
                      Click trash icon to delete
                    </div>
                  )}
                </div>
                {/* Tooltip Arrow */}
                <div className="absolute top-full left-1/2 transform -translate-x-1/2 -mt-1">
                  <div className="border-8 border-transparent border-t-gray-900"></div>
                </div>
              </div>
            )}
          </div>
        </div>
      </EdgeLabelRenderer>
    </>
  )
}

export default RelationEdge
