import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Search } from 'lucide-react'
import { useProjects } from '../hooks/useProjects'
import { StatusBadge } from '../components/shared/StatusBadge'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'
import { ErrorAlert } from '../components/shared/ErrorAlert'

export const ServiceList = () => {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')

  const { data, isLoading, error } = useProjects({ page, limit: 20 })

  if (isLoading) {
    return <LoadingSpinner size="lg" className="mt-20" />
  }

  const projects = data?.data || []
  const pagination = data?.pagination || {}

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Services</h1>
          <p className="mt-1 text-gray-600">
            Manage your microservices
          </p>
        </div>
        <Link to="/services/new" className="btn btn-primary">
          <Plus className="w-5 h-5 mr-2 inline" />
          New Service
        </Link>
      </div>

      {error && <ErrorAlert message={error.message} />}

      {/* Search */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search services..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input pl-10"
          />
        </div>
      </div>

      {/* Services Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {projects.map((project) => (
          <Link
            key={project.id}
            to={`/services/${project.id}`}
            className="card hover:shadow-md transition-shadow"
          >
            <div className="flex items-start justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                {project.name}
              </h3>
              <StatusBadge status={project.status} />
            </div>

            <p className="text-sm text-gray-600 mb-4 line-clamp-2">
              {project.description || 'No description'}
            </p>

            <div className="flex items-center justify-between text-sm text-gray-500">
              <span>{project.namespace}</span>
              <span>{new Date(project.created_at).toLocaleDateString()}</span>
            </div>
          </Link>
        ))}
      </div>

      {/* Pagination */}
      {pagination.total_pages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="btn btn-secondary disabled:opacity-50"
          >
            Previous
          </button>
          <span className="text-sm text-gray-600">
            Page {page} of {pagination.total_pages}
          </span>
          <button
            onClick={() => setPage(Math.min(pagination.total_pages, page + 1))}
            disabled={page === pagination.total_pages}
            className="btn btn-secondary disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
