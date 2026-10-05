import type { Rummy500Response } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { Rummy500Phase } from '../../types/phases';
import { cardAlt } from '../cardAlt';
import { rummy500CardPenalty } from '../rummy500HandPenalty';

/**
 * Heuristic hint for Rummy 500. Returns null when no actionable
 * recommendation is available (e.g. it is not the human's turn or the round
 * has ended).
 *
 * Suggestions:
 * - During the Draw phase, recommend the discard pile when it has cards.
 * - During the Play phase, recommend laying any obvious 3-of-a-kind.
 * - Otherwise, name the highest-penalty card with no lay-off target and no meld
 *   partner (same rank, or same suit within two ranks) as the discard; fall back
 *   to the generic high-card advice when no such card exists.
 */
export function getRummy500Hint(state: Rummy500Response): HintResult | null {
  if (state.gameEndFlag) return null;
  const me = state.players.find((p) => p.isHuman);
  if (!me) return null;
  const isMyTurn = state.players[state.currentPlayerIdx]?.isHuman === true;
  if (!isMyTurn) return null;

  if (state.phase === Rummy500Phase.DRAW) {
    if (state.discardPile.length > 0) {
      return { targetAction: 'drawdiscard', reason: 'hint.drawDiscardTop', confidence: 'moderate' };
    }
    return { targetAction: 'drawstock', reason: 'hint.drawStock', confidence: 'moderate' };
  }

  if (state.phase === Rummy500Phase.PLAY) {
    const ranks: Record<number, number> = {};
    for (const c of me.cards) {
      ranks[c.value] = (ranks[c.value] ?? 0) + 1;
    }
    const hasTriple = Object.values(ranks).some((n) => n >= 3);
    if (hasTriple) {
      return { targetAction: 'meld', reason: 'hint.meldSet', confidence: 'strong' };
    }
    const discardCandidates = me.cards
      .map((candidate, index) => {
        if (state.layoffTargets[index].length > 0) return false;
        const hasMeldPartner = me.cards.some((other, otherIndex) => {
          if (otherIndex === index) return false;
          const sameRank = other.value === candidate.value;
          const canExtendRun = other.design === candidate.design && Math.abs(other.value - candidate.value) <= 2;
          return sameRank || canExtendRun;
        });
        return hasMeldPartner ? false : index;
      })
      .filter((index): index is number => index !== false)
      .sort((left, right) => rummy500CardPenalty(me.cards[right].value) - rummy500CardPenalty(me.cards[left].value));
    const discardIndex = discardCandidates[0];
    if (discardIndex >= 0) {
      const card = me.cards[discardIndex];
      return {
        targetAction: 'discard',
        targetPos: discardIndex,
        reason: 'hint.discardCard',
        reasonParams: { card: cardAlt(card) },
        confidence: 'moderate',
      };
    }
    return { targetAction: 'discard', reason: 'hint.discardHighCard', confidence: 'moderate' };
  }

  return null;
}
