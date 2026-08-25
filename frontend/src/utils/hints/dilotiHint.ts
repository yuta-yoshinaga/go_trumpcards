import type { DilotiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Diloti, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain as a hand index plus the move to
 * make, with a bare `hintReason` identifier (`capture`, `declare`, `trail`);
 * this adapter re-maps it into the shared HintResult shape for
 * {@link hooks/useGameHint.useGameHint | useGameHint}.
 */
export function getDilotiHint(state: DilotiResponse): HintResult | null {
  if (state.hintHandIdx < 0 || !state.hintReason || state.hintReason === 'none') return null;
  return {
    targetAction: 'play',
    reason: `hint.${state.hintReason}`,
    confidence: 'moderate',
  };
}
