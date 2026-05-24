import { useEffect, useRef, useState } from 'react';

/** Outcome of the slap that triggered the burst. */
export type SlapOutcome = 'correct' | 'wrong';

/** Props for the SlapBurst component. */
export interface SlapBurstProps {
  /**
   * Monotonically-increasing trigger key. Whenever this value changes (and is
   * non-zero), the burst re-fires. Pass an incrementing counter from the
   * parent: counters survive back-to-back events within a single
   * millisecond, which `Date.now()` does not.
   */
  triggerKey: number;
  /** Slap outcome — drives the badge tint. */
  outcome: SlapOutcome;
  /** Human-readable label rendered inside the burst (e.g. PAIR! / SANDWICH!). */
  label: string;
  /** Test id forwarded to the rendered root. */
  testId?: string;
}

const BURST_MS = 1200;

/**
 * Comic-style burst overlay shown on top of the center pile when a slap fires.
 * Auto-dismisses after BURST_MS milliseconds. The component is purely visual;
 * the underlying slap result is owned by the game state.
 */
export function SlapBurst({ triggerKey, outcome, label, testId }: SlapBurstProps) {
  const [visible, setVisible] = useState(false);
  const prevKey = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    if (triggerKey === 0 || triggerKey === prevKey.current) return;
    prevKey.current = triggerKey;
    setVisible(true);
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => setVisible(false), BURST_MS);
    return () => clearTimeout(timerRef.current);
  }, [triggerKey]);

  if (!visible) return null;

  const colorClass = outcome === 'correct' ? 'bg-ds-success text-ds-text-on-accent' : 'bg-ds-error text-white';

  return (
    <div
      role="status"
      aria-live="assertive"
      data-testid={testId ?? 'slap-burst'}
      data-outcome={outcome}
      className="pointer-events-none absolute inset-0 flex items-center justify-center"
    >
      <div
        className={`-rotate-6 transform rounded-md px-4 py-2 text-2xl font-extrabold tracking-wider shadow-2xl ring-4 ring-white/40 motion-safe:animate-[pulse-once_0.6s_ease-out] ${colorClass}`}
      >
        <span aria-hidden="true" className="mr-1">
          💥
        </span>
        {label}
      </div>
    </div>
  );
}
