import type { Card, DeuceToSevenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DeuceToSevenPhase } from '../../types/phases';
import { isMadePatLow } from '../deuceToSevenUtils';

/** Card value with the Ace counted as 14 (high), per 2-7 rules. */
function highValue(c: Card): number {
  return c.value === 1 ? 14 : c.value;
}

/**
 * Computes a 1..4 lowball strength rating for a 2-7 hand, mirroring the
 * backend deuceLowStrength scale: 4 = made 8-or-better pat low, 3 = 9-to-jack
 * high no pair, 2 = queen+ high no pair, 1 = paired/straight/flush.
 */
function deuceLowStrength(cards: readonly Card[]): number {
  if (cards.length !== 5) return 1;
  // Any pair → strength 1 (rough proxy for "made hand / drawing").
  const ranks = new Set<number>();
  for (const c of cards) {
    if (ranks.has(c.value)) return 1;
    ranks.add(c.value);
  }
  if (isMadePatLow(cards)) return 4;
  const high = Math.max(...cards.map(highValue));
  if (high <= 11) return 3;
  return 2;
}

/**
 * Returns a 2-7 Triple Draw hint for the human seat. Surfaces stand-pat on a
 * made low, draw guidance on weaker hands, value bets on strong lows, and a
 * fold suggestion on a clearly weak hand facing a bet.
 */
export function getDeuceToSevenHint(state: DeuceToSevenResponse): HintResult | null {
  const human = state.players?.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn) return null;
  const phase = state.phase;
  const isBet = phase === DeuceToSevenPhase.DEAL || phase === DeuceToSevenPhase.BET;
  const isDraw = phase === DeuceToSevenPhase.DRAW;
  if (!isBet && !isDraw) return null;
  if (state.currentTurn !== human.id) return null;

  const cards = human.cards ?? [];
  const strength = deuceLowStrength(cards);

  if (isDraw) {
    if (strength >= 4) return { targetAction: 'stand', reason: 'hint.standPat', confidence: 'strong' };
    return { targetAction: 'exchange', reason: 'hint.dropDup', confidence: 'moderate' };
  }

  if (strength >= 4) return { targetAction: 'raise', reason: 'hint.bigBet', confidence: 'strong' };
  if (strength === 3) return { targetAction: 'call', reason: 'hint.callWithValue', confidence: 'moderate' };
  const hasBetToCall = state.lastBet > human.currentBet;
  if (hasBetToCall) return { targetAction: 'fold', reason: 'hint.foldWeak', confidence: 'moderate' };
  return { targetAction: 'check', reason: 'hint.checkWeak', confidence: 'moderate' };
}
