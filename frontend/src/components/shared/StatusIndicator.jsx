import React from 'react'
import clsx from 'clsx'

/**
 * StatusIndicator - Enhanced status indicator with animations
 *
 * @param {string} status - running, stopped, not_deployed, deploying, error
 * @param {string} size - sm, md, lg
 * @param {boolean} showLabel - whether to show the status label
 * @param {boolean} pulse - whether to show pulse animation
 */
export const StatusIndicator = ({
  status = 'not_deployed',
  size = 'md',
  showLabel = true,
  pulse = true,
  className = ''
}) => {
  const statusConfig = {
    running: {
      label: 'Running',
      dotColor: 'bg-green-500',
      bgColor: 'bg-green-50',
      textColor: 'text-green-700',
      borderColor: 'border-green-200',
      ringColor: 'ring-green-400',
      pulseColor: 'bg-green-400',
    },
    stopped: {
      label: 'Stopped',
      dotColor: 'bg-amber-500',
      bgColor: 'bg-amber-50',
      textColor: 'text-amber-700',
      borderColor: 'border-amber-200',
      ringColor: 'ring-amber-400',
      pulseColor: 'bg-amber-400',
    },
    not_deployed: {
      label: 'Not Deployed',
      dotColor: 'bg-gray-400',
      bgColor: 'bg-gray-50',
      textColor: 'text-gray-600',
      borderColor: 'border-gray-200',
      ringColor: 'ring-gray-300',
      pulseColor: 'bg-gray-300',
    },
    deploying: {
      label: 'Deploying',
      dotColor: 'bg-blue-500',
      bgColor: 'bg-blue-50',
      textColor: 'text-blue-700',
      borderColor: 'border-blue-200',
      ringColor: 'ring-blue-400',
      pulseColor: 'bg-blue-400',
    },
    error: {
      label: 'Error',
      dotColor: 'bg-red-500',
      bgColor: 'bg-red-50',
      textColor: 'text-red-700',
      borderColor: 'border-red-200',
      ringColor: 'ring-red-400',
      pulseColor: 'bg-red-400',
    },
  }

  const sizeConfig = {
    sm: {
      container: 'px-2 py-1 gap-1.5',
      dot: 'w-1.5 h-1.5',
      pulseRing: 'w-3 h-3',
      text: 'text-xs',
    },
    md: {
      container: 'px-3 py-1.5 gap-2',
      dot: 'w-2 h-2',
      pulseRing: 'w-4 h-4',
      text: 'text-sm',
    },
    lg: {
      container: 'px-4 py-2 gap-2.5',
      dot: 'w-2.5 h-2.5',
      pulseRing: 'w-5 h-5',
      text: 'text-base',
    },
  }

  const config = statusConfig[status] || statusConfig.not_deployed
  const sizes = sizeConfig[size] || sizeConfig.md
  const shouldPulse = pulse && (status === 'running' || status === 'deploying')

  return (
    <div
      className={clsx(
        'inline-flex items-center rounded-full font-medium border transition-all duration-300',
        config.bgColor,
        config.textColor,
        config.borderColor,
        sizes.container,
        className
      )}
    >
      {/* Animated Dot */}
      <div className="relative flex items-center justify-center">
        {shouldPulse && (
          <>
            {/* Outer pulse ring */}
            <span
              className={clsx(
                'absolute rounded-full opacity-75 animate-ping',
                config.pulseColor,
                sizes.pulseRing
              )}
              style={{ animationDuration: '1.5s' }}
            />
            {/* Inner glow */}
            <span
              className={clsx(
                'absolute rounded-full opacity-40 animate-pulse',
                config.pulseColor,
                sizes.dot,
                'scale-150'
              )}
            />
          </>
        )}
        {/* Main dot */}
        <span
          className={clsx(
            'relative rounded-full',
            config.dotColor,
            sizes.dot,
            status === 'deploying' && 'animate-bounce'
          )}
        />
      </div>

      {/* Label */}
      {showLabel && (
        <span className={clsx('font-medium', sizes.text)}>
          {config.label}
        </span>
      )}
    </div>
  )
}

/**
 * StatusDot - Just the dot part, for inline use
 */
export const StatusDot = ({ status = 'not_deployed', size = 'md', pulse = true }) => {
  const colors = {
    running: 'bg-green-500',
    stopped: 'bg-amber-500',
    not_deployed: 'bg-gray-400',
    deploying: 'bg-blue-500',
    error: 'bg-red-500',
  }

  const pulseColors = {
    running: 'bg-green-400',
    deploying: 'bg-blue-400',
  }

  const sizes = {
    sm: 'w-1.5 h-1.5',
    md: 'w-2 h-2',
    lg: 'w-3 h-3',
  }

  const pulseSizes = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  }

  const shouldPulse = pulse && (status === 'running' || status === 'deploying')

  return (
    <div className="relative flex items-center justify-center">
      {shouldPulse && (
        <span
          className={clsx(
            'absolute rounded-full opacity-75 animate-ping',
            pulseColors[status],
            pulseSizes[size]
          )}
          style={{ animationDuration: '1.5s' }}
        />
      )}
      <span
        className={clsx(
          'relative rounded-full',
          colors[status] || colors.not_deployed,
          sizes[size] || sizes.md
        )}
      />
    </div>
  )
}
