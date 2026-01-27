import React, { useState } from 'react'
import { Copy, Check, FileJson, Braces } from 'lucide-react'

// Generate example body from JSON Schema
export const generateExampleFromSchema = (schema) => {
  if (!schema || !schema.properties) {
    return null
  }

  const example = {}

  Object.entries(schema.properties).forEach(([key, prop]) => {
    // Use example value if provided
    if (prop.example !== undefined) {
      example[key] = prop.example
    } else {
      // Generate default based on type
      switch (prop.type) {
        case 'string':
          if (prop.format === 'date-time') {
            example[key] = new Date().toISOString()
          } else if (prop.format === 'date') {
            example[key] = new Date().toISOString().split('T')[0]
          } else if (prop.format === 'uuid' || key.endsWith('_uuid')) {
            // Generate UUID for uuid format or fields ending with _uuid (FK fields)
            example[key] = '550e8400-e29b-41d4-a716-446655440000'
          } else {
            example[key] = `example_${key}`
          }
          break
        case 'integer':
        case 'number':
          example[key] = prop.type === 'integer' ? 1 : 1.0
          break
        case 'boolean':
          example[key] = true
          break
        case 'array':
          if (prop.items?.properties) {
            example[key] = [generateExampleFromSchema(prop.items)]
          } else {
            example[key] = []
          }
          break
        case 'object':
          if (prop.properties) {
            example[key] = generateExampleFromSchema(prop)
          } else {
            example[key] = {}
          }
          break
        default:
          example[key] = null
      }
    }
  })

  return example
}

// Example Body Viewer - shows the actual JSON example
export const ExampleBody = ({ schema, title, type = 'request', className = '' }) => {
  const [copied, setCopied] = useState(false)

  if (!schema || Object.keys(schema).length === 0) {
    return (
      <div className={`text-center py-4 text-gray-400 text-sm ${className}`}>
        <FileJson className="w-6 h-6 mx-auto mb-1 opacity-50" />
        No {type} body
      </div>
    )
  }

  const exampleBody = generateExampleFromSchema(schema)
  const jsonString = JSON.stringify(exampleBody, null, 2)

  const handleCopy = () => {
    navigator.clipboard.writeText(jsonString)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const typeConfig = {
    request: { color: 'text-green-600', bg: 'bg-green-50', border: 'border-green-200' },
    response: { color: 'text-blue-600', bg: 'bg-blue-50', border: 'border-blue-200' },
  }

  const config = typeConfig[type] || typeConfig.response

  return (
    <div className={`border ${config.border} rounded-lg overflow-hidden ${className}`}>
      {/* Header */}
      <div className={`flex items-center justify-between px-3 py-2 ${config.bg} border-b ${config.border}`}>
        <div className="flex items-center gap-2">
          <Braces className={`w-4 h-4 ${config.color}`} />
          <span className={`text-sm font-medium ${config.color}`}>{title}</span>
        </div>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 px-2 py-1 text-xs rounded hover:bg-white/50 transition-colors"
        >
          {copied ? (
            <>
              <Check className="w-3.5 h-3.5 text-green-600" />
              <span className="text-green-600">Copied!</span>
            </>
          ) : (
            <>
              <Copy className="w-3.5 h-3.5 text-gray-500" />
              <span className="text-gray-500">Copy</span>
            </>
          )}
        </button>
      </div>

      {/* JSON Content with syntax highlighting */}
      <div className="max-h-64 overflow-auto bg-gray-900 p-3">
        <pre className="text-sm font-mono">
          <JsonSyntaxHighlight json={jsonString} />
        </pre>
      </div>
    </div>
  )
}

// Simple JSON syntax highlighter
const JsonSyntaxHighlight = ({ json }) => {
  // Tokenize and colorize JSON
  const highlighted = json
    .replace(/"([^"]+)":/g, '<span class="text-purple-400">"$1"</span>:')
    .replace(/: "([^"]*)"/g, ': <span class="text-green-400">"$1"</span>')
    .replace(/: (\d+\.?\d*)/g, ': <span class="text-yellow-400">$1</span>')
    .replace(/: (true|false)/g, ': <span class="text-blue-400">$1</span>')
    .replace(/: (null)/g, ': <span class="text-gray-500">$1</span>')

  return (
    <code
      className="text-gray-300"
      dangerouslySetInnerHTML={{ __html: highlighted }}
    />
  )
}

export default ExampleBody
