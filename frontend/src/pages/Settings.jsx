import React from 'react'
import { Settings as SettingsIcon, Database, GitBranch, Server } from 'lucide-react'

export const Settings = () => {
  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="mt-1 text-gray-600">
          Configure your Lambra platform settings
        </p>
      </div>

      <div className="space-y-6">
        {/* Platform Info */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <SettingsIcon className="w-6 h-6 text-primary-600" />
            <h2 className="text-xl font-bold text-gray-900">Platform Information</h2>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between py-3 border-b border-gray-200">
              <span className="text-gray-600">Version</span>
              <span className="font-medium text-gray-900">1.0.0 (Phase 1)</span>
            </div>
            <div className="flex items-center justify-between py-3 border-b border-gray-200">
              <span className="text-gray-600">Environment</span>
              <span className="font-medium text-gray-900">Development</span>
            </div>
            <div className="flex items-center justify-between py-3">
              <span className="text-gray-600">Backend API</span>
              <span className="font-medium text-gray-900">http://localhost:8080</span>
            </div>
          </div>
        </div>

        {/* Database Settings */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Database className="w-6 h-6 text-primary-600" />
            <h2 className="text-xl font-bold text-gray-900">Database Configuration</h2>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between py-3 border-b border-gray-200">
              <span className="text-gray-600">Database Type</span>
              <span className="font-medium text-gray-900">MySQL 8.0</span>
            </div>
            <div className="flex items-center justify-between py-3 border-b border-gray-200">
              <span className="text-gray-600">Host</span>
              <span className="font-medium text-gray-900">localhost:3306</span>
            </div>
            <div className="flex items-center justify-between py-3">
              <span className="text-gray-600">Database Name</span>
              <span className="font-medium text-gray-900">lambra_db</span>
            </div>
          </div>
        </div>

        {/* Git Configuration */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <GitBranch className="w-6 h-6 text-primary-600" />
            <h2 className="text-xl font-bold text-gray-900">Git Configuration</h2>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <p className="text-sm text-yellow-800">
              <strong>Phase 2:</strong> GitLab integration will be configured here.
              You'll be able to set up GitLab URL, access token, and group ID for automatic repository creation.
            </p>
          </div>
        </div>

        {/* Deployment Settings */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Server className="w-6 h-6 text-primary-600" />
            <h2 className="text-xl font-bold text-gray-900">Deployment Settings</h2>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between py-3 border-b border-gray-200">
              <span className="text-gray-600">Deployment Target</span>
              <span className="font-medium text-gray-900">Local Docker</span>
            </div>
            <div className="flex items-center justify-between py-3 border-b border-gray-200">
              <span className="text-gray-600">Workspace Path</span>
              <span className="font-medium text-gray-900">/workspace</span>
            </div>
            <div className="flex items-center justify-between py-3">
              <span className="text-gray-600">Auto Deploy</span>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                Enabled
              </span>
            </div>
          </div>
        </div>

        {/* Phase Information */}
        <div className="card bg-blue-50 border-blue-200">
          <h3 className="text-lg font-bold text-blue-900 mb-3">
            Current Phase: Phase 1 Completed ✅
          </h3>
          <div className="space-y-2 text-sm text-blue-800">
            <p><strong>Completed:</strong></p>
            <ul className="list-disc list-inside space-y-1 ml-2">
              <li>Backend API with CRUD for Projects</li>
              <li>Frontend Dashboard and Service List</li>
              <li>Database schema and migrations</li>
              <li>Docker local development setup</li>
            </ul>
            <p className="mt-3"><strong>Next Phase:</strong></p>
            <ul className="list-disc list-inside space-y-1 ml-2">
              <li>Core Generator Engine</li>
              <li>Template System</li>
              <li>GitLab Integration</li>
              <li>Code Generation Flow</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
