import type { SutdaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Sutda, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain as an action (`call` / `raise` /
 * `fold`) plus a bare `hintReason` identifier; this adapter re-maps it into the
 * shared HintResult shape for {@link hooks/useGameHint.useGameHint | useGameHint}.
 *
 * **There is no card to point at.** Every Sutda decision is a bet, so the
 * suggestion rides on the action rather than on a hand index.
 */
export function getSutdaHint(state: SutdaResponse): HintResult | null {
  if (!state.hintAction || !state.hintReason || state.hintReason === 'none') return null;
  return {
    targetAction: state.hintAction,
    reason: `hint.${state.hintReason}`,
    confidence: 'moderate',
  };
}
