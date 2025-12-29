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

export const TestEndpointModal = ({ endpoint, serviceUrl, onClose }) => {
  const [requestBody, setRequestBody] = useState('')
  const [response, setResponse] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState(null)
  const [copied, setCopied] = useState(false)

  // Generate example request body from schema
  useEffect(() => {
    if (endpoint?.request_schema && Object.keys(endpoint.request_schema).length > 0) {
      const example = generateExampleFromSchema(endpoint.request_schema)
      setRequestBody(JSON.stringify(example, null, 2))
    } else {
      setRequestBody('')
    }
  }, [endpoint])

  const handleTest = async () => {
    setIsLoading(true)
    setError(null)
    setResponse(null)

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

      // Call the test endpoint API
      const result = await endpointsApi.test(endpoint.id, {
        body: body,
        headers: { 'Content-Type': 'application/json' }
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
              <p className="text-sm font-mono text-gray-500">{serviceUrl}{endpoint?.path}</p>
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
              {needsBody && (
                <p className="text-xs text-green-600 mt-0.5">Edit the JSON body below</p>
              )}
            </div>

            <div className="flex-1 overflow-auto p-4">
              {needsBody ? (
                <div className="h-full">
                  <textarea
                    value={requestBody}
                    onChange={(e) => setRequestBody(e.target.value)}
                    className="w-full h-full min-h-[200px] font-mono text-sm p-3 bg-gray-900 text-gray-100 rounded-lg border-0 resize-none focus:ring-2 focus:ring-green-500"
                    placeholder="Enter request body JSON..."
                    spellCheck={false}
                  />
                </div>
              ) : (
                <div className="flex items-center justify-center h-full text-gray-400">
                  <div className="text-center">
                    <p className="text-sm">No request body required for {endpoint?.method}</p>
                  </div>
                </div>
              )}
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
