/**
 * HamburgerMenu — three-line mobile/tablet toggle button.
 *
 * Hidden on desktop (≥1024px) via Tailwind. Accessible: aria-label and
 * aria-pressed communicate state to screen readers.
 */

interface HamburgerMenuProps {
  isOpen: boolean
  onToggle: () => void
}

export function HamburgerMenu({ isOpen, onToggle }: HamburgerMenuProps) {
  return (
    <button
      type="button"
      aria-label={isOpen ? 'Close navigation menu' : 'Open navigation menu'}
      aria-pressed={isOpen}
      onClick={onToggle}
      className="
        inline-flex items-center justify-center
        rounded-md p-2
        text-[color:var(--color-hamburger-icon)]
        hover:bg-surface-raised
        focus:outline-none focus:ring-2 focus:ring-[color:var(--color-primary)]
        lg:hidden
      "
    >
      {/* Three-bar icon */}
      <span className="block h-5 w-5 relative" aria-hidden="true">
        <span
          className={`
            absolute left-0 top-0.5 h-0.5 w-full bg-current transition-transform duration-200
            ${isOpen ? 'translate-y-2 rotate-45' : ''}
          `}
        />
        <span
          className={`
            absolute left-0 top-2 h-0.5 w-full bg-current transition-opacity duration-200
            ${isOpen ? 'opacity-0' : ''}
          `}
        />
        <span
          className={`
            absolute left-0 top-3.5 h-0.5 w-full bg-current transition-transform duration-200
            ${isOpen ? '-translate-y-2 -rotate-45' : ''}
          `}
        />
      </span>
    </button>
  )
}
