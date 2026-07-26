import { useEffect, useRef, useState } from 'react';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { TOAST_DURATION } from '../styles/toastDurations';

/** Props for {@link CpuActionBubble}. */
export interface CpuActionBubbleProps {
  /** Text shown in the bubble. Pass an empty string or undefined to hide. */
  message: string | undefined;
  /**
   * Opaque identifier for the event. The bubble is re-shown (and the dismiss
   * timer is reset) whenever this value changes, even if `message` is
   * identical — e.g. CPU A asks for 5, CPU B asks for 5.
   */
  triggerKey: string | number | undefined;
  /** Milliseconds before auto-dismissing. Defaults to the short toast duration. */
  durationMs?: number;
  /** Extra Tailwind classes for positioning. */
  className?: string;
}

/**
 * Transient floating speech bubble announcing a CPU action (Go Fish ask,
 * Old Maid draw, …). Auto-dismisses after `durationMs` and re-shows when
 * `triggerKey` changes. The bubble is also a `role="status"`
 * `aria-live="polite"` region so screen readers pick up each new event.
 *
 * Honors `prefers-reduced-motion`: animation is suppressed when set.
 */
export function CpuActionBubble({
  message,
  triggerKey,
  durationMs = TOAST_DURATION.short,
  className = '',
}: CpuActionBubbleProps) {
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const reduced = useReducedMotion();

  useEffect(() => {
    // When the upstream event is gone (triggerKey cleared or message emptied),
    // hide the bubble eagerly rather than waiting for the timer — otherwise a
    // mid-visibility prop clear would leave the bubble on screen indefinitely.
    if (triggerKey === undefined || !message) {
      clearTimeout(timerRef.current);
      setVisible(false);
      return;
    }
    setVisible(true);
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => setVisible(false), durationMs);
    return () => clearTimeout(timerRef.current);
  }, [triggerKey, message, durationMs]);

  if (!visible || !message) {
    // Keep an always-mounted sr-only live region so assistive tech can pick up
    // announcements that land within the polite-region throttling window.
    return (
      <div role="status" aria-live="polite" aria-atomic="true" className="sr-only" data-testid="cpu-action-bubble-live">
        {message ?? ''}
      </div>
    );
  }

  const animation = reduced ? '' : 'animate-[slideDown_0.25s_ease-out] ';

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      data-testid="cpu-action-bubble"
      className={`pointer-events-none ${animation}inline-block rounded-full bg-ds-accent px-3 py-1 text-ds-surface text-xs font-bold shadow-lg ${className}`}
    >
      {message}
    </div>
  );
}
