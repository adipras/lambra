import { create } from 'zustand'

export const useDiagramStore = create((set) => ({
  // ========================
  // UI State
  // ========================
  showEntityModal: false,
  showEndpointModal: false,
  selectedEntity: null,
  viewMode: 'list', // 'list' or 'diagram'
  isExpanded: true, // For accordion/expansion state

  // Diagram-specific state
  diagramPositions: {}, // Store entity positions: { entityId: { x, y } }
  selectedNodes: [], // Multi-select for batch operations
  zoomLevel: 1,
  viewport: { x: 0, y: 0, zoom: 1 },

  // ========================
  // Basic Actions
  // ========================
  setShowEntityModal: (show) => set({ showEntityModal: show }),
  setShowEndpointModal: (show) => set({ showEndpointModal: show }),
  setSelectedEntity: (entity) => set({ selectedEntity: entity }),
  setViewMode: (mode) => set({ viewMode: mode }),
  setExpanded: (isExpanded) => set({ isExpanded }),

  // ========================
  // Entity Modal Actions
  // ========================
  openEntityModal: () => set({ showEntityModal: true }),
  closeEntityModal: () => set({ showEntityModal: false, selectedEntity: null }),

  // ========================
  // Endpoint Modal Actions
  // ========================
  openEndpointModal: (entity) => set({
    showEndpointModal: true,
    selectedEntity: entity
  }),
  closeEndpointModal: () => set({
    showEndpointModal: false,
    selectedEntity: null
  }),

  // ========================
  // View Mode Actions
  // ========================
  toggleViewMode: () => set((state) => ({
    viewMode: state.viewMode === 'list' ? 'diagram' : 'list'
  })),
  setListViewMode: () => set({ viewMode: 'list' }),
  setDiagramViewMode: () => set({ viewMode: 'diagram' }),

  // ========================
  // Diagram State Actions
  // ========================
  setDiagramPositions: (positions) => set({ diagramPositions: positions }),
  updateNodePosition: (entityId, position) => set((state) => ({
    diagramPositions: {
      ...state.diagramPositions,
      [entityId]: position
    }
  })),
  setSelectedNodes: (nodes) => set({ selectedNodes: nodes }),
  setViewport: (viewport) => set({ viewport }),
  setZoomLevel: (zoom) => set((state) => ({
    zoomLevel: zoom,
    viewport: { ...state.viewport, zoom }
  })),

  // ========================
  // Convenience Actions
  // ========================
  // Open entity modal and switch to diagram view
  openEntityModalInDiagram: () => set({ viewMode: 'diagram', showEntityModal: true }),

  // Open endpoint modal for specific entity
  openEndpointModalForEntity: (entity) => set({
    showEndpointModal: true,
    selectedEntity: entity
  }),

  // Reset all diagram state
  resetDiagramState: () => set({
    diagramPositions: {},
    selectedNodes: [],
    zoomLevel: 1,
    viewport: { x: 0, y: 0, zoom: 1 }
  }),
}))
