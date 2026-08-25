import type { CirullaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Cirulla, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain as a hand index plus the table
 * indices to take, with a bare `hintReason` identifier (`capture`, `sweep`,
 * `lay_off`, `next_round`); this adapter re-maps it into the shared HintResult
 * shape for {@link hooks/useGameHint.useGameHint | useGameHint}.
 */
export function getCirullaHint(state: CirullaResponse): HintResult | null {
  if (state.hintHandIdx < 0 || !state.hintReason || state.hintReason === 'none') return null;
  return {
    targetAction: 'play',
    reason: `hint.${state.hintReason}`,
    confidence: 'moderate',
  };
}
