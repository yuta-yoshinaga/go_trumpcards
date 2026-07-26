import type { HachiHachiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Hachi-Hachi (八八), or null when no
 * suggestion is available.
 *
 * Hachi-Hachi's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field (with a `reason` i18n suffix such as `capture` or
 * `discard_low`). This adapter re-maps that server hint into the frontend
 * HintResult shape so the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getHachiHachiHint(state: HachiHachiResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
