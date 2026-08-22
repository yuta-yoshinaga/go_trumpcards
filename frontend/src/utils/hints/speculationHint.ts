import type { SpeculationResponse } from '../../types/card';
import { SPECULATION_HUMAN_SEAT } from '../../types/games/speculation';
import type { HintResult } from '../../types/hint';
import { SpeculationPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Speculation, or null when there is
 * nothing to advise.
 *
 * **The advice turns on how many cards are still face down, not on how pretty
 * the card is** — the same judgement `speculationAuctionHintKey` makes in
 * SpeculationCuiPresenter.go. A king of trumps with twelve cards still to come
 * is worth selling; the same king with two left is worth keeping, because
 * nothing is likely to beat it before the pot is awarded.
 *
 * Which side of the auction the player is on is decided by `offerTo`: the
 * offer's *recipient* is the card's owner, so `offerTo === 0` means the human
 * is being asked to sell and anything else means the human is being asked to
 * buy. Reading `offerFrom` instead inverts every piece of advice.
 */
export function getSpeculationHint(state: SpeculationResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === SpeculationPhase.FLIP) {
    return { targetAction: 'flip', reason: 'frontendHint.speculationFlip', confidence: 'moderate' };
  }
  if (state.phase !== SpeculationPhase.AUCTION || state.seats.length === 0) return null;

  const remaining = state.seats.reduce((sum, seat) => sum + seat.hiddenCount, 0);
  // 「残りが席数以下」= 各席あと 1 枚以下。ここから先は上を出される目が急に減る。
  const lateInRound = remaining <= state.seats.length;

  if (state.offerTo === SPECULATION_HUMAN_SEAT) {
    return lateInRound
      ? { targetAction: 'decline', reason: 'frontendHint.speculationHold', confidence: 'strong' }
      : { targetAction: 'accept', reason: 'frontendHint.speculationSell', confidence: 'moderate' };
  }
  return lateInRound
    ? { targetAction: 'accept', reason: 'frontendHint.speculationBuy', confidence: 'moderate' }
    : { targetAction: 'decline', reason: 'frontendHint.speculationPass', confidence: 'strong' };
}
