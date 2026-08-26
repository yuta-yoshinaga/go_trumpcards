import type { QuodlibetResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Quodlibet, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain on the response's `hint` field with a
 * bare `reason` identifier (`pick_contract`, `avoid_penalty`, `shed_low`,
 * `pass`, `next_deal`); this adapter re-maps it into the shared HintResult
 * shape for {@link hooks/useGameHint.useGameHint | useGameHint}.
 *
 * **The contract suggestion counts as a hint too.** While the dealer is
 * choosing there is no card to point at, so a null-on-empty-cardIndices check
 * would silently drop the advice that matters most.
 */
export function getQuodlibetHint(state: QuodlibetResponse): HintResult | null {
  const hint = state.hint;
  const reason = hint?.reason ?? (state.hintContract >= 0 ? 'pick_contract' : '');
  if (!reason || reason === 'none') return null;
  return {
    targetAction: reason === 'pick_contract' ? 'select' : 'play',
    reason: `hint.${reason}`,
    confidence: 'moderate',
  };
}
