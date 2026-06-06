import type { TichuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Returns a frontend HintResult from Tichu game state, or null if no suggestion. */
export function getTichuHint(state: TichuResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || state.currentTurn !== human.id) return null;

  // Declaration phase: the conservative play is to hold (declare nothing).
  if (state.phase !== 'play') return null;

  if (!state.tableCards || state.tableCards.length === 0) {
    return {
      targetAction: 'play lowest',
      reason: 'hintReason.playLowest',
      confidence: 'moderate',
    };
  }

  // A teammate (same team) controlling the table — let them win the trick.
  const owner = state.players.find((p) => p.id === state.lastPlayIdx);
  if (owner && owner.team === human.team) {
    return {
      targetAction: 'pass',
      reason: 'hintReason.pass',
      confidence: 'strong',
    };
  }

  return {
    targetAction: 'pass',
    reason: 'hintReason.pass',
    confidence: 'moderate',
  };
}
