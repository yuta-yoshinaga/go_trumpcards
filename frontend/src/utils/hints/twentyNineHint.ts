import type { TwentyNineResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Twenty-Nine (29), or null when no
 * suggestion is available.
 *
 * Like Auction Forty-Fives, Twenty-Nine's hint is computed entirely by the Go
 * backend and surfaced on the response's `hint` field (with a `reason` i18n
 * suffix). This adapter re-maps that server hint into the frontend HintResult
 * shape so the shared {@link useGameHint} tooltip can render it. The
 * `targetAction` is fixed to `play` because the hint only applies during the
 * trick-play phase.
 */
export function getTwentyNineHint(state: TwentyNineResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
