import type { SoloWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Solo Whist, or null when no
 * suggestion is available.
 *
 * Like Mariáš, Solo Whist's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix). This
 * adapter re-maps that server hint into the frontend HintResult shape so the
 * shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is fixed
 * to `play` because the hint only applies during the trick-play phase.
 */
export function getSoloWhistHint(state: SoloWhistResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
