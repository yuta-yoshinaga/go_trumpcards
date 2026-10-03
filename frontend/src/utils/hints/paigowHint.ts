import type { PaiGowResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PaiGowPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Pai Gow Poker, or null when no
 * suggestion is available.
 *
 * The page renders only `reason` — there is no `data-hint-action` wiring — so the
 * advice has to read on its own.
 *
 * The set-hands advice follows the server's house-way hint used by the auto button.
 *
 * **The staged selection is not in the response.** `selectedIndices` is page
 * state until `set` is sent, so this cannot comment on the split in progress —
 * only on the seven cards that were dealt.
 */
export function getPaiGowHint(state: PaiGowResponse): HintResult | null {
  if (state.phase === PaiGowPhase.BET) {
    return state.chips <= 0 ? null : { targetAction: 'bet', reason: 'frontendHint.paigowBet', confidence: 'moderate' };
  }

  if (state.phase !== PaiGowPhase.SET_HANDS) return null;

  return state.hint == null
    ? { targetAction: 'setHands', reason: 'frontendHint.paigowSplitByHand', confidence: 'moderate' }
    : { targetAction: 'autoSet', reason: 'frontendHint.paigowAutoSplit', confidence: 'moderate' };
}
