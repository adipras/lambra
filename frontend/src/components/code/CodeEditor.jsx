import { useState, useMemo } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Copy, Check, Sun, Moon, Maximize2, Minimize2 } from 'lucide-react'

/**
 * CodeEditor Component
 *
 * A syntax-highlighted code viewer with line numbers and copy functionality.
 *
 * @param {Object} props
 * @param {string} props.code - The code to display
 * @param {string} props.language - Programming language for syntax highlighting (go, javascript, sql, etc.)
 * @param {string} props.filename - Optional filename to display
 * @param {boolean} props.showLineNumbers - Show line numbers (default: true)
 * @param {number} props.startingLineNumber - Starting line number (default: 1)
 * @param {string} props.className - Additional CSS classes
 * @param {boolean} props.wrapLines - Wrap long lines (default: false)
 * @param {number} props.maxHeight - Max height in pixels (default: none)
 * @param {Function} props.onCopy - Callback when code is copied
 */
export const CodeEditor = ({
  code = '',
  language = 'go',
  filename = '',
  showLineNumbers = true,
  startingLineNumber = 1,
  className = '',
  wrapLines = false,
  maxHeight,
  onCopy,
}) => {
  const [copied, setCopied] = useState(false)
  const [isDark, setIsDark] = useState(true)
  const [isExpanded, setIsExpanded] = useState(false)

  // Detect language from filename extension if not explicitly provided
  const detectedLanguage = useMemo(() => {
    if (language) return language

    const ext = filename.split('.').pop()?.toLowerCase()
    const languageMap = {
      go: 'go',
      js: 'javascript',
      jsx: 'jsx',
      ts: 'typescript',
      tsx: 'tsx',
      sql: 'sql',
      json: 'json',
      yaml: 'yaml',
      yml: 'yaml',
      md: 'markdown',
      sh: 'bash',
      bash: 'bash',
      py: 'python',
      rb: 'ruby',
      java: 'java',
      rs: 'rust',
      html: 'html',
      css: 'css',
      scss: 'scss',
      xml: 'xml',
      dockerfile: 'docker',
    }
    return languageMap[ext] || 'text'
  }, [language, filename])

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      onCopy?.()
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const lineCount = useMemo(() => {
    return code.split('\n').length
  }, [code])

  const containerStyle = {
    maxHeight: isExpanded ? 'none' : maxHeight ? `${maxHeight}px` : undefined,
  }

  return (
    <div className={`relative rounded-lg overflow-hidden border border-gray-700 ${className}`}>
      {/* Header */}
      <div className={`flex items-center justify-between px-4 py-2 border-b ${
        isDark ? 'bg-gray-800 border-gray-700' : 'bg-gray-100 border-gray-300'
      }`}>
        <div className="flex items-center gap-3">
          {/* Filename */}
          {filename && (
            <span className={`text-sm font-mono ${isDark ? 'text-gray-300' : 'text-gray-600'}`}>
              {filename}
            </span>
          )}
          {/* Language badge */}
          <span className={`text-xs px-2 py-0.5 rounded ${
            isDark ? 'bg-gray-700 text-gray-400' : 'bg-gray-200 text-gray-600'
          }`}>
            {detectedLanguage}
          </span>
          {/* Line count */}
          <span className={`text-xs ${isDark ? 'text-gray-500' : 'text-gray-400'}`}>
            {lineCount} lines
          </span>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {/* Theme toggle */}
          <button
            onClick={() => setIsDark(!isDark)}
            className={`p-1.5 rounded transition-colors ${
              isDark
                ? 'hover:bg-gray-700 text-gray-400 hover:text-gray-200'
                : 'hover:bg-gray-200 text-gray-500 hover:text-gray-700'
            }`}
            title={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
          >
            {isDark ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          </button>

          {/* Expand/Collapse toggle */}
          {maxHeight && (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className={`p-1.5 rounded transition-colors ${
                isDark
                  ? 'hover:bg-gray-700 text-gray-400 hover:text-gray-200'
                  : 'hover:bg-gray-200 text-gray-500 hover:text-gray-700'
              }`}
              title={isExpanded ? 'Collapse' : 'Expand'}
            >
              {isExpanded ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
            </button>
          )}

          {/* Copy button */}
          <button
            onClick={handleCopy}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded text-sm transition-colors ${
              copied
                ? 'bg-green-600 text-white'
                : isDark
                  ? 'bg-gray-700 hover:bg-gray-600 text-gray-300'
                  : 'bg-gray-200 hover:bg-gray-300 text-gray-700'
            }`}
          >
            {copied ? (
              <>
                <Check className="w-4 h-4" />
                Copied!
              </>
            ) : (
              <>
                <Copy className="w-4 h-4" />
                Copy
              </>
            )}
          </button>
        </div>
      </div>

      {/* Code Content */}
      <div
        className="overflow-auto"
        style={containerStyle}
      >
        <SyntaxHighlighter
          language={detectedLanguage}
          style={isDark ? oneDark : oneLight}
          showLineNumbers={showLineNumbers}
          startingLineNumber={startingLineNumber}
          wrapLines={wrapLines}
          wrapLongLines={wrapLines}
          customStyle={{
            margin: 0,
            padding: '1rem',
            fontSize: '0.875rem',
            background: isDark ? '#282c34' : '#fafafa',
          }}
          lineNumberStyle={{
            minWidth: '3em',
            paddingRight: '1em',
            color: isDark ? '#636d83' : '#999',
            userSelect: 'none',
          }}
          codeTagProps={{
            style: {
              fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
            },
          }}
        >
          {code || '// No code to display'}
        </SyntaxHighlighter>
      </div>
    </div>
  )
}

/**
 * CodeBlock Component
 *
 * A simplified inline code block for smaller code snippets.
 */
export const CodeBlock = ({ code, language = 'text', className = '' }) => {
  return (
    <SyntaxHighlighter
      language={language}
      style={oneDark}
      customStyle={{
        margin: 0,
        padding: '0.5rem 0.75rem',
        fontSize: '0.75rem',
        borderRadius: '0.375rem',
        display: 'inline-block',
      }}
      codeTagProps={{
        style: {
          fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
        },
      }}
    >
      {code}
    </SyntaxHighlighter>
  )
}

export default CodeEditor
