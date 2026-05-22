import { useEffect, useState } from 'react';
import { useReducedMotion } from '../hooks/useReducedMotion';

/** Props for the AutoFlipCountdown component. */
export interface AutoFlipCountdownProps {
  /** Total countdown duration in milliseconds. */
  durationMs: number;
  /** Diameter of the indicator in pixels. */
  size?: number;
  /** Accessible label announced to screen readers. */
  ariaLabel: string;
  /**
   * Localised template used to compose the remaining-seconds string.
   * Must contain `{{n}}`; replaced with the integer second count.
   */
  remainingLabel: string;
}

const STROKE_WIDTH_RATIO = 0.12;
const TICK_INTERVAL_MS = 60;

/**
 * Renders a small circular SVG countdown that drains over `durationMs`.
 * The remaining seconds are also rendered in the centre and announced
 * via `aria-live`. Honours `prefers-reduced-motion` by skipping the
 * animation and showing only the numeric countdown.
 */
export function AutoFlipCountdown({ durationMs, size = 56, ariaLabel, remainingLabel }: AutoFlipCountdownProps) {
  const reduced = useReducedMotion();
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    setElapsed(0);
    const id = window.setInterval(() => {
      setElapsed((prev) => {
        const next = prev + TICK_INTERVAL_MS;
        return next >= durationMs ? durationMs : next;
      });
    }, TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [durationMs]);

  const radius = size / 2 - (size * STROKE_WIDTH_RATIO) / 2;
  const circumference = 2 * Math.PI * radius;
  const progress = Math.min(elapsed / durationMs, 1);
  const offset = circumference * progress;
  const remainingSeconds = Math.max(0, Math.ceil((durationMs - elapsed) / 1000));

  return (
    <div
      className="flex flex-col items-center justify-center"
      role="timer"
      aria-live="polite"
      aria-label={ariaLabel}
      data-testid="auto-flip-countdown"
    >
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} aria-hidden="true">
        <title>{ariaLabel}</title>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeOpacity={0.2}
          strokeWidth={size * STROKE_WIDTH_RATIO}
        />
        {!reduced && (
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="currentColor"
            strokeWidth={size * STROKE_WIDTH_RATIO}
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            transform={`rotate(-90 ${size / 2} ${size / 2})`}
            strokeLinecap="round"
          />
        )}
        <text
          x="50%"
          y="50%"
          textAnchor="middle"
          dominantBaseline="central"
          fontSize={size * 0.4}
          fontWeight="bold"
          fill="currentColor"
        >
          {remainingSeconds}
        </text>
      </svg>
      <span className="sr-only">{remainingLabel.replace('{{n}}', String(remainingSeconds))}</span>
    </div>
  );
}
