import type { HintConfidence } from '../../types/hint';

/** Props for the HintTooltip component. */
export interface HintTooltipProps {
  /** The hint reasoning text (already translated). */
  reason: string;
  /** Confidence level of the hint. */
  confidence: HintConfidence;
}

/** Displays a tooltip with the hint reasoning and confidence indicator. */
export function HintTooltip({ reason, confidence }: HintTooltipProps) {
  const borderClass = confidence === 'strong' ? 'border-ds-accent' : 'border-ds-accent/50 border-dashed';

  return (
    <div
      className={`text-sm text-ds-accent bg-ds-surface/90 border ${borderClass} rounded px-3 py-1.5 mt-1`}
      role="status"
      aria-live="polite"
      data-testid="hint-tooltip"
    >
      {reason}
    </div>
  );
}
