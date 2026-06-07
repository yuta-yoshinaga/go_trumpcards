import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { badgeSuccessColors } from '../styles/badgeStyles';

/** Props for the AutoCompleteReadyBadge component. */
export interface AutoCompleteReadyBadgeProps {
  /** Whether the auto-complete action is currently available. */
  ready: boolean;
  /** Optional test id forwarded to the rendered badge. */
  testId?: string;
}

const SHOW_MS = 4000;

/**
 * Small badge that briefly appears when an auto-complete action becomes available.
 * Renders an aria-live="polite" status so screen readers announce the transition,
 * and auto-dismisses after a few seconds.
 */
export function AutoCompleteReadyBadge({ ready, testId }: AutoCompleteReadyBadgeProps) {
  const { t } = useTranslation('common');
  const [visible, setVisible] = useState(false);
  const prevReady = useRef(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    if (ready && !prevReady.current) {
      setVisible(true);
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setVisible(false), SHOW_MS);
    } else if (!ready) {
      setVisible(false);
      clearTimeout(timerRef.current);
    }
    prevReady.current = ready;
    return () => clearTimeout(timerRef.current);
  }, [ready]);

  if (!visible) return null;

  return (
    <output
      aria-live="polite"
      data-testid={testId ?? 'autocomplete-ready-badge'}
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold animate-pulse ${badgeSuccessColors}`}
    >
      <span aria-hidden="true">✨</span>
      {t('button.autoCompleteReady')}
    </output>
  );
}
