import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

/** Props for the shared {@link Toast} banner primitive. */
export interface ToastProps {
  /** Toast body. */
  children: ReactNode;
  /**
   * When provided, renders a 44×44px close button that calls this handler.
   * Omit for toasts that are display-only (auto-dismiss with no manual close).
   */
  onDismiss?: () => void;
  /** aria-live politeness for the announcement region. Defaults to 'polite'. */
  live?: 'polite' | 'assertive';
  /** Forwarded to the root element for tests. */
  testId?: string;
}

/**
 * Shared top-of-board notification banner (issue #4313). Standardizes the
 * position, an **opaque** surface background (DESIGN.md Opacity rule — no
 * alpha-suffixed color whose contrast would collapse over the game felt), the
 * `role="status"` live region, and a WCAG-2.5.5 44×44px close button. Entrance
 * animation is handled globally: `animate-[slideDown…]` is near-zeroed by the
 * universal `prefers-reduced-motion` block in index.css, so no per-toast guard
 * is needed. Lifecycle (auto-dismiss, Escape, focus restore) lives in
 * `useTransientToast`; this component is presentational.
 */
export function Toast({ children, onDismiss, live = 'polite', testId }: ToastProps) {
  const { t } = useTranslation('common');
  return (
    <div
      role="status"
      aria-live={live}
      data-testid={testId}
      className={`absolute top-0 left-0 right-0 z-20 mx-4 mt-1 flex min-h-[44px] flex-col justify-center animate-[slideDown_0.3s_ease-out] rounded bg-ds-surface-elevated px-3 py-1.5 text-ds-text-primary text-xs shadow-lg${
        onDismiss ? ' pr-11' : ''
      }`}
    >
      {children}
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          aria-label={t('button.dismiss')}
          className="absolute top-0 right-0 inline-flex min-h-[44px] min-w-[44px] items-center justify-center text-ds-text-muted hover:text-ds-text-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent"
        >
          ×
        </button>
      )}
    </div>
  );
}
