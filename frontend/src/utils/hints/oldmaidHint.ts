import type { OldMaidResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Threshold for "many cards" — edges are more meaningful with more cards. */
const MANY_CARDS = 3;

/** Returns a frontend HintResult for Old Maid, or null if no suggestion. */
export function getOldMaidHint(state: OldMaidResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isFinished) return null;
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.currentTurn !== humanIdx) return null;

  const target = state.players[state.nextDrawTargetIdx];
  if (!target) return null;

  if (target.cardCount >= MANY_CARDS) {
    return { targetAction: 'draw', reason: 'hint.drawFromEdge', confidence: 'moderate' };
  }

  return { targetAction: 'draw', reason: 'hint.drawAny', confidence: 'moderate' };
}
