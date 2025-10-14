import React from 'react'
import { Link } from 'react-router-dom'
import { Plus, Server, Activity, Clock } from 'lucide-react'
import { useProjects } from '../hooks/useProjects'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'
import { ErrorAlert } from '../components/shared/ErrorAlert'

export const Dashboard = () => {
  const { data, isLoading, error } = useProjects({ page: 1, limit: 5 })

  if (isLoading) {
    return <LoadingSpinner size="lg" className="mt-20" />
  }

  const projects = data?.data || []
  const stats = {
    total: data?.pagination?.total_items || 0,
    active: projects.filter(p => p.status === 'active').length,
    generating: projects.filter(p => p.status === 'generating').length,
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-gray-600">
            Welcome to Lambra Service Generator Platform
          </p>
        </div>
        <Link to="/services/new" className="btn btn-primary">
          <Plus className="w-5 h-5 mr-2 inline" />
          New Service
        </Link>
      </div>

      {error && <ErrorAlert message={error.message} />}

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="card">
          <div className="flex items-center">
            <div className="p-3 bg-primary-100 rounded-lg">
              <Server className="w-6 h-6 text-primary-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm text-gray-600">Total Services</p>
              <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center">
            <div className="p-3 bg-green-100 rounded-lg">
              <Activity className="w-6 h-6 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm text-gray-600">Active</p>
              <p className="text-2xl font-bold text-gray-900">{stats.active}</p>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center">
            <div className="p-3 bg-yellow-100 rounded-lg">
              <Clock className="w-6 h-6 text-yellow-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm text-gray-600">Generating</p>
              <p className="text-2xl font-bold text-gray-900">{stats.generating}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Services */}
      <div className="card">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-bold text-gray-900">Recent Services</h2>
          <Link to="/services" className="text-sm text-primary-600 hover:text-primary-700">
            View all
          </Link>
        </div>

        {projects.length === 0 ? (
          <div className="text-center py-12">
            <Server className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No services yet</p>
            <Link to="/services/new" className="btn btn-primary mt-4">
              Create your first service
            </Link>
          </div>
        ) : (
          <div className="space-y-4">
            {projects.map((project) => (
              <Link
                key={project.id}
                to={`/services/${project.id}`}
                className="block p-4 border border-gray-200 rounded-lg hover:border-primary-300 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900">{project.name}</h3>
                    <p className="text-sm text-gray-600 mt-1">
                      {project.description || 'No description'}
                    </p>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-xs text-gray-500">
                      {new Date(project.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
