import type { DoudizhuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Returns a frontend HintResult from Dou Dizhu game state, or null if no suggestion. */
export function getDoudizhuHint(state: DoudizhuResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || state.currentTurn !== human.id) return null;

  if (state.phase === 'bid') {
    return null;
  }

  if (state.phase !== 'play') return null;

  if (!state.tableCards || state.tableCards.length === 0) {
    return {
      targetAction: 'play lowest',
      reason: 'hintReason.playLowest',
      confidence: 'moderate',
    };
  }

  return {
    targetAction: 'pass',
    reason: 'hintReason.pass',
    confidence: 'moderate',
  };
}
