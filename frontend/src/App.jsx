import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { Dashboard } from './pages/Dashboard'
import { ServiceList } from './pages/ServiceList'
import { ServiceNew } from './pages/ServiceNew'
import { Settings } from './pages/Settings'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/services" element={<ServiceList />} />
        <Route path="/services/new" element={<ServiceNew />} />
        <Route path="/settings" element={<Settings />} />
        {/* More routes will be added in next phases */}
      </Routes>
    </Layout>
  )
}

export default App
