import { useState } from 'react'
import { AlertTriangle, Trash2, X } from 'lucide-react'
import { LoadingSpinner } from './LoadingSpinner'

export const DeleteServiceModal = ({
  serviceName,
  onConfirm,
  onCancel,
  isDeleting = false
}) => {
  const [confirmText, setConfirmText] = useState('')
  const [step, setStep] = useState(1)

  const expectedText = serviceName.toLowerCase().replace(/\s+/g, '-')
  const isConfirmValid = confirmText === expectedText

  const handleProceed = () => {
    if (step === 1) {
      setStep(2)
    } else if (isConfirmValid) {
      onConfirm()
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl max-w-md w-full overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="bg-red-50 border-b border-red-100 px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-red-100 rounded-full">
              <AlertTriangle className="w-6 h-6 text-red-600" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-red-900">Delete Service</h2>
              <p className="text-sm text-red-700">This action cannot be undone</p>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="p-6">
          {step === 1 ? (
            <div className="space-y-4">
              <p className="text-gray-700">
                You are about to permanently delete the service:
              </p>
              <div className="p-4 bg-gray-100 rounded-xl">
                <p className="font-bold text-gray-900 text-lg">{serviceName}</p>
              </div>
              <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-4">
                <h4 className="font-semibold text-yellow-800 mb-2">This will permanently delete:</h4>
                <ul className="text-sm text-yellow-700 space-y-1.5">
                  <li className="flex items-start gap-2">
                    <span className="text-yellow-600 mt-0.5">-</span>
                    <span>All Docker containers and volumes for this service</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-yellow-600 mt-0.5">-</span>
                    <span>Generated source code in the workspace directory</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-yellow-600 mt-0.5">-</span>
                    <span>All deployment configurations and logs</span>
                  </li>
                </ul>
              </div>
              <p className="text-sm text-gray-500">
                Note: The project definition (entities, endpoints) will remain in the database.
                You can redeploy at any time.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-gray-700">
                To confirm deletion, please type the service name:
              </p>
              <div className="p-3 bg-gray-100 rounded-lg font-mono text-sm text-center">
                {expectedText}
              </div>
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="Type service name to confirm..."
                className="input w-full"
                autoFocus
              />
              {confirmText && !isConfirmValid && (
                <p className="text-sm text-red-600">
                  Text doesn't match. Please type exactly: <code className="font-mono bg-red-50 px-1 rounded">{expectedText}</code>
                </p>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="border-t border-gray-200 px-6 py-4 flex items-center justify-between bg-gray-50">
          <button
            onClick={onCancel}
            disabled={isDeleting}
            className="btn btn-secondary"
          >
            <X className="w-4 h-4 mr-1.5" />
            Cancel
          </button>
          <div className="flex items-center gap-2">
            {step === 2 && (
              <button
                onClick={() => setStep(1)}
                disabled={isDeleting}
                className="btn btn-secondary"
              >
                Back
              </button>
            )}
            <button
              onClick={handleProceed}
              disabled={isDeleting || (step === 2 && !isConfirmValid)}
              className={`btn ${
                step === 2
                  ? 'bg-red-600 hover:bg-red-700 text-white disabled:bg-red-300'
                  : 'bg-orange-500 hover:bg-orange-600 text-white'
              }`}
            >
              {isDeleting ? (
                <>
                  <LoadingSpinner size="sm" />
                  Deleting...
                </>
              ) : step === 1 ? (
                'Continue'
              ) : (
                <>
                  <Trash2 className="w-4 h-4 mr-1.5" />
                  Delete Forever
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
