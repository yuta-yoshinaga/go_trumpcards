import type { Card, PrsiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PrsiPhase } from '../../types/phases';

/** Value of 7 — stacks a draw-two penalty in Prší. */
const SEVEN = 7;
/** Value of Ace — skips the next player. */
const ACE = 1;
/** Value of the Under (Jack) — skips the next player. */
const UNDER = 11;

/** Returns a frontend HintResult for Prší, or null if no suggestion applies. */
export function getPrsiHint(state: PrsiResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.phase !== PrsiPhase.PLAY) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.currentPlayerIdx !== humanIdx) return null;

  return getPlayHint(human.cards, state);
}

/** Hint for the play phase: stack a 7 under penalty, play a match, or draw. */
function getPlayHint(cards: Card[], state: PrsiResponse): HintResult {
  // Active 7-stack penalty: the only legal move is another 7 or take the stack.
  if (state.penaltyDrawCount > 0) {
    if (cards.some((c) => c.value === SEVEN)) {
      return { targetAction: 'play', reason: 'hint.stackSeven', confidence: 'strong' };
    }
    return { targetAction: 'draw', reason: 'hint.drawPenalty', confidence: 'moderate' };
  }

  const top = state.discardTop;
  if (!top) {
    return { targetAction: 'play', reason: 'hint.playMatchingSuit', confidence: 'moderate' };
  }

  const matches = cards.filter((c) => c.design === top.design || c.value === top.value);
  if (matches.length === 0) {
    return { targetAction: 'draw', reason: 'hint.drawCard', confidence: 'moderate' };
  }

  // Prefer a plain (non-action) card so action cards stay available for tempo.
  const isAction = (c: Card) => c.value === SEVEN || c.value === ACE || c.value === UNDER;
  const plain = matches.find((c) => !isAction(c));
  if (plain) {
    const reason = plain.design === top.design ? 'hint.playMatchingSuit' : 'hint.playMatchingValue';
    return { targetAction: 'play', reason, confidence: 'strong' };
  }

  return { targetAction: 'play', reason: 'hint.playActionCard', confidence: 'strong' };
}
