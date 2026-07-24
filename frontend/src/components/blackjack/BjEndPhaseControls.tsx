import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { btnPrimary } from '../../styles/buttonStyles';

/** Props for BlackJack end phase controls. */
export interface BjEndPhaseControlsProps {
  loading: boolean;
  /** Raw reset callback. Fired directly by the auto-advance countdown (no confirmation). */
  onReset: () => void;
  /**
   * Handler for a manual button click. Typically opens a reset confirmation dialog before
   * invoking the reset. Falls back to {@link BjEndPhaseControlsProps.onReset} when omitted.
   */
  onRequestReset?: () => void;
  autoAdvanceSeconds?: number;
}

/**
 * Renders the "Next Game" button with optional auto-advance countdown for BlackJack end phase.
 * A manual click routes through `onRequestReset` (reset confirmation dialog) so a stray tap does
 * not discard the session; the auto-advance countdown still fires `onReset` directly.
 */
export function BjEndPhaseControls(props: BjEndPhaseControlsProps) {
  const { t } = useTranslation('common');
  const [countdown, setCountdown] = useState<number | null>(null);
  const onResetRef = useRef(props.onReset);
  onResetRef.current = props.onReset;

  useEffect(() => {
    if (!props.autoAdvanceSeconds || props.autoAdvanceSeconds <= 0) {
      setCountdown(null);
      return;
    }
    setCountdown(props.autoAdvanceSeconds);
    const id = setInterval(() => {
      setCountdown((prev) => {
        if (prev === null) return null;
        const next = prev - 1;
        if (next <= 0) {
          clearInterval(id);
          onResetRef.current();
          return null;
        }
        return next;
      });
    }, 1000);
    return () => clearInterval(id);
  }, [props.autoAdvanceSeconds]);

  return (
    <button
      type="button"
      className={`${btnPrimary} animate-pulse ring-2 ring-white ring-offset-2 ring-offset-green-800`}
      disabled={props.loading}
      onClick={props.onRequestReset ?? props.onReset}
    >
      {t('button.nextGame')}
      {countdown !== null ? ` (${countdown}s)` : ''}
    </button>
  );
}
