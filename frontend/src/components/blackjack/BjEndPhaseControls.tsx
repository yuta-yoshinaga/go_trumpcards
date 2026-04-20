import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { btnPrimary } from '../../styles/buttonStyles';

/** Props for BlackJack end phase controls. */
export interface BjEndPhaseControlsProps {
  loading: boolean;
  onReset: () => void;
  autoAdvanceSeconds?: number;
}

/**
 * Renders the "Next Game" button with optional auto-advance countdown for BlackJack end phase.
 * No confirmation dialog — the hand is already resolved, so clicking just deals the next one.
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
      onClick={props.onReset}
    >
      {t('button.nextGame')}
      {countdown !== null ? ` (${countdown}s)` : ''}
    </button>
  );
}
