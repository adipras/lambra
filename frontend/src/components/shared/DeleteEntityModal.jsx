import { useState } from 'react'
import { AlertTriangle, Trash2, X, Database, Table } from 'lucide-react'
import { LoadingSpinner } from './LoadingSpinner'

export const DeleteEntityModal = ({
  entityName,
  tableName,
  endpointsCount = 0,
  isDeployed = false,
  onConfirm,
  onCancel,
  isDeleting = false
}) => {
  const [confirmText, setConfirmText] = useState('')

  const expectedText = entityName.toLowerCase()
  const isConfirmValid = confirmText === expectedText

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl max-w-md w-full overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="bg-orange-50 border-b border-orange-100 px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-orange-100 rounded-full">
              <AlertTriangle className="w-6 h-6 text-orange-600" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-orange-900">Delete Entity</h2>
              <p className="text-sm text-orange-700">This will affect your database</p>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          <p className="text-gray-700">
            You are about to delete the entity:
          </p>

          <div className="p-4 bg-gray-100 rounded-xl">
            <div className="flex items-center gap-3">
              <Database className="w-5 h-5 text-gray-500" />
              <div>
                <p className="font-bold text-gray-900">{entityName}</p>
                <p className="text-sm text-gray-500 font-mono">Table: {tableName}</p>
              </div>
            </div>
          </div>

          {/* Warning Box */}
          <div className="bg-orange-50 border border-orange-200 rounded-xl p-4">
            <h4 className="font-semibold text-orange-800 mb-2 flex items-center gap-2">
              <Table className="w-4 h-4" />
              What will happen:
            </h4>
            <ul className="text-sm text-orange-700 space-y-1.5">
              <li className="flex items-start gap-2">
                <span className="text-orange-500 mt-0.5">•</span>
                <span>Entity definition will be deleted immediately</span>
              </li>
              {endpointsCount > 0 && (
                <li className="flex items-start gap-2">
                  <span className="text-orange-500 mt-0.5">•</span>
                  <span><strong>{endpointsCount} endpoint{endpointsCount > 1 ? 's' : ''}</strong> will also be deleted</span>
                </li>
              )}
              {isDeployed && (
                <li className="flex items-start gap-2">
                  <span className="text-red-500 mt-0.5">•</span>
                  <span className="text-red-700">
                    <strong>Database table "{tableName}" will be DROPPED</strong> on next deploy
                  </span>
                </li>
              )}
            </ul>
          </div>

          {isDeployed && (
            <div className="bg-red-50 border border-red-200 rounded-xl p-4">
              <p className="text-sm text-red-700">
                <strong>Warning:</strong> Your service is currently deployed. On the next deployment,
                the table <code className="font-mono bg-red-100 px-1 rounded">{tableName}</code> and
                all its data will be permanently deleted.
              </p>
            </div>
          )}

          {/* Confirmation Input */}
          <div className="space-y-2">
            <p className="text-sm text-gray-600">
              Type <code className="font-mono bg-gray-100 px-1 rounded">{expectedText}</code> to confirm:
            </p>
            <input
              type="text"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              placeholder="Type entity name..."
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent"
              autoFocus
            />
          </div>
        </div>

        {/* Footer */}
        <div className="border-t border-gray-200 px-6 py-4 flex items-center justify-between bg-gray-50">
          <button
            onClick={onCancel}
            disabled={isDeleting}
            className="flex items-center gap-1.5 px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-4 h-4" />
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={isDeleting || !isConfirmValid}
            className="flex items-center gap-1.5 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-300 text-white rounded-lg transition-colors"
          >
            {isDeleting ? (
              <>
                <LoadingSpinner size="sm" />
                Deleting...
              </>
            ) : (
              <>
                <Trash2 className="w-4 h-4" />
                Delete Entity
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

export default DeleteEntityModal
