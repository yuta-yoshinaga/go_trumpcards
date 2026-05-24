import { useEffect, useRef, useState } from 'react';

/** Props for the PartnerRevealFlash component. */
export interface PartnerRevealFlashProps {
  /**
   * Whether the partner/adjutant has been revealed in the current round.
   * The flash fires when this transitions from false to true; resetting back
   * to false (e.g., next round) re-arms the trigger.
   */
  revealed: boolean;
  /** Display name of the revealed partner. Used inside the announcement text. */
  partnerName: string;
  /** i18n-localized headline shown above the partner's name. */
  headline: string;
  /** Test id forwarded to the root element. */
  testId?: string;
}

const FLASH_MS = 1800;

/**
 * Renders a one-shot full-screen flash + centered name plate when a hidden
 * partner role is revealed (Napoleon's adjutant, Mighty's partner). Auto-
 * dismisses after FLASH_MS milliseconds.
 */
export function PartnerRevealFlash({ revealed, partnerName, headline, testId }: PartnerRevealFlashProps) {
  const [visible, setVisible] = useState(false);
  const prevRevealed = useRef(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    if (revealed && !prevRevealed.current) {
      setVisible(true);
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setVisible(false), FLASH_MS);
    } else if (!revealed) {
      setVisible(false);
      clearTimeout(timerRef.current);
    }
    prevRevealed.current = revealed;
    return () => clearTimeout(timerRef.current);
  }, [revealed]);

  if (!visible) return null;

  return (
    <div
      role="status"
      aria-live="assertive"
      data-testid={testId ?? 'partner-reveal-flash'}
      className="pointer-events-none fixed inset-0 z-50 flex items-center justify-center bg-black/55 motion-safe:animate-[fadeIn_0.2s_ease-out]"
    >
      <div className="rounded-lg border-2 border-yellow-300 bg-yellow-500/20 px-6 py-4 text-center shadow-2xl motion-safe:animate-pulse">
        <div className="mb-1 text-2xl" aria-hidden="true">
          🛡️
        </div>
        <div className="text-sm font-semibold text-yellow-100">{headline}</div>
        <div className="mt-1 text-xl font-bold text-yellow-50">{partnerName}</div>
      </div>
    </div>
  );
}
