import React from 'react'
import clsx from 'clsx'

const statusConfig = {
  active: {
    label: 'Active',
    className: 'bg-green-100 text-green-800 border-green-200',
  },
  generating: {
    label: 'Generating',
    className: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  },
  failed: {
    label: 'Failed',
    className: 'bg-red-100 text-red-800 border-red-200',
  },
  archived: {
    label: 'Archived',
    className: 'bg-gray-100 text-gray-800 border-gray-200',
  },
  pending: {
    label: 'Pending',
    className: 'bg-blue-100 text-blue-800 border-blue-200',
  },
  deploying: {
    label: 'Deploying',
    className: 'bg-purple-100 text-purple-800 border-purple-200',
  },
  success: {
    label: 'Success',
    className: 'bg-green-100 text-green-800 border-green-200',
  },
}

export const StatusBadge = ({ status }) => {
  const config = statusConfig[status] || statusConfig.active

  return (
    <span
      className={clsx(
        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border',
        config.className
      )}
    >
      {config.label}
    </span>
  )
}
