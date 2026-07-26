import type { BouillotteResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Bouillotte, or null when no
 * suggestion is available.
 *
 * Like Three Card Brag and Guts, Bouillotte's hint is computed entirely by the
 * Go backend and surfaced on the response's `hint` field. The hint carries an
 * `action` (`"call"` / `"raise"` / `"fold"`) mapped to the `targetAction`
 * string, and a `reason` i18n suffix (`strong_hand` / `medium_hand` /
 * `weak_hand`) re-mapped into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getBouillotteHint(state: BouillotteResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.action,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
