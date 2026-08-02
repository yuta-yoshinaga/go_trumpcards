import type { PaiGowResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PaiGowPhase } from '../../types/phases';
import { paiGowAutoSplit } from '../paiGowFoul';

/**
 * Returns a frontend {@link HintResult} for Pai Gow Poker, or null when no
 * suggestion is available.
 *
 * The page renders only `reason` — there is no `data-hint-action` wiring — so the
 * advice has to read on its own.
 *
 * The set-hands advice does **not** re-derive a split. The page already computes
 * one with {@link paiGowAutoSplit}, and it is bound to a button ("A"), so the
 * useful thing to say is to press it. Computing a second split here would mean
 * two house-way implementations that can disagree with each other, and the one
 * the player can actually apply with one keystroke is the page's.
 *
 * `paiGowAutoSplit` returns null on a hand holding the joker, because foul
 * evaluation is unavailable there. That is exactly when the auto button is
 * disabled, so the hint switches to stating the rule the player has to satisfy
 * by hand rather than pointing at a control that does nothing.
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

  return paiGowAutoSplit(state.playerCards) === null
    ? { targetAction: 'setHands', reason: 'frontendHint.paigowSplitByHand', confidence: 'moderate' }
    : { targetAction: 'autoSet', reason: 'frontendHint.paigowAutoSplit', confidence: 'moderate' };
}
