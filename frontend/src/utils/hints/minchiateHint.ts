import type { MinchiateResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Minchiate, or null when no
 * suggestion is available.
 *
 * The hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field with a `reason` i18n suffix. That matters more here
 * than in most games because the 40-trump ladder is far longer than the usual
 * 21, so "is anything still above this card" is not something the player can
 * eyeball from the hand. This adapter re-maps
 * the server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getMinchiateHint(state: MinchiateResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
