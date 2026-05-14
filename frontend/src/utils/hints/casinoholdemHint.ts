import type { CasinoHoldemResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CasinoHoldemPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/**
 * Returns a Casino Hold'em hint for the FLOP (call/fold) decision.
 *
 * Casino Hold'em basic strategy: with ~82% positive EV on calls, the simplest
 * useful rule is "call any pair or better", and "call any flush draw" or
 * "call any open-ended straight draw". This implementation simplifies further
 * to: any pair-or-better → strong call recommendation; otherwise a moderate
 * fold suggestion. The backend hand rank already reflects the 5-card eval of
 * (hole + flop), so we lean on it directly instead of re-evaluating cards.
 */
export function getCasinoHoldemHint(state: CasinoHoldemResponse): HintResult | null {
  if (state.phase !== CasinoHoldemPhase.FLOP) return null;
  if (!state.playerHand || state.playerHand.length === 0) return null;

  if (state.playerHandRank >= RANK_ONE_PAIR) {
    return { targetAction: 'call', reason: 'hint.pairOrBetter', confidence: 'strong' };
  }

  return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
}
