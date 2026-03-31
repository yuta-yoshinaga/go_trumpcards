import type { IndianPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { IndianPokerPhase } from '../../types/phases';

/** Threshold for high average opponent card rank (suggesting own card may be weak). */
const HIGH_RANK_THRESHOLD = 9;

/** Threshold for low average opponent card rank (suggesting own card may be strong). */
const LOW_RANK_THRESHOLD = 5;

/** Returns a frontend HintResult for Indian Poker, or null if no suggestion available. */
export function getIndianPokerHint(state: IndianPokerResponse): HintResult | null {
  if (state.phase !== IndianPokerPhase.BETTING) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn) return null;

  const opponents = state.players.filter((p) => !p.isHuman && !p.folded);
  if (opponents.length === 0) return null;

  const avgRank = opponents.reduce((sum, p) => sum + p.cardRank, 0) / opponents.length;

  if (avgRank >= HIGH_RANK_THRESHOLD) {
    return { targetAction: 'fold', reason: 'hint.opponentsStrong', confidence: 'moderate' };
  }
  if (avgRank <= LOW_RANK_THRESHOLD) {
    return { targetAction: 'raise', reason: 'hint.opponentsWeak', confidence: 'moderate' };
  }
  return { targetAction: 'call', reason: 'hint.uncertain', confidence: 'moderate' };
}
