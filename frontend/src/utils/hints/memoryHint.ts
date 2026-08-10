import type { MemoryResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MemoryPhase } from '../../types/phases';

/**
 * Returns a frontend HintResult for Memory, or null if no suggestion.
 *
 * `knownMatchIdx` is the position of a card the player has already seen that
 * matches the one now face up (see `memoryKnownMatch`). When present the advice
 * stops being generic and names the square, and the confidence rises to `strong` — that is real recall the player
 * earned, not information the server leaked (#4775).
 */
export function getMemoryHint(state: MemoryResponse, knownMatchIdx?: number | null): HintResult | null {
  if (state.gameEndFlag) return null;

  const isHumanTurn = state.players[state.currentPlayerIdx]?.isHuman;
  if (!isHumanTurn) return null;

  if (state.phase === MemoryPhase.FLIP1) {
    return { targetAction: 'flip', reason: 'frontendHint.flipAny', confidence: 'moderate' };
  }

  if (state.phase === MemoryPhase.FLIP2) {
    if (knownMatchIdx != null) {
      return {
        targetAction: 'flip',
        reason: 'frontendHint.knownMatch',
        reasonParams: { position: String(knownMatchIdx + 1) },
        confidence: 'strong',
      };
    }
    return { targetAction: 'flip', reason: 'frontendHint.findMatch', confidence: 'moderate' };
  }

  return null;
}
