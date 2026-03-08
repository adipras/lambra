import React, { useCallback, useMemo, useState, useEffect } from 'react'
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  Panel,
} from 'reactflow'
import 'reactflow/dist/style.css'
import EntityNode from './EntityNode'
import RelationEdge from './RelationEdge'
import RelationModal from './RelationModal'
import { LayoutGrid, Save, Download, Info, Plus } from 'lucide-react'
import { createRelation, getRelationsByProject, deleteRelation } from '../../api/relations'
import { useQueryClient } from '@tanstack/react-query'
import { useDiagramStore } from '../../stores/useDiagramStore'

const nodeTypes = {
  entity: EntityNode,
}

const edgeTypes = {
  relation: RelationEdge,
}

// Simple auto-layout without dagre (grid-based)
const getLayoutedElements = (nodes, edges) => {
  const cols = Math.ceil(Math.sqrt(nodes.length))
  const nodeWidth = 350
  const nodeHeight = 450
  const padding = 100

  const layoutedNodes = nodes.map((node, index) => {
    const col = index % cols
    const row = Math.floor(index / cols)
    return {
      ...node,
      position: {
        x: col * (nodeWidth + padding) + 50,
        y: row * (nodeHeight + padding) + 50,
      },
    }
  })

  return { nodes: layoutedNodes, edges }
}

const DatabaseDiagram = ({ entities = [], onSave, projectId }) => {
  // Zustand store
  const openEntityModal = useDiagramStore((state) => state.openEntityModal)

  const [selectedField, setSelectedField] = useState(null)
  const [isRelationModalOpen, setIsRelationModalOpen] = useState(false)
  const [pendingConnection, setPendingConnection] = useState(null)
  const [relations, setRelations] = useState([])
  const queryClient = useQueryClient()

  // Fetch relations for the project (single call, no duplicates)
  useEffect(() => {
    const fetchAllRelations = async () => {
      if (!projectId) return
      
      try {
        const response = await getRelationsByProject(projectId)
        if (response.data && response.data.relations) {
          setRelations(response.data.relations)
        }
      } catch (error) {
        console.error('Failed to fetch relations:', error)
      }
    }
    
    fetchAllRelations()
  }, [projectId])

  // Handle delete relation
  const handleDeleteRelation = useCallback(async (relationId) => {
    try {
      await deleteRelation(relationId)
      
      // Refetch relations
      const response = await getRelationsByProject(projectId)
      if (response.data && response.data.relations) {
        setRelations(response.data.relations)
      }
      
      queryClient.invalidateQueries(['entities'])
      queryClient.invalidateQueries(['project'])
    } catch (error) {
      console.error('Failed to delete relation:', error)
      alert('Failed to delete relation: ' + (error.response?.data?.error || error.message))
    }
  }, [projectId, queryClient])

  // Convert entities to nodes
  const initialNodes = useMemo(() => {
    return entities.map((entity, index) => ({
      id: entity.id,
      type: 'entity',
      position: entity.diagramPosition || { x: 50 + index * 400, y: 50 + (index % 2) * 300 },
      data: {
        entity,
        onFieldClick: (entityId, field) => {
          setSelectedField({ entityId, fieldName: field.name, field })
        },
        selectedField,
      },
    }))
  }, [entities, selectedField])

  // Convert relations to edges
  const initialEdges = useMemo(() => {
    const edges = []
    
    // Convert relations from API to edges
    relations.forEach((relation) => {
      const sourceEntity = entities.find(e => e.id === relation.source_entity_id)
      const targetEntity = entities.find(e => e.id === relation.target_entity_id)
      
      if (sourceEntity && targetEntity) {
        edges.push({
          id: relation.id,
          source: sourceEntity.id,
          target: targetEntity.id,
          type: 'relation',
          data: {
            relationType: relation.relation_type,
            fieldName: relation.source_field_name,
            onDeleteBehavior: relation.on_delete,
            relationId: relation.id,
            onDelete: handleDeleteRelation,
          },
          animated: true,
        })
      }
    })
    
    // Also convert old field-based relations for backward compatibility
    entities.forEach((entity) => {
      entity.fields?.forEach((field) => {
        if (field.type === 'relation' && field.related_entity) {
          const targetEntity = entities.find(e => e.name === field.related_entity)
          if (targetEntity) {
            // Check if this relation already exists in new format
            const existsInNewFormat = edges.some(
              e => e.source === entity.id && e.target === targetEntity.id && e.data?.fieldName === field.name
            )
            
            if (!existsInNewFormat) {
              edges.push({
                id: `${entity.id}-${field.name}-${targetEntity.id}`,
                source: entity.id,
                target: targetEntity.id,
                sourceHandle: `${entity.id}-${field.name}-source`,
                targetHandle: `${targetEntity.id}-default-target`,
                type: 'relation',
                data: {
                  relationType: field.relation_type || 'belongsTo',
                  fieldName: field.name,
                  isLegacy: true, // Mark as legacy
                },
                animated: true,
              })
            }
          }
        }
      })
    })
    
    return edges
  }, [entities, relations, handleDeleteRelation])

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)

  // Update nodes and edges when entities or relations change
  useEffect(() => {
    setNodes(initialNodes)
  }, [initialNodes, setNodes])

  useEffect(() => {
    setEdges(initialEdges)
  }, [initialEdges, setEdges])

  const onConnect = useCallback(
    (params) => {
      // When user draws a connection, open modal instead of directly creating edge
      const sourceEntity = entities.find(e => e.id === params.source)
      const targetEntity = entities.find(e => e.id === params.target)
      
      if (sourceEntity && targetEntity) {
        setPendingConnection({ sourceEntity, targetEntity, params })
        setIsRelationModalOpen(true)
      }
    },
    [entities]
  )

  const handleRelationSubmit = async (formData) => {
    try {
      await createRelation(formData)

      // Refetch relations
      const refetchResponse = await getRelationsByProject(projectId)
      if (refetchResponse.data && refetchResponse.data.relations) {
        setRelations(refetchResponse.data.relations)
      }
      
      // Invalidate queries to refresh entity list if needed
      queryClient.invalidateQueries(['entities'])
      queryClient.invalidateQueries(['project'])
      
      // Show success feedback
      alert('Relation created successfully!')
      
      setIsRelationModalOpen(false)
      setPendingConnection(null)
    } catch (error) {
      console.error('Failed to create relation:', error)
      alert('Failed to create relation: ' + (error.response?.data?.error || error.message))
    }
  }

  const handleRelationModalClose = () => {
    setIsRelationModalOpen(false)
    setPendingConnection(null)
  }

  const onLayout = useCallback(() => {
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(nodes, edges)
    setNodes([...layoutedNodes])
    setEdges([...layoutedEdges])
  }, [nodes, edges, setNodes, setEdges])

  const onSavePositions = useCallback(() => {
    const positions = {}
    nodes.forEach((node) => {
      positions[node.id] = node.position
    })
    if (onSave) {
      onSave(positions)
    }
  }, [nodes, onSave])

  const onExportImage = useCallback(() => {
    // TODO: Implement export as PNG/SVG
    console.log('Export diagram as image')
  }, [])

  return (
    <div className="h-full w-full relative">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        minZoom={0.1}
        maxZoom={2}
        defaultEdgeOptions={{
          animated: true,
          style: { stroke: '#94a3b8', strokeWidth: 2 },
        }}
      >
        <Background color="#e2e8f0" gap={16} />
        <Controls />
        <MiniMap
          nodeColor={(node) => {
            return '#6366f1'
          }}
          maskColor="rgba(0, 0, 0, 0.1)"
        />

        {/* Custom Controls */}
        <Panel position="top-right" className="space-y-2">
          <div className="bg-white rounded-lg shadow-lg p-2 space-y-2">
            <button
              onClick={onLayout}
              className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors w-full"
              title="Auto-layout diagram"
            >
              <LayoutGrid className="w-4 h-4" />
              Auto Layout
            </button>

            <button
              onClick={onSavePositions}
              className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors w-full"
              title="Save positions"
            >
              <Save className="w-4 h-4" />
              Save Layout
            </button>

            <button
              onClick={onExportImage}
              className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors w-full"
              title="Export as image"
            >
              <Download className="w-4 h-4" />
              Export Image
            </button>
          </div>

          {/* Add Entity Button - FAB Style */}
          <button
            onClick={openEntityModal}
            className="flex items-center gap-2 px-4 py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg shadow-lg text-sm font-medium transition-colors w-full"
            title="Create new entity"
          >
            <Plus className="w-4 h-4" />
            Add Entity
          </button>
        </Panel>

        {/* Info Panel */}
        <Panel position="bottom-left" className="bg-white rounded-lg shadow-lg p-4">
          <div className="flex items-start gap-2">
            <Info className="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-gray-600 space-y-1">
              <div><strong>Drag</strong> entities to reposition</div>
              <div><strong>Scroll</strong> to zoom in/out</div>
              <div><strong>Drag from entity to entity</strong> to create relation</div>
              <div><strong>Hover over relation</strong> to view details & delete</div>
              <div className="text-xs text-gray-500 mt-2 space-y-0.5">
                <div>🔗 Relations: Pink=belongsTo, Purple=hasOne, Blue=hasMany, Orange=manyToMany</div>
                <div>Dashed lines = Legacy field-based relations</div>
              </div>
            </div>
          </div>
        </Panel>

        {/* Selected Field Info */}
        {selectedField && (
          <Panel position="top-left" className="bg-white rounded-lg shadow-lg p-4 max-w-xs">
            <div className="space-y-2">
              <div className="font-semibold text-gray-900 border-b pb-2">
                Selected Field
              </div>
              <div className="text-sm space-y-1">
                <div className="flex justify-between">
                  <span className="text-gray-500">Name:</span>
                  <span className="font-medium">{selectedField.fieldName}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Type:</span>
                  <span className="font-medium">{selectedField.field.type}</span>
                </div>
                {selectedField.field.relation_type && (
                  <div className="flex justify-between">
                    <span className="text-gray-500">Relation:</span>
                    <span className="font-medium">{selectedField.field.relation_type}</span>
                  </div>
                )}
                {selectedField.field.related_entity && (
                  <div className="flex justify-between">
                    <span className="text-gray-500">Related To:</span>
                    <span className="font-medium">{selectedField.field.related_entity}</span>
                  </div>
                )}
              </div>
              <button
                onClick={() => setSelectedField(null)}
                className="text-xs text-gray-500 hover:text-gray-700 mt-2"
              >
                Clear selection
              </button>
            </div>
          </Panel>
        )}
      </ReactFlow>

      {/* Relation Modal */}
      {isRelationModalOpen && pendingConnection && (
        <RelationModal
          isOpen={isRelationModalOpen}
          onClose={handleRelationModalClose}
          onSubmit={handleRelationSubmit}
          sourceEntity={pendingConnection.sourceEntity}
          targetEntity={pendingConnection.targetEntity}
        />
      )}
    </div>
  )
}

export default DatabaseDiagram
