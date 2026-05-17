import type { Rummy500Response } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { Rummy500Phase } from '../../types/phases';

/**
 * Heuristic hint for Rummy 500. Returns null when no actionable
 * recommendation is available (e.g. it is not the human's turn or the round
 * has ended).
 *
 * Suggestions:
 * - During the Draw phase, recommend the discard pile when it has cards.
 * - During the Play phase, recommend laying any obvious 3-of-a-kind.
 */
export function getRummy500Hint(state: Rummy500Response): HintResult | null {
  if (state.gameEndFlag) return null;
  const me = state.players.find((p) => p.isHuman);
  if (!me) return null;
  const isMyTurn = state.players[state.currentPlayerIdx]?.isHuman === true;
  if (!isMyTurn) return null;

  if (state.phase === Rummy500Phase.DRAW) {
    if (state.discardPile.length > 0) {
      return { targetAction: 'drawdiscard', reason: 'rummy500.hint.drawDiscardTop', confidence: 'moderate' };
    }
    return { targetAction: 'drawstock', reason: 'rummy500.hint.drawStock', confidence: 'moderate' };
  }

  if (state.phase === Rummy500Phase.PLAY) {
    const ranks: Record<number, number> = {};
    for (const c of me.cards) {
      ranks[c.value] = (ranks[c.value] ?? 0) + 1;
    }
    const hasTriple = Object.values(ranks).some((n) => n >= 3);
    if (hasTriple) {
      return { targetAction: 'meld', reason: 'rummy500.hint.meldSet', confidence: 'strong' };
    }
    return { targetAction: 'discard', reason: 'rummy500.hint.discardHighCard', confidence: 'moderate' };
  }

  return null;
}
