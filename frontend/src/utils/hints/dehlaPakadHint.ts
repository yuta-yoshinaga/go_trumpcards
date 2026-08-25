import type { DehlaPakadResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Dehla Pakad, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain on the response's `hint` field with a
 * bare `reason` identifier (`call_longest`, `take_the_ten`, `keep_the_lead`,
 * `next_hand`); this adapter re-maps it into the shared HintResult shape for
 * {@link hooks/useGameHint.useGameHint | useGameHint}.
 *
 * **The trump suggestion counts as a hint too.** While the trump is being
 * called there is no card to point at, so a null-on-empty-cardIndices check
 * would drop the advice at the one moment the whole hand is being decided.
 */
export function getDehlaPakadHint(state: DehlaPakadResponse): HintResult | null {
  const hint = state.hint;
  const reason = hint?.reason ?? (state.hintTrumpSuit > 0 ? 'call_longest' : '');
  if (!reason || reason === 'none') return null;
  return {
    targetAction: reason === 'call_longest' ? 'select' : 'play',
    reason: `hint.${reason}`,
    confidence: 'moderate',
  };
}
