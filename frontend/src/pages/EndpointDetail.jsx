import { useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Play, Copy, Check, AlertCircle, Clock, CheckCircle2, XCircle, Loader2, Shield, ShieldOff, Edit, Save, X } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { endpointsApi } from '../api/endpoints'
import { LoadingSpinner } from '../components/shared/LoadingSpinner'
import { ErrorAlert } from '../components/shared/ErrorAlert'
import { SchemaViewer, JsonEditor } from '../components/code/SchemaViewer'
import { CodeEditor } from '../components/code/CodeEditor'

export const EndpointDetail = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [isEditing, setIsEditing] = useState(false)
  const [editData, setEditData] = useState(null)
  const [testResult, setTestResult] = useState(null)
  const [isTestLoading, setIsTestLoading] = useState(false)

  // Request body state for testing
  const [requestBody, setRequestBody] = useState('{}')
  const [requestHeaders, setRequestHeaders] = useState('{}')

  // Fetch endpoint
  const { data: endpointData, isLoading, error } = useQuery({
    queryKey: ['endpoint', id],
    queryFn: () => endpointsApi.getById(id),
  })

  // Update endpoint mutation
  const updateMutation = useMutation({
    mutationFn: (data) => endpointsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries(['endpoint', id])
      setIsEditing(false)
      setEditData(null)
    },
  })

  const endpoint = endpointData?.data

  // Initialize edit data when entering edit mode
  const handleStartEdit = () => {
    setEditData({
      name: endpoint?.name || '',
      path: endpoint?.path || '',
      method: endpoint?.method || 'GET',
      description: endpoint?.description || '',
      request_schema: JSON.stringify(endpoint?.request_schema || {}, null, 2),
      response_schema: JSON.stringify(endpoint?.response_schema || {}, null, 2),
      require_auth: endpoint?.require_auth || false,
    })
    setIsEditing(true)
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    setEditData(null)
  }

  const handleSaveEdit = () => {
    try {
      const data = {
        ...editData,
        request_schema: editData.request_schema ? JSON.parse(editData.request_schema) : {},
        response_schema: editData.response_schema ? JSON.parse(editData.response_schema) : {},
      }
      updateMutation.mutate(data)
    } catch (err) {
      alert('Invalid JSON in schema fields')
    }
  }

  // Test endpoint via backend API
  const handleTestEndpoint = async () => {
    setIsTestLoading(true)
    setTestResult(null)

    try {
      // Parse request body and headers
      let body = {}
      let headers = {}

      try {
        body = requestBody ? JSON.parse(requestBody) : {}
      } catch {
        setTestResult({
          success: false,
          error: 'Invalid JSON in request body',
          statusCode: 0,
          responseTime: 0,
        })
        setIsTestLoading(false)
        return
      }

      try {
        headers = requestHeaders ? JSON.parse(requestHeaders) : {}
      } catch {
        setTestResult({
          success: false,
          error: 'Invalid JSON in headers',
          statusCode: 0,
          responseTime: 0,
        })
        setIsTestLoading(false)
        return
      }

      // Call the test API
      const response = await endpointsApi.test(id, { headers, body })
      const data = response.data?.data

      if (data?.error) {
        setTestResult({
          success: false,
          error: data.error,
          statusCode: data.status_code || 0,
          responseTime: data.response_time || 0,
        })
      } else {
        setTestResult({
          success: true,
          statusCode: data?.status_code || 200,
          responseTime: data?.response_time || 0,
          headers: data?.headers || {},
          body: data?.body || {},
        })
      }
    } catch (err) {
      const errorMessage = err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to test endpoint'
      setTestResult({
        success: false,
        error: errorMessage,
        statusCode: err.response?.status || 0,
        responseTime: 0,
      })
    } finally {
      setIsTestLoading(false)
    }
  }

  const getMethodBadgeColor = (method) => {
    const colors = {
      GET: 'bg-blue-500 text-white',
      POST: 'bg-green-500 text-white',
      PUT: 'bg-yellow-500 text-white',
      DELETE: 'bg-red-500 text-white',
      PATCH: 'bg-purple-500 text-white',
    }
    return colors[method] || 'bg-gray-500 text-white'
  }

  const getStatusColor = (statusCode) => {
    if (statusCode >= 200 && statusCode < 300) return 'text-green-600'
    if (statusCode >= 300 && statusCode < 400) return 'text-blue-600'
    if (statusCode >= 400 && statusCode < 500) return 'text-yellow-600'
    if (statusCode >= 500) return 'text-red-600'
    return 'text-gray-600'
  }

  if (isLoading) {
    return <LoadingSpinner size="lg" className="mt-20" />
  }

  if (error) {
    return <ErrorAlert message="Failed to load endpoint details" />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate(-1)}
            className="btn btn-secondary"
          >
            <ArrowLeft className="w-4 h-4" />
          </button>
          <div>
            <div className="flex items-center gap-3">
              <span className={`inline-flex items-center px-3 py-1 rounded-lg text-sm font-bold ${getMethodBadgeColor(endpoint?.method)}`}>
                {endpoint?.method}
              </span>
              <h1 className="text-2xl font-bold text-gray-900">{endpoint?.name}</h1>
              {endpoint?.require_auth ? (
                <Shield className="w-5 h-5 text-green-600" title="Authentication Required" />
              ) : (
                <ShieldOff className="w-5 h-5 text-gray-400" title="No Authentication" />
              )}
            </div>
            <p className="text-gray-600 mt-1 font-mono">{endpoint?.path}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isEditing ? (
            <>
              <button
                onClick={handleCancelEdit}
                className="btn btn-secondary"
              >
                <X className="w-4 h-4 mr-1" />
                Cancel
              </button>
              <button
                onClick={handleSaveEdit}
                disabled={updateMutation.isPending}
                className="btn btn-primary"
              >
                {updateMutation.isPending ? (
                  <Loader2 className="w-4 h-4 mr-1 animate-spin" />
                ) : (
                  <Save className="w-4 h-4 mr-1" />
                )}
                Save
              </button>
            </>
          ) : (
            <button
              onClick={handleStartEdit}
              className="btn btn-secondary"
            >
              <Edit className="w-4 h-4 mr-1" />
              Edit
            </button>
          )}
        </div>
      </div>

      {/* Description */}
      {(endpoint?.description || isEditing) && (
        <div className="card">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Description</h3>
          {isEditing ? (
            <textarea
              value={editData?.description || ''}
              onChange={(e) => setEditData({ ...editData, description: e.target.value })}
              className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              rows={2}
              placeholder="Endpoint description..."
            />
          ) : (
            <p className="text-gray-600">{endpoint?.description || 'No description'}</p>
          )}
        </div>
      )}

      <div className="grid grid-cols-2 gap-6">
        {/* Request Schema */}
        <div className="card">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Request Schema</h3>
          {isEditing ? (
            <JsonEditor
              value={editData?.request_schema || '{}'}
              onChange={(val) => setEditData({ ...editData, request_schema: val })}
              placeholder='{\n  "property": "type"\n}'
            />
          ) : (
            <SchemaViewer
              schema={endpoint?.request_schema}
              title="Request Body"
              emptyMessage="No request body schema defined"
            />
          )}
        </div>

        {/* Response Schema */}
        <div className="card">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Response Schema</h3>
          {isEditing ? (
            <JsonEditor
              value={editData?.response_schema || '{}'}
              onChange={(val) => setEditData({ ...editData, response_schema: val })}
              placeholder='{\n  "property": "type"\n}'
            />
          ) : (
            <SchemaViewer
              schema={endpoint?.response_schema}
              title="Response Body"
              emptyMessage="No response schema defined"
            />
          )}
        </div>
      </div>

      {/* Testing Section */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Test Endpoint</h3>
          <button
            onClick={handleTestEndpoint}
            disabled={isTestLoading}
            className="btn btn-primary flex items-center gap-2"
          >
            {isTestLoading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Testing...
              </>
            ) : (
              <>
                <Play className="w-4 h-4" />
                Send Request
              </>
            )}
          </button>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* Request Body Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Request Body
            </label>
            <textarea
              value={requestBody}
              onChange={(e) => setRequestBody(e.target.value)}
              placeholder='{\n  "key": "value"\n}'
              className="w-full h-40 p-3 font-mono text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          {/* Request Headers Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Headers
            </label>
            <textarea
              value={requestHeaders}
              onChange={(e) => setRequestHeaders(e.target.value)}
              placeholder='{\n  "Content-Type": "application/json"\n}'
              className="w-full h-40 p-3 font-mono text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
        </div>

        {/* Test Result */}
        {testResult && (
          <div className="mt-6 border-t pt-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-4">
                <h4 className="text-md font-medium text-gray-900">Response</h4>
                {testResult.success ? (
                  <span className="flex items-center gap-1 text-green-600">
                    <CheckCircle2 className="w-4 h-4" />
                    Success
                  </span>
                ) : (
                  <span className="flex items-center gap-1 text-red-600">
                    <XCircle className="w-4 h-4" />
                    Failed
                  </span>
                )}
              </div>
              <div className="flex items-center gap-4 text-sm">
                <span className={`font-mono font-bold ${getStatusColor(testResult.statusCode)}`}>
                  {testResult.statusCode || 'N/A'}
                </span>
                <span className="flex items-center gap-1 text-gray-500">
                  <Clock className="w-4 h-4" />
                  {testResult.responseTime}ms
                </span>
              </div>
            </div>

            {testResult.error ? (
              <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                <div className="flex items-center gap-2 text-red-700">
                  <AlertCircle className="w-5 h-5" />
                  <span className="font-medium">Error</span>
                </div>
                <p className="mt-1 text-red-600">{testResult.error}</p>
              </div>
            ) : (
              <div className="space-y-4">
                {/* Response Headers */}
                {testResult.headers && Object.keys(testResult.headers).length > 0 && (
                  <div>
                    <h5 className="text-sm font-medium text-gray-700 mb-2">Response Headers</h5>
                    <div className="bg-gray-50 rounded-lg p-3 font-mono text-sm">
                      {Object.entries(testResult.headers).map(([key, value]) => (
                        <div key={key} className="flex gap-2">
                          <span className="text-gray-500">{key}:</span>
                          <span className="text-gray-900">{value}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Response Body */}
                <div>
                  <h5 className="text-sm font-medium text-gray-700 mb-2">Response Body</h5>
                  <CodeEditor
                    code={JSON.stringify(testResult.body, null, 2)}
                    language="json"
                    showLineNumbers={true}
                    maxHeight={300}
                  />
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Info Card */}
      <div className="card bg-gray-50">
        <h3 className="text-sm font-medium text-gray-700 mb-3">Endpoint Info</h3>
        <div className="grid grid-cols-4 gap-4 text-sm">
          <div>
            <p className="text-gray-500">ID</p>
            <p className="font-mono text-gray-900 truncate" title={endpoint?.id}>
              {endpoint?.id?.slice(0, 8)}...
            </p>
          </div>
          <div>
            <p className="text-gray-500">Authentication</p>
            <p className="text-gray-900">
              {endpoint?.require_auth ? 'Required' : 'Not Required'}
            </p>
          </div>
          <div>
            <p className="text-gray-500">Created</p>
            <p className="text-gray-900">
              {new Date(endpoint?.created_at).toLocaleDateString()}
            </p>
          </div>
          <div>
            <p className="text-gray-500">Updated</p>
            <p className="text-gray-900">
              {new Date(endpoint?.updated_at).toLocaleDateString()}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default EndpointDetail
