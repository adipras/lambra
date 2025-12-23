import { useState, useMemo } from 'react'
import { ChevronRight, ChevronDown, Copy, Check, Type, Hash, List, ToggleLeft, Calendar, FileJson, AlertCircle } from 'lucide-react'

/**
 * SchemaViewer Component
 *
 * A visual JSON schema viewer with tree structure display.
 *
 * @param {Object} props
 * @param {object|string} props.schema - JSON schema object or string
 * @param {string} props.title - Title for the schema section
 * @param {string} props.emptyMessage - Message when schema is empty
 * @param {string} props.className - Additional CSS classes
 */
export const SchemaViewer = ({
  schema,
  title = 'Schema',
  emptyMessage = 'No schema defined',
  className = '',
}) => {
  const [copied, setCopied] = useState(false)

  // Parse schema if it's a string
  const parsedSchema = useMemo(() => {
    if (!schema) return null
    if (typeof schema === 'string') {
      try {
        return JSON.parse(schema)
      } catch {
        return null
      }
    }
    return schema
  }, [schema])

  const formattedJson = useMemo(() => {
    if (!parsedSchema) return ''
    return JSON.stringify(parsedSchema, null, 2)
  }, [parsedSchema])

  const handleCopy = async () => {
    if (!formattedJson) return
    try {
      await navigator.clipboard.writeText(formattedJson)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const isEmpty = !parsedSchema || (typeof parsedSchema === 'object' && Object.keys(parsedSchema).length === 0)

  return (
    <div className={`rounded-lg border border-gray-200 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-gray-50 border-b border-gray-200 rounded-t-lg">
        <h4 className="text-sm font-medium text-gray-700">{title}</h4>
        {!isEmpty && (
          <button
            onClick={handleCopy}
            className={`flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors ${
              copied
                ? 'bg-green-100 text-green-700'
                : 'hover:bg-gray-100 text-gray-500 hover:text-gray-700'
            }`}
          >
            {copied ? (
              <>
                <Check className="w-3 h-3" />
                Copied
              </>
            ) : (
              <>
                <Copy className="w-3 h-3" />
                Copy
              </>
            )}
          </button>
        )}
      </div>

      {/* Content */}
      <div className="p-4">
        {isEmpty ? (
          <div className="text-center py-6 text-gray-500 text-sm">
            <FileJson className="w-8 h-8 mx-auto mb-2 opacity-50" />
            {emptyMessage}
          </div>
        ) : (
          <SchemaTree schema={parsedSchema} />
        )}
      </div>
    </div>
  )
}

/**
 * SchemaTree Component
 *
 * Renders a tree view of the schema properties.
 */
const SchemaTree = ({ schema, level = 0 }) => {
  if (!schema || typeof schema !== 'object') {
    return null
  }

  // Handle different schema formats
  const properties = schema.properties || schema
  const required = schema.required || []

  return (
    <div className={level > 0 ? 'ml-4 border-l border-gray-200 pl-4' : ''}>
      {Object.entries(properties).map(([key, value]) => (
        <SchemaProperty
          key={key}
          name={key}
          value={value}
          isRequired={required.includes(key)}
          level={level}
        />
      ))}
    </div>
  )
}

/**
 * SchemaProperty Component
 *
 * Renders a single property in the schema tree.
 */
const SchemaProperty = ({ name, value, isRequired, level }) => {
  const [isExpanded, setIsExpanded] = useState(level < 2)

  // Determine property type
  const getPropertyType = (val) => {
    if (typeof val === 'string') return val
    if (val?.type) return val.type
    if (val?.properties || val?.items) return 'object'
    if (Array.isArray(val)) return 'array'
    return typeof val
  }

  const type = getPropertyType(value)
  const hasChildren = typeof value === 'object' && value !== null &&
    (value.properties || value.items || (typeof value !== 'string' && Object.keys(value).length > 0))

  const getTypeIcon = (t) => {
    switch (t) {
      case 'string':
        return <Type className="w-3 h-3" />
      case 'number':
      case 'integer':
      case 'int':
      case 'float':
        return <Hash className="w-3 h-3" />
      case 'array':
        return <List className="w-3 h-3" />
      case 'boolean':
      case 'bool':
        return <ToggleLeft className="w-3 h-3" />
      case 'date':
      case 'datetime':
        return <Calendar className="w-3 h-3" />
      case 'object':
        return <FileJson className="w-3 h-3" />
      default:
        return <Type className="w-3 h-3" />
    }
  }

  const getTypeBadgeColor = (t) => {
    switch (t) {
      case 'string':
        return 'bg-green-100 text-green-700 border-green-200'
      case 'number':
      case 'integer':
      case 'int':
      case 'float':
        return 'bg-blue-100 text-blue-700 border-blue-200'
      case 'array':
        return 'bg-purple-100 text-purple-700 border-purple-200'
      case 'boolean':
      case 'bool':
        return 'bg-orange-100 text-orange-700 border-orange-200'
      case 'object':
        return 'bg-gray-100 text-gray-700 border-gray-200'
      default:
        return 'bg-gray-100 text-gray-700 border-gray-200'
    }
  }

  return (
    <div className="py-1.5">
      <div className="flex items-center gap-2 group">
        {/* Expand/Collapse button */}
        {hasChildren ? (
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-0.5 rounded hover:bg-gray-100"
          >
            {isExpanded ? (
              <ChevronDown className="w-3 h-3 text-gray-500" />
            ) : (
              <ChevronRight className="w-3 h-3 text-gray-500" />
            )}
          </button>
        ) : (
          <span className="w-4" />
        )}

        {/* Property name */}
        <span className="font-mono text-sm text-gray-900">{name}</span>

        {/* Required badge */}
        {isRequired && (
          <span className="text-red-500 text-xs">*</span>
        )}

        {/* Type badge */}
        <span className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs border ${getTypeBadgeColor(type)}`}>
          {getTypeIcon(type)}
          {type}
        </span>

        {/* Description */}
        {value?.description && (
          <span className="text-xs text-gray-500 truncate max-w-xs">
            - {value.description}
          </span>
        )}

        {/* Example */}
        {value?.example !== undefined && (
          <span className="text-xs text-gray-400 font-mono">
            e.g. {JSON.stringify(value.example)}
          </span>
        )}
      </div>

      {/* Nested properties */}
      {hasChildren && isExpanded && (
        <SchemaTree
          schema={value.properties || value.items || value}
          level={level + 1}
        />
      )}
    </div>
  )
}

/**
 * JsonEditor Component
 *
 * A textarea-based JSON editor with validation.
 */
export const JsonEditor = ({
  value,
  onChange,
  placeholder = '{}',
  label,
  error,
  className = '',
}) => {
  const [localError, setLocalError] = useState('')

  const handleChange = (e) => {
    const newValue = e.target.value
    onChange(newValue)

    // Validate JSON
    if (newValue.trim()) {
      try {
        JSON.parse(newValue)
        setLocalError('')
      } catch (err) {
        setLocalError('Invalid JSON format')
      }
    } else {
      setLocalError('')
    }
  }

  const formatJson = () => {
    if (!value) return
    try {
      const parsed = JSON.parse(value)
      onChange(JSON.stringify(parsed, null, 2))
      setLocalError('')
    } catch (err) {
      setLocalError('Cannot format: Invalid JSON')
    }
  }

  const displayError = error || localError

  return (
    <div className={className}>
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1">
          {label}
        </label>
      )}
      <div className="relative">
        <textarea
          value={value}
          onChange={handleChange}
          placeholder={placeholder}
          className={`w-full h-40 p-3 font-mono text-sm rounded-lg border ${
            displayError
              ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
              : 'border-gray-300 focus:border-indigo-500 focus:ring-indigo-500'
          } focus:ring-1`}
        />
        <button
          type="button"
          onClick={formatJson}
          className="absolute top-2 right-2 px-2 py-1 text-xs bg-gray-100 hover:bg-gray-200 rounded text-gray-600"
        >
          Format
        </button>
      </div>
      {displayError && (
        <p className="mt-1 text-sm text-red-600 flex items-center gap-1">
          <AlertCircle className="w-3 h-3" />
          {displayError}
        </p>
      )}
    </div>
  )
}

export default SchemaViewer
