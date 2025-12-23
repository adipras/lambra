import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { Dashboard } from './pages/Dashboard'
import { ServiceList } from './pages/ServiceList'
import { ServiceNew } from './pages/ServiceNew'
import { ServiceDetail } from './pages/ServiceDetail'
import { EndpointDetail } from './pages/EndpointDetail'
import { Settings } from './pages/Settings'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/services" element={<ServiceList />} />
        <Route path="/services/new" element={<ServiceNew />} />
        <Route path="/services/:id" element={<ServiceDetail />} />
        <Route path="/endpoints/:id" element={<EndpointDetail />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Layout>
  )
}

export default App
