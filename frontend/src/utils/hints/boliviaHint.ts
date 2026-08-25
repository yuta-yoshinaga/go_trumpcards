import type { BoliviaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BoliviaPhase } from '../../types/phases';

/**
 * Returns a Bolivia hint suggesting the right phase action for the human turn.
 * Focuses on the common decisions: draw from the stock, lay down melds
 * (sets or sequences), and discard the highest safe card otherwise.
 */
export function getBoliviaHint(state: BoliviaResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0 || state.currentPlayerIdx !== humanIdx) return null;

  const human = state.players[humanIdx];
  if (!human) return null;

  if (state.phase === BoliviaPhase.DRAW) {
    if (state.discardTop && state.discardPileCount > 0 && human.hasInitMeld && !state.isFrozen) {
      return { targetAction: 'draw', reason: 'hint.takeDiscardPile', confidence: 'moderate' };
    }
    return { targetAction: 'draw', reason: 'hint.drawStock', confidence: 'strong' };
  }

  if (state.phase === BoliviaPhase.MELD) {
    if (!human.hasInitMeld) {
      return { targetAction: 'meld', reason: 'hint.meldInitial', confidence: 'moderate' };
    }
    if (!human.hasBolivia) {
      return { targetAction: 'meld', reason: 'hint.buildBolivia', confidence: 'moderate' };
    }
    return { targetAction: 'meld', reason: 'hint.meldExtend', confidence: 'moderate' };
  }

  if (state.phase === BoliviaPhase.DISCARD) {
    return { targetAction: 'discard', reason: 'hint.discardHighSafe', confidence: 'moderate' };
  }

  return null;
}
