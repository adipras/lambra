import React, { useState, useEffect } from 'react'
import { X, Play, Clock, CheckCircle, XCircle, Copy, Check, Loader2, AlertTriangle } from 'lucide-react'
import { generateExampleFromSchema } from './ExampleBody'
import { endpointsApi } from '../../api/endpoints'

const METHOD_COLORS = {
  GET: 'bg-blue-500',
  POST: 'bg-green-500',
  PUT: 'bg-yellow-500',
  DELETE: 'bg-red-500',
  PATCH: 'bg-purple-500',
}

// Extract query parameters from request schema
// Checks for x-parameter-style: query in schema
const extractQueryParams = (schema) => {
  if (!schema || schema['x-parameter-style'] !== 'query') {
    return []
  }

  const properties = schema.properties || {}
  const required = schema.required || []

  return Object.entries(properties).map(([name, prop]) => ({
    name,
    type: prop.type || 'string',
    description: prop.description || '',
    example: prop.example || '',
    required: required.includes(name)
  }))
}

// Extract body fields from schema (excluding query params)
const extractBodySchema = (schema) => {
  if (!schema || schema['x-parameter-style'] === 'query') {
    return null
  }

  // For merged schemas (query + body), separate them
  // If schema has x-parameter-style but also has body fields
  const properties = schema.properties || {}
  const queryParams = ['id'] // Known query param names

  const bodyProperties = {}
  Object.entries(properties).forEach(([key, value]) => {
    if (!queryParams.includes(key) || !value.description?.includes('query')) {
      bodyProperties[key] = value
    }
  })

  if (Object.keys(bodyProperties).length === 0) {
    return null
  }

  return {
    ...schema,
    properties: bodyProperties,
    required: (schema.required || []).filter(r => !queryParams.includes(r))
  }
}

export const TestEndpointModal = ({ endpoint, serviceUrl, onClose }) => {
  const [requestBody, setRequestBody] = useState('')
  const [queryParams, setQueryParams] = useState({})
  const [response, setResponse] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState(null)
  const [copied, setCopied] = useState(false)

  // Extract query params from request schema
  const queryParamDefs = endpoint?.request_schema ? extractQueryParams(endpoint.request_schema) : []
  const hasQueryParams = queryParamDefs.length > 0

  // Get body schema (excluding query params)
  const bodySchema = endpoint?.request_schema ? extractBodySchema(endpoint.request_schema) : null

  // Initialize query params with example values
  useEffect(() => {
    if (queryParamDefs.length > 0) {
      const initialParams = {}
      queryParamDefs.forEach(param => {
        initialParams[param.name] = param.example || ''
      })
      setQueryParams(initialParams)
    }
  }, [endpoint?.request_schema])

  // Generate example request body from schema (excluding query params)
  useEffect(() => {
    if (bodySchema && Object.keys(bodySchema.properties || {}).length > 0) {
      const example = generateExampleFromSchema(bodySchema)
      setRequestBody(JSON.stringify(example, null, 2))
    } else if (endpoint?.request_schema && !hasQueryParams && Object.keys(endpoint.request_schema).length > 0) {
      const example = generateExampleFromSchema(endpoint.request_schema)
      setRequestBody(JSON.stringify(example, null, 2))
    } else {
      setRequestBody('')
    }
  }, [endpoint, hasQueryParams])

  // Update a single query param
  const updateQueryParam = (key, value) => {
    setQueryParams(prev => ({ ...prev, [key]: value }))
  }

  // Build preview URL with query params
  const buildPreviewUrl = () => {
    let url = endpoint?.path || ''
    if (hasQueryParams) {
      const params = new URLSearchParams()
      Object.entries(queryParams).forEach(([key, value]) => {
        if (value) {
          params.append(key, value)
        }
      })
      const paramString = params.toString()
      if (paramString) {
        url += '?' + paramString
      }
    }
    return url
  }

  const previewUrl = buildPreviewUrl()

  const handleTest = async () => {
    setIsLoading(true)
    setError(null)
    setResponse(null)

    // Validate required query params are filled
    if (hasQueryParams) {
      const missingParams = queryParamDefs
        .filter(param => param.required && !queryParams[param.name]?.trim())
        .map(param => param.name)
      if (missingParams.length > 0) {
        setError(`Missing required query parameters: ${missingParams.join(', ')}`)
        setIsLoading(false)
        return
      }
    }

    try {
      // Parse request body if present
      let body = null
      if (requestBody.trim() && ['POST', 'PUT', 'PATCH'].includes(endpoint.method)) {
        try {
          body = JSON.parse(requestBody)
        } catch (e) {
          setError('Invalid JSON in request body')
          setIsLoading(false)
          return
        }
      }

      // Call the test endpoint API with query params
      const result = await endpointsApi.test(endpoint.id, {
        body: body,
        headers: { 'Content-Type': 'application/json' },
        params: hasQueryParams ? queryParams : undefined
      })

      console.log('Test result:', result) // Debug log

      // The axios interceptor returns response.data, so result = { success, message, data }
      if (result?.data) {
        setResponse(result.data)
      } else if (result?.status_code !== undefined) {
        // In case result is already the data object
        setResponse(result)
      } else {
        setError('Unexpected response format')
      }
    } catch (err) {
      console.error('Test error:', err) // Debug log
      setError(err.response?.data?.message || err.message || 'Failed to test endpoint')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCopyResponse = () => {
    const bodyStr = formatResponseBody(response?.body)
    if (bodyStr) {
      navigator.clipboard.writeText(bodyStr)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const formatResponseBody = (body) => {
    if (body === null || body === undefined) return 'null'
    if (typeof body === 'string') {
      // Try to parse and prettify if it's a JSON string
      try {
        const parsed = JSON.parse(body)
        return JSON.stringify(parsed, null, 2)
      } catch {
        return body
      }
    }
    // It's already an object
    return JSON.stringify(body, null, 2)
  }

  const needsBody = ['POST', 'PUT', 'PATCH'].includes(endpoint?.method)
  const hasResponseBody = response?.body !== undefined && response?.body !== null

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center gap-4">
            <span className={`px-3 py-1.5 text-sm font-bold text-white rounded-lg ${METHOD_COLORS[endpoint?.method]}`}>
              {endpoint?.method}
            </span>
            <div>
              <h2 className="text-lg font-bold text-gray-900">{endpoint?.name}</h2>
              <p className="text-sm font-mono text-gray-500">
                {serviceUrl}
                <span className={hasQueryParams ? 'text-blue-600' : ''}>{previewUrl}</span>
              </p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-gray-200 rounded-lg transition-colors">
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-hidden flex">
          {/* Request Panel */}
          <div className="w-1/2 border-r border-gray-200 flex flex-col">
            <div className="px-4 py-3 bg-green-50 border-b border-green-200">
              <h3 className="font-semibold text-green-800">Request</h3>
              {(hasQueryParams || needsBody) && (
                <p className="text-xs text-green-600 mt-0.5">
                  {hasQueryParams && needsBody
                    ? 'Configure query params and JSON body below'
                    : hasQueryParams
                      ? 'Configure query parameters below'
                      : 'Edit the JSON body below'}
                </p>
              )}
            </div>

            <div className="flex-1 overflow-auto p-4 space-y-4">
              {/* Query Parameters Section */}
              {hasQueryParams && (
                <div className="space-y-3">
                  <h4 className="text-sm font-medium text-gray-700">Query Parameters</h4>
                  {queryParamDefs.map(param => (
                    <div key={param.name} className="space-y-1">
                      <label className="flex items-center gap-2 text-sm text-gray-600">
                        <span className="font-medium">{param.name}</span>
                        {param.required && (
                          <span className="text-red-500 text-xs">*required</span>
                        )}
                      </label>
                      <input
                        type="text"
                        value={queryParams[param.name] || ''}
                        onChange={(e) => updateQueryParam(param.name, e.target.value)}
                        placeholder={param.example || `Enter ${param.name}...`}
                        className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
                      />
                      {param.description && (
                        <p className="text-xs text-gray-500">{param.description}</p>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {/* Request Body Section */}
              {needsBody ? (
                <div className="flex-1">
                  {hasQueryParams && (
                    <h4 className="text-sm font-medium text-gray-700 mb-2">Request Body</h4>
                  )}
                  <textarea
                    value={requestBody}
                    onChange={(e) => setRequestBody(e.target.value)}
                    className="w-full h-full min-h-[200px] font-mono text-sm p-3 bg-gray-900 text-gray-100 rounded-lg border-0 resize-none focus:ring-2 focus:ring-green-500"
                    placeholder="Enter request body JSON..."
                    spellCheck={false}
                  />
                </div>
              ) : !hasQueryParams ? (
                <div className="flex items-center justify-center h-full text-gray-400">
                  <div className="text-center">
                    <p className="text-sm">No request parameters required for {endpoint?.method}</p>
                  </div>
                </div>
              ) : null}
            </div>

            {/* Send Button */}
            <div className="p-4 border-t border-gray-200 bg-gray-50">
              <button
                onClick={handleTest}
                disabled={isLoading}
                className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-semibold rounded-xl transition-colors"
              >
                {isLoading ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" />
                    Sending...
                  </>
                ) : (
                  <>
                    <Play className="w-5 h-5" />
                    Send Request
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Response Panel */}
          <div className="w-1/2 flex flex-col bg-gray-50">
            <div className="px-4 py-3 bg-blue-50 border-b border-blue-200 flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-blue-800">Response</h3>
                {response && (
                  <div className="flex items-center gap-3 mt-1">
                    <span className={`flex items-center gap-1 text-xs font-medium ${
                      response.status_code >= 200 && response.status_code < 300
                        ? 'text-green-600'
                        : 'text-red-600'
                    }`}>
                      {response.status_code >= 200 && response.status_code < 300 ? (
                        <CheckCircle className="w-3.5 h-3.5" />
                      ) : (
                        <XCircle className="w-3.5 h-3.5" />
                      )}
                      {response.status_code}
                    </span>
                    <span className="flex items-center gap-1 text-xs text-gray-500">
                      <Clock className="w-3.5 h-3.5" />
                      {response.response_time}ms
                    </span>
                  </div>
                )}
              </div>
              {hasResponseBody && (
                <button
                  onClick={handleCopyResponse}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-blue-600 hover:bg-blue-100 rounded transition-colors"
                >
                  {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                  {copied ? 'Copied!' : 'Copy'}
                </button>
              )}
            </div>

            <div className="flex-1 overflow-auto p-4">
              {error ? (
                <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg">
                  <XCircle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="font-medium text-red-800">Error</p>
                    <p className="text-sm text-red-600 mt-1">{error}</p>
                  </div>
                </div>
              ) : response ? (
                <div className="h-full flex flex-col gap-3">
                  {/* Response Error */}
                  {response.error && (
                    <div className="flex items-start gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                      <AlertTriangle className="w-4 h-4 text-yellow-600 flex-shrink-0 mt-0.5" />
                      <p className="text-sm text-yellow-800">{response.error}</p>
                    </div>
                  )}

                  {/* Response Body */}
                  <div className="flex-1">
                    <pre className="w-full h-full min-h-[200px] font-mono text-sm p-3 bg-gray-900 text-gray-100 rounded-lg overflow-auto whitespace-pre-wrap">
                      <JsonHighlight json={formatResponseBody(response.body)} />
                    </pre>
                  </div>
                </div>
              ) : (
                <div className="flex items-center justify-center h-full text-gray-400">
                  <div className="text-center">
                    <Play className="w-12 h-12 mx-auto mb-3 opacity-30" />
                    <p className="text-sm">Click "Send Request" to test the endpoint</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// JSON syntax highlighter component
const JsonHighlight = ({ json }) => {
  if (!json) return <span className="text-gray-500">Empty response</span>

  const highlighted = json
    .replace(/"([^"]+)":/g, '<span class="text-purple-400">"$1"</span>:')
    .replace(/: "([^"]*)"/g, ': <span class="text-green-400">"$1"</span>')
    .replace(/: (\d+\.?\d*)/g, ': <span class="text-yellow-400">$1</span>')
    .replace(/: (true|false)/g, ': <span class="text-blue-400">$1</span>')
    .replace(/: (null)/g, ': <span class="text-gray-500">$1</span>')

  return (
    <code dangerouslySetInnerHTML={{ __html: highlighted }} />
  )
}

export default TestEndpointModal
