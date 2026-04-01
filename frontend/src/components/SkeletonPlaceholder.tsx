/**
 * SkeletonPlaceholder — reusable skeleton shimmer block for login and
 * protected-page loading states.
 *
 * Renders within 300ms (no async boundary needed) via CSS animation.
 */

interface SkeletonPlaceholderProps {
  /** Tailwind height class, e.g. "h-10" */
  height?: string
  /** Tailwind width class, e.g. "w-full" */
  width?: string
  /** Tailwind border-radius class, e.g. "rounded-md" */
  rounded?: string
  className?: string
}

export function SkeletonPlaceholder({
  height = 'h-10',
  width = 'w-full',
  rounded = 'rounded-md',
  className = '',
}: SkeletonPlaceholderProps) {
  return (
    <div
      role="status"
      aria-label="Loading"
      className={`animate-pulse bg-loading-skeleton ${height} ${width} ${rounded} ${className}`}
    />
  )
}
