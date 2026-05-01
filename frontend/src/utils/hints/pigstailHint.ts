import type { PigsTailResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Returns a Pig's Tail frontend hint or null.
 *
 * Pig's Tail is a pure-luck game (no decisions other than "draw"). Hint just
 * nudges the human player on their turn and reminds them after a penalty. */
export function getPigstailHint(state: PigsTailResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human) return null;
  if (state.currentTurn !== human.id) return null;
  if (state.lastPenalty) {
    return { targetAction: 'draw', reason: 'hint.afterPenalty', confidence: 'moderate' };
  }
  return { targetAction: 'draw', reason: 'hint.draw', confidence: 'moderate' };
}
