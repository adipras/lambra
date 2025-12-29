import React, { useState } from 'react'
import { ChevronDown, ChevronRight, Copy, Check, Code2, FileJson, Braces } from 'lucide-react'

// Type badge colors
const TYPE_COLORS = {
  string: 'bg-blue-100 text-blue-700',
  integer: 'bg-green-100 text-green-700',
  number: 'bg-green-100 text-green-700',
  boolean: 'bg-purple-100 text-purple-700',
  object: 'bg-orange-100 text-orange-700',
  array: 'bg-pink-100 text-pink-700',
}

// Property row component
const PropertyRow = ({ name, schema, required, depth = 0 }) => {
  const [isExpanded, setIsExpanded] = useState(depth < 2)
  const hasChildren = schema.type === 'object' && schema.properties
  const isArray = schema.type === 'array'

  const typeColor = TYPE_COLORS[schema.type] || 'bg-gray-100 text-gray-700'

  return (
    <div className="border-l-2 border-gray-100 ml-2">
      <div
        className={`flex items-start gap-2 py-1.5 px-2 hover:bg-gray-50 rounded-r ${hasChildren || isArray ? 'cursor-pointer' : ''}`}
        onClick={() => (hasChildren || isArray) && setIsExpanded(!isExpanded)}
      >
        {(hasChildren || isArray) ? (
          <button className="p-0.5 text-gray-400 hover:text-gray-600">
            {isExpanded ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
          </button>
        ) : (
          <div className="w-4" />
        )}

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-mono text-sm font-medium text-gray-900">{name}</span>
            {required && (
              <span className="text-[10px] px-1 py-0.5 bg-red-100 text-red-600 rounded font-medium">required</span>
            )}
            <span className={`text-[10px] px-1.5 py-0.5 rounded font-medium ${typeColor}`}>
              {schema.type}
              {schema.format && <span className="opacity-75"> ({schema.format})</span>}
              {isArray && schema.items?.type && <span className="opacity-75">[{schema.items.type}]</span>}
            </span>
            {schema.maxLength && (
              <span className="text-[10px] text-gray-400">max: {schema.maxLength}</span>
            )}
          </div>
          {schema.description && (
            <p className="text-xs text-gray-500 mt-0.5">{schema.description}</p>
          )}
          {schema.example !== undefined && (
            <p className="text-xs text-gray-400 mt-0.5 font-mono">
              example: <span className="text-indigo-600">{JSON.stringify(schema.example)}</span>
            </p>
          )}
        </div>
      </div>

      {/* Nested properties */}
      {isExpanded && hasChildren && (
        <div className="ml-4">
          {Object.entries(schema.properties).map(([key, value]) => (
            <PropertyRow
              key={key}
              name={key}
              schema={value}
              required={schema.required?.includes(key)}
              depth={depth + 1}
            />
          ))}
        </div>
      )}

      {/* Array items */}
      {isExpanded && isArray && schema.items?.properties && (
        <div className="ml-4">
          <div className="text-xs text-gray-400 py-1 px-2 italic">array items:</div>
          {Object.entries(schema.items.properties).map(([key, value]) => (
            <PropertyRow
              key={key}
              name={key}
              schema={value}
              required={schema.items.required?.includes(key)}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// Main SchemaViewer component
export const SchemaViewer = ({ schema, title, type = 'request' }) => {
  const [viewMode, setViewMode] = useState('visual') // 'visual' or 'json'
  const [copied, setCopied] = useState(false)

  if (!schema || Object.keys(schema).length === 0) {
    return (
      <div className="text-center py-6 text-gray-400 text-sm">
        <FileJson className="w-8 h-8 mx-auto mb-2 opacity-50" />
        No schema defined
      </div>
    )
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(JSON.stringify(schema, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const typeConfig = {
    request: { color: 'text-green-600', bg: 'bg-green-50', label: 'Request Body' },
    response: { color: 'text-blue-600', bg: 'bg-blue-50', label: 'Response' },
  }

  const config = typeConfig[type] || typeConfig.response

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div className={`flex items-center justify-between px-3 py-2 ${config.bg} border-b border-gray-200`}>
        <div className="flex items-center gap-2">
          <Braces className={`w-4 h-4 ${config.color}`} />
          <span className={`text-sm font-medium ${config.color}`}>{title || config.label}</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setViewMode('visual')}
            className={`p-1.5 rounded text-xs ${viewMode === 'visual' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}`}
            title="Visual view"
          >
            <Code2 className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={() => setViewMode('json')}
            className={`p-1.5 rounded text-xs ${viewMode === 'json' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}`}
            title="JSON view"
          >
            <FileJson className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={handleCopy}
            className="p-1.5 rounded text-gray-500 hover:text-gray-700 hover:bg-white"
            title="Copy JSON"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-green-600" /> : <Copy className="w-3.5 h-3.5" />}
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="max-h-80 overflow-auto bg-white">
        {viewMode === 'visual' ? (
          <div className="p-2">
            {schema.properties ? (
              Object.entries(schema.properties).map(([key, value]) => (
                <PropertyRow
                  key={key}
                  name={key}
                  schema={value}
                  required={schema.required?.includes(key)}
                />
              ))
            ) : (
              <div className="text-sm text-gray-500 p-2">No properties defined</div>
            )}
          </div>
        ) : (
          <pre className="p-3 text-xs font-mono text-gray-800 overflow-x-auto">
            {JSON.stringify(schema, null, 2)}
          </pre>
        )}
      </div>
    </div>
  )
}

// Compact inline schema preview
export const SchemaPreview = ({ schema, maxFields = 3 }) => {
  if (!schema?.properties) return null

  const properties = Object.entries(schema.properties)
  const displayProps = properties.slice(0, maxFields)
  const remaining = properties.length - maxFields

  return (
    <div className="flex flex-wrap gap-1">
      {displayProps.map(([key, value]) => (
        <span
          key={key}
          className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-mono ${TYPE_COLORS[value.type] || 'bg-gray-100 text-gray-600'}`}
        >
          {key}
        </span>
      ))}
      {remaining > 0 && (
        <span className="text-[10px] text-gray-400 px-1">+{remaining} more</span>
      )}
    </div>
  )
}

export default SchemaViewer
