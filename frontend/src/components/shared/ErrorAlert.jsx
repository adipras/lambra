import React from 'react'
import { AlertCircle, X } from 'lucide-react'

export const ErrorAlert = ({ message, onClose }) => {
  if (!message) return null

  return (
    <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
      <div className="flex items-start">
        <AlertCircle className="w-5 h-5 text-red-600 mt-0.5" />
        <div className="ml-3 flex-1">
          <h3 className="text-sm font-medium text-red-800">Error</h3>
          <div className="mt-1 text-sm text-red-700">{message}</div>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="ml-auto text-red-400 hover:text-red-600"
          >
            <X className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  )
}
