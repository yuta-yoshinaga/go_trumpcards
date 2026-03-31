import type { HoldemResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HoldemPhase } from '../../types/phases';

/** Hand rank thresholds for hint logic. */
const RANK_ONE_PAIR = 1;
const RANK_THREE_OF_A_KIND = 3;

/** Win probability margin above pot odds to recommend a raise. */
const EV_RAISE_MARGIN = 0.1;

/** Returns a frontend HintResult for Hold'em-family games, or null if no suggestion available. */
export function getHoldemBaseHint(state: HoldemResponse): HintResult | null {
  if (
    state.phase === HoldemPhase.INIT ||
    state.phase === HoldemPhase.SHOWDOWN ||
    state.phase === HoldemPhase.END ||
    state.phase === HoldemPhase.REBUY
  ) {
    return null;
  }

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn) return null;

  if (state.equity?.winProbability != null && state.potOdds != null) {
    return getEquityHint(state.equity.winProbability, state.potOdds);
  }

  return getHandRankHint(human.handRank);
}

/** Hint based on equity vs pot odds comparison. */
function getEquityHint(winProbability: number, potOdds: number): HintResult {
  if (winProbability > potOdds + EV_RAISE_MARGIN) {
    return { targetAction: 'raise', reason: 'hint.positiveEV', confidence: 'strong' };
  }
  if (winProbability >= potOdds) {
    return { targetAction: 'call', reason: 'hint.marginalEV', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'hint.negativeEV', confidence: 'moderate' };
}

/** Fallback hint based on hand rank when equity data is unavailable. */
function getHandRankHint(handRank: number): HintResult {
  if (handRank >= RANK_THREE_OF_A_KIND) {
    return { targetAction: 'raise', reason: 'hint.strongHandRank', confidence: 'strong' };
  }
  if (handRank >= RANK_ONE_PAIR) {
    return { targetAction: 'call', reason: 'hint.decentHandRank', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'hint.weakHandRank', confidence: 'moderate' };
}
