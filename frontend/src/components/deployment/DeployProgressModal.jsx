import React, { useState, useEffect, useRef } from 'react'
import { X, Check, Loader2, AlertCircle, Clock, Rocket, FileCode, FolderPlus, Box, Play, CheckCircle2 } from 'lucide-react'
import clsx from 'clsx'

// Deployment steps configuration
const DEPLOY_STEPS = [
  { id: 'init', label: 'Initializing', icon: Rocket, description: 'Preparing deployment...' },
  { id: 'snapshot', label: 'Creating Snapshot', icon: Clock, description: 'Backing up current state...' },
  { id: 'generate_code', label: 'Generating Code', icon: FileCode, description: 'Creating service files...' },
  { id: 'write_files', label: 'Writing Files', icon: FolderPlus, description: 'Writing to workspace...' },
  { id: 'docker_build', label: 'Building Container', icon: Box, description: 'Building Docker image...' },
  { id: 'docker_start', label: 'Starting Service', icon: Play, description: 'Starting container...' },
  { id: 'complete', label: 'Complete', icon: CheckCircle2, description: 'Deployment finished!' },
]

/**
 * DeployProgressModal - Shows real-time deployment progress
 *
 * @param {boolean} isOpen - Whether modal is open
 * @param {function} onClose - Close handler
 * @param {string} projectId - Project UUID for SSE connection
 * @param {string} deploymentId - Deployment UUID (if available)
 * @param {function} onComplete - Called when deployment completes
 */
export const DeployProgressModal = ({
  isOpen,
  onClose,
  projectId,
  deploymentId,
  onComplete
}) => {
  const [currentStep, setCurrentStep] = useState('init')
  const [completedSteps, setCompletedSteps] = useState([])
  const [logs, setLogs] = useState([])
  const [error, setError] = useState(null)
  const [isComplete, setIsComplete] = useState(false)
  const [result, setResult] = useState(null)
  const logsEndRef = useRef(null)
  const eventSourceRef = useRef(null)

  // Auto-scroll logs
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  // Connect to SSE for deployment logs
  useEffect(() => {
    if (!isOpen || !deploymentId) return

    const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
    const eventSource = new EventSource(`${baseUrl}/deployments/${deploymentId}/logs/stream`)
    eventSourceRef.current = eventSource

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)

        // Add to logs
        setLogs(prev => [...prev, data])

        // Update current step based on log
        if (data.step) {
          setCurrentStep(data.step)

          // Mark previous steps as completed
          const stepIndex = DEPLOY_STEPS.findIndex(s => s.id === data.step)
          if (stepIndex > 0) {
            const completed = DEPLOY_STEPS.slice(0, stepIndex).map(s => s.id)
            setCompletedSteps(completed)
          }

          // Check if deployment is complete
          if (data.step === 'complete' && data.level === 'info') {
            setCompletedSteps(DEPLOY_STEPS.map(s => s.id))
            setIsComplete(true)
            setTimeout(() => {
              onComplete?.()
            }, 1500)
          }
        }

        // Check for errors
        if (data.level === 'error') {
          setError(data.message)
        }
      } catch (e) {
        console.error('Failed to parse deployment log:', e)
      }
    }

    eventSource.onerror = () => {
      eventSource.close()
    }

    return () => {
      eventSource.close()
    }
  }, [isOpen, deploymentId, onComplete])

  // Close SSE on unmount
  useEffect(() => {
    return () => {
      eventSourceRef.current?.close()
    }
  }, [])

  if (!isOpen) return null

  const currentStepIndex = DEPLOY_STEPS.findIndex(s => s.id === currentStep)
  const progress = ((completedSteps.length) / (DEPLOY_STEPS.length - 1)) * 100

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl w-full max-w-2xl max-h-[85vh] flex flex-col overflow-hidden shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3">
            {isComplete ? (
              <div className="p-2 bg-green-100 rounded-lg">
                <CheckCircle2 className="w-5 h-5 text-green-600" />
              </div>
            ) : error ? (
              <div className="p-2 bg-red-100 rounded-lg">
                <AlertCircle className="w-5 h-5 text-red-600" />
              </div>
            ) : (
              <div className="p-2 bg-indigo-100 rounded-lg">
                <Loader2 className="w-5 h-5 text-indigo-600 animate-spin" />
              </div>
            )}
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                {isComplete ? 'Deployment Complete!' : error ? 'Deployment Failed' : 'Deploying Service'}
              </h2>
              <p className="text-sm text-gray-500">
                {isComplete ? 'Your service is now running' : error ? 'An error occurred' : 'Please wait while we deploy your service...'}
              </p>
            </div>
          </div>
          {(isComplete || error) && (
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <X className="w-5 h-5 text-gray-500" />
            </button>
          )}
        </div>

        {/* Progress Bar */}
        <div className="px-6 py-4 bg-gray-50 border-b border-gray-200">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-gray-700">Progress</span>
            <span className="text-sm text-gray-500">{Math.round(progress)}%</span>
          </div>
          <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
            <div
              className={clsx(
                'h-full transition-all duration-500 ease-out rounded-full',
                error ? 'bg-red-500' : isComplete ? 'bg-green-500' : 'bg-indigo-600'
              )}
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>

        {/* Steps */}
        <div className="px-6 py-4 border-b border-gray-100">
          <div className="flex items-center justify-between">
            {DEPLOY_STEPS.map((step, index) => {
              const isCompleted = completedSteps.includes(step.id)
              const isCurrent = step.id === currentStep
              const isError = error && isCurrent
              const Icon = step.icon

              return (
                <React.Fragment key={step.id}>
                  {/* Step indicator */}
                  <div className="flex flex-col items-center">
                    <div
                      className={clsx(
                        'w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300',
                        isCompleted
                          ? 'bg-green-500 text-white'
                          : isError
                          ? 'bg-red-500 text-white'
                          : isCurrent
                          ? 'bg-indigo-600 text-white ring-4 ring-indigo-100'
                          : 'bg-gray-100 text-gray-400'
                      )}
                    >
                      {isCompleted ? (
                        <Check className="w-5 h-5" />
                      ) : isError ? (
                        <AlertCircle className="w-5 h-5" />
                      ) : isCurrent ? (
                        <Loader2 className="w-5 h-5 animate-spin" />
                      ) : (
                        <Icon className="w-5 h-5" />
                      )}
                    </div>
                    <span
                      className={clsx(
                        'text-xs mt-2 font-medium text-center max-w-[70px]',
                        isCompleted || isCurrent ? 'text-gray-900' : 'text-gray-400'
                      )}
                    >
                      {step.label}
                    </span>
                  </div>

                  {/* Connector line */}
                  {index < DEPLOY_STEPS.length - 1 && (
                    <div
                      className={clsx(
                        'flex-1 h-0.5 mx-1 rounded transition-colors duration-300',
                        completedSteps.includes(DEPLOY_STEPS[index + 1]?.id)
                          ? 'bg-green-500'
                          : index < currentStepIndex
                          ? 'bg-indigo-600'
                          : 'bg-gray-200'
                      )}
                    />
                  )}
                </React.Fragment>
              )
            })}
          </div>
        </div>

        {/* Current Step Description */}
        <div className="px-6 py-3 bg-gray-50 border-b border-gray-100">
          <p className="text-sm text-gray-600">
            {DEPLOY_STEPS.find(s => s.id === currentStep)?.description}
          </p>
        </div>

        {/* Logs */}
        <div className="flex-1 overflow-y-auto p-4 bg-gray-900 font-mono text-xs">
          {logs.length === 0 ? (
            <div className="text-gray-500 text-center py-8">
              Waiting for deployment logs...
            </div>
          ) : (
            logs.map((log, index) => (
              <div
                key={index}
                className={clsx(
                  'py-1 flex items-start gap-2',
                  log.level === 'error' ? 'text-red-400' :
                  log.level === 'warning' ? 'text-yellow-400' :
                  'text-gray-300'
                )}
              >
                <span className="text-gray-500 w-20 flex-shrink-0">
                  {new Date(log.created_at).toLocaleTimeString()}
                </span>
                <span className={clsx(
                  'w-14 flex-shrink-0 uppercase text-[10px] font-bold px-1.5 py-0.5 rounded',
                  log.level === 'error' ? 'bg-red-500/20 text-red-400' :
                  log.level === 'warning' ? 'bg-yellow-500/20 text-yellow-400' :
                  'bg-blue-500/20 text-blue-400'
                )}>
                  {log.level}
                </span>
                <span className="flex-1">{log.message}</span>
              </div>
            ))
          )}
          <div ref={logsEndRef} />
        </div>

        {/* Footer */}
        {(isComplete || error) && (
          <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
            <div className="flex justify-end gap-3">
              {result?.url && (
                <a
                  href={result.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn btn-secondary"
                >
                  Open Service
                </a>
              )}
              <button
                onClick={onClose}
                className={clsx(
                  'btn',
                  isComplete ? 'btn-primary' : 'btn-secondary'
                )}
              >
                {isComplete ? 'Done' : 'Close'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default DeployProgressModal
