import type { HintResult } from '../../types/hint';
import { HintTooltip } from './HintTooltip';

/** Props for {@link FrontendHintTooltip}. */
export interface FrontendHintTooltipProps {
  /** The current frontend hint, or null when none is available. */
  hint: HintResult | null;
  /** Whether the frontend-hint toggle is enabled. */
  enabled: boolean;
  /** Game-namespace translation function used to resolve the hint reason key. */
  t: (key: string, params?: Record<string, string | number>) => string;
}

/**
 * Renders the frontend hint tooltip when hints are enabled and a hint exists.
 *
 * Consolidates the `{enabled && hint && <HintTooltip … />}` block that was
 * duplicated across ~143 game pages (issue #4302); a change to how hints are
 * displayed now lives in one place.
 */
export function FrontendHintTooltip({ hint, enabled, t }: FrontendHintTooltipProps) {
  if (!enabled || !hint) {
    return null;
  }
  return <HintTooltip reason={t(hint.reason, hint.reasonParams)} confidence={hint.confidence} />;
}
