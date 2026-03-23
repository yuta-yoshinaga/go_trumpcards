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
  const borderClass = confidence === 'strong' ? 'border-yellow-400' : 'border-yellow-400/50 border-dashed';

  return (
    <div
      className={`text-sm text-yellow-200 bg-gray-800/90 border ${borderClass} rounded px-3 py-1.5 mt-1`}
      role="status"
      aria-live="polite"
      data-testid="hint-tooltip"
    >
      {reason}
    </div>
  );
}
