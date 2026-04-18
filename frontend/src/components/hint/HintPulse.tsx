import type { ReactNode } from 'react';

/** Props for the HintPulse wrapper component. */
export interface HintPulseProps {
  /** Whether the pulse animation is active. */
  active: boolean;
  /** Whether the user prefers reduced motion. */
  reducedMotion: boolean;
  /** Child elements to wrap. */
  children: ReactNode;
}

/** Wraps children with a pulse animation or static icon when a hint is active. */
export function HintPulse({ active, reducedMotion, children }: HintPulseProps) {
  if (!active) return <>{children}</>;

  if (reducedMotion) {
    return (
      <span className="relative inline-flex" data-testid="hint-pulse">
        {children}
        <span className="absolute -top-1 -right-1 text-ds-accent text-xs" data-testid="hint-icon">
          💡
        </span>
      </span>
    );
  }

  return (
    <span className="relative inline-flex" data-testid="hint-pulse">
      {children}
      <span
        className="absolute inset-0 rounded animate-pulse ring-2 ring-ds-accent pointer-events-none"
        data-testid="hint-ring"
      />
    </span>
  );
}
