import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { useReducedMotion } from '../../hooks/useReducedMotion';

/** Props for {@link LossFeedback}. */
export interface LossFeedbackProps {
  /** Whether to show the loss feedback. */
  show: boolean;
}

/**
 * Renders a loss feedback overlay as an edge vignette using the error color.
 * Includes ARIA live region for screen readers. Reduced motion users see a
 * text banner instead of the animated vignette.
 */
export function LossFeedback({ show }: LossFeedbackProps) {
  const reduced = useReducedMotion();
  const { t } = useTranslation('common');

  if (!show) return null;

  // Reduced motion: show text banner instead of animated vignette
  if (reduced) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="fixed inset-x-0 top-0 z-50 flex justify-center pointer-events-none"
        data-testid="loss-feedback"
      >
        <div className="mt-4 px-4 py-2 rounded-md bg-ds-error text-white text-sm font-medium">
          {t('result.lose', 'You lost.')}
        </div>
      </div>
    );
  }

  return (
    <>
      <motion.div
        className="pointer-events-none fixed inset-0 z-40"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.2, ease: 'easeOut' }}
        style={{
          background:
            'radial-gradient(ellipse at center, transparent 40%, color-mix(in srgb, var(--ds-error, #c95555) 8%, transparent) 100%)',
        }}
        data-testid="loss-feedback"
        aria-hidden="true"
      />
      <div role="status" aria-live="polite" className="sr-only">
        {t('result.lose', 'You lost.')}
      </div>
    </>
  );
}
