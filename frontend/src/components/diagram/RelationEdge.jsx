import React from 'react'
import { BaseEdge, EdgeLabelRenderer, getBezierPath } from 'reactflow'

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

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          ...style,
          stroke: relationStyle.color,
          strokeWidth: 2,
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
        >
          <div
            className="px-2 py-1 rounded text-xs font-medium text-white shadow-lg"
            style={{ backgroundColor: relationStyle.color }}
          >
            {relationStyle.arrow} {relationStyle.label}
          </div>
        </div>
      </EdgeLabelRenderer>
    </>
  )
}

export default RelationEdge
