import type { Card, SevenCardStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SevenCardStudPhase } from '../../types/phases';

const LOW_CARD_MAX = 8;
const STRONG_LOW_CAP = 5;
const DECENT_LOW_CAP = 4;

const BETTING_PHASES: ReadonlySet<number> = new Set([
  SevenCardStudPhase.THIRD_STREET,
  SevenCardStudPhase.FOURTH_STREET,
  SevenCardStudPhase.FIFTH_STREET,
  SevenCardStudPhase.SIXTH_STREET,
  SevenCardStudPhase.SEVENTH_STREET,
]);

function isLow(card: Card): boolean {
  return card.value <= LOW_CARD_MAX;
}

function hasPair(cards: readonly Card[]): boolean {
  const seen = new Set<number>();
  for (const c of cards) {
    if (seen.has(c.value)) return true;
    seen.add(c.value);
  }
  return false;
}

/**
 * Returns a Razz hint based on the human player's visible cards.
 * Razz rewards low hands (A-5 is the nuts); pairs and high cards weaken the hand.
 */
export function getRazzHint(state: SevenCardStudResponse): HintResult | null {
  const human = state.players?.find((p) => p.isHuman);
  if (!human || human.folded) return null;
  if (!BETTING_PHASES.has(state.phase)) return null;
  if (state.currentTurn !== human.id) return null;

  const cards = [...(human.holeCards ?? []), ...(human.doorCards ?? [])];
  if (cards.length === 0) return null;

  if (hasPair(cards)) {
    return { targetAction: 'fold', reason: 'hint.weakHandRank', confidence: 'moderate' };
  }

  const lowCount = cards.filter(isLow).length;
  const totalCards = cards.length;
  if (lowCount >= Math.min(STRONG_LOW_CAP, totalCards)) {
    return { targetAction: 'raise', reason: 'hint.strongHandRank', confidence: 'strong' };
  }
  if (lowCount >= Math.min(DECENT_LOW_CAP, totalCards - 1)) {
    return { targetAction: 'call', reason: 'hint.decentHandRank', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'hint.weakHandRank', confidence: 'moderate' };
}
