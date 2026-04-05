import { useReducedMotion } from '../../hooks/useReducedMotion';

/** Props for the CountdownBar component. */
interface CountdownBarProps {
  /** Seconds remaining in the countdown. */
  remaining: number;
  /** Total countdown duration in seconds. */
  total: number;
  /** Visible label text (e.g. "残り 10 秒"). */
  label?: string;
}

/** Returns the design-system background class for the countdown bar based on remaining seconds. */
function barColorClass(remaining: number): string {
  if (remaining > 6) return 'bg-ds-success';
  if (remaining > 3) return 'bg-ds-warning';
  return 'bg-ds-error';
}

/**
 * Visual countdown progress bar with color-coded urgency and screen-reader support.
 * Renders a horizontal bar that shrinks as time runs out, transitioning green → yellow → red.
 */
export function CountdownBar({ remaining, total, label }: CountdownBarProps) {
  const reduced = useReducedMotion();
  const pct = total > 0 ? (remaining / total) * 100 : 0;

  return (
    <div className="mb-2">
      <div
        role="progressbar"
        aria-valuenow={remaining}
        aria-valuemax={total}
        aria-valuemin={0}
        aria-label={label ?? 'Countdown'}
        className="h-3 rounded-full bg-white/20 overflow-hidden"
      >
        <div
          data-testid="countdown-bar-fill"
          className={`h-full rounded-full ${barColorClass(remaining)}`}
          style={{
            width: `${pct}%`,
            ...(reduced ? {} : { transition: 'width 1s linear' }),
          }}
        />
      </div>
      {label && (
        <div className="text-yellow-300 text-lg font-bold mt-1" aria-live="assertive" aria-atomic="true">
          {label}
        </div>
      )}
    </div>
  );
}
