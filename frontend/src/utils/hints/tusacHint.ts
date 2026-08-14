import type { TuSacResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TuSacPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Tu Sac, or null when there is
 * nothing to advise.
 *
 * **The page cannot see anyone else's hand**, and it deliberately does not
 * re-derive which cards form a combination — the server publishes that through
 * the hint endpoint. Duplicating the meld rules here would give the page a
 * second opinion that can drift from the one that actually scores.
 */
export function getTusacHint(state: TuSacResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === TuSacPhase.ROUND_END) {
    return { targetAction: 'next', reason: 'frontendHint.tusacRoundIsOver', confidence: 'strong' };
  }
  if (!state.isHumanTurn) return null;

  if (state.phase === TuSacPhase.DRAW) {
    return { targetAction: 'draw', reason: 'frontendHint.tusacDrawFirst', confidence: 'strong' };
  }

  const seat = state.seats[state.humanSeat];
  if (!seat) return null;

  // **抱えた枚数がそのまま減点。** 手札が多いほど捨てを急ぐ理由がある。
  if (seat.cards.length > state.handSize) {
    return { targetAction: 'discard', reason: 'frontendHint.tusacDiscardToEndTurn', confidence: 'strong' };
  }
  return { targetAction: 'discard', reason: 'frontendHint.tusacHoldingCostsPoints', confidence: 'moderate' };
}
