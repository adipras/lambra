import React from 'react'
import clsx from 'clsx'

/**
 * Base Skeleton component with shimmer animation
 */
export const Skeleton = ({ className, ...props }) => {
  return (
    <div
      className={clsx(
        'animate-pulse bg-gradient-to-r from-gray-200 via-gray-100 to-gray-200 bg-[length:200%_100%] rounded',
        className
      )}
      style={{
        animation: 'shimmer 1.5s infinite linear',
      }}
      {...props}
    />
  )
}

/**
 * Skeleton for Entity Card
 */
export const EntityCardSkeleton = () => {
  return (
    <div className="border border-gray-200 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between p-4 bg-gray-50">
        <div className="flex items-center gap-3">
          <Skeleton className="w-10 h-10 rounded-lg" />
          <div className="space-y-2">
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-24" />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="w-8 h-8 rounded-lg" />
          <Skeleton className="w-8 h-8 rounded-lg" />
          <Skeleton className="w-8 h-8 rounded-lg" />
        </div>
      </div>

      {/* Content */}
      <div className="p-4 space-y-4">
        {/* Fields section */}
        <div>
          <Skeleton className="h-3 w-16 mb-2" />
          <div className="flex flex-wrap gap-2">
            <Skeleton className="h-8 w-20 rounded-lg" />
            <Skeleton className="h-8 w-24 rounded-lg" />
            <Skeleton className="h-8 w-28 rounded-lg" />
            <Skeleton className="h-8 w-16 rounded-lg" />
          </div>
        </div>

        {/* Endpoints section */}
        <div>
          <Skeleton className="h-3 w-20 mb-2" />
          <div className="space-y-2">
            <EndpointSkeleton />
            <EndpointSkeleton />
          </div>
        </div>
      </div>
    </div>
  )
}

/**
 * Skeleton for Endpoint item
 */
export const EndpointSkeleton = () => {
  return (
    <div className="flex items-center justify-between p-3 rounded-lg border border-gray-200">
      <div className="flex items-center gap-3">
        <Skeleton className="h-7 w-14 rounded" />
        <div className="space-y-1.5">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-3 w-32" />
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Skeleton className="h-7 w-14 rounded-lg" />
        <Skeleton className="w-7 h-7 rounded" />
        <Skeleton className="w-7 h-7 rounded" />
      </div>
    </div>
  )
}

/**
 * Skeleton for Stats Card
 */
export const StatsCardSkeleton = () => {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4">
      <div className="flex items-center gap-3">
        <Skeleton className="w-12 h-12 rounded-lg" />
        <div className="space-y-2">
          <Skeleton className="h-7 w-12" />
          <Skeleton className="h-4 w-16" />
        </div>
      </div>
    </div>
  )
}

/**
 * Skeleton for Deployment History item
 */
export const DeploymentHistorySkeleton = () => {
  return (
    <div className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
      <div className="flex items-center gap-3">
        <Skeleton className="w-10 h-10 rounded-lg" />
        <div className="space-y-2">
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-3 w-32" />
        </div>
      </div>
      <div className="flex items-center gap-3">
        <Skeleton className="h-6 w-16 rounded-full" />
        <Skeleton className="w-8 h-8 rounded" />
      </div>
    </div>
  )
}

/**
 * Skeleton for Snapshot item
 */
export const SnapshotSkeleton = () => {
  return (
    <div className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
      <div className="flex items-center gap-3">
        <Skeleton className="w-10 h-10 rounded-lg" />
        <div className="space-y-2">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-3 w-40" />
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Skeleton className="h-8 w-20 rounded-lg" />
      </div>
    </div>
  )
}

/**
 * Loading overlay with spinner and optional message
 */
export const LoadingOverlay = ({ message = 'Loading...' }) => {
  return (
    <div className="absolute inset-0 bg-white/80 backdrop-blur-sm flex items-center justify-center z-10 rounded-xl">
      <div className="flex flex-col items-center gap-3">
        <div className="relative">
          <div className="w-10 h-10 border-4 border-gray-200 rounded-full" />
          <div className="absolute inset-0 w-10 h-10 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin" />
        </div>
        <span className="text-sm font-medium text-gray-600">{message}</span>
      </div>
    </div>
  )
}

// Add CSS for shimmer animation
const style = document.createElement('style')
style.textContent = `
  @keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }
`
if (typeof document !== 'undefined') {
  document.head.appendChild(style)
}
