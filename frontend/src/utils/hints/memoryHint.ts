import type { MemoryResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MemoryPhase } from '../../types/phases';

/** Returns a frontend HintResult for Memory, or null if no suggestion. */
export function getMemoryHint(state: MemoryResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const isHumanTurn = state.players[state.currentPlayerIdx]?.isHuman;
  if (!isHumanTurn) return null;

  if (state.phase === MemoryPhase.FLIP1) {
    return { targetAction: 'flip', reason: 'frontendHint.flipAny', confidence: 'moderate' };
  }

  if (state.phase === MemoryPhase.FLIP2) {
    return { targetAction: 'flip', reason: 'frontendHint.findMatch', confidence: 'moderate' };
  }

  return null;
}
