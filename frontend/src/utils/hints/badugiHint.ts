import type { BadugiResponse, Card, CardDesign } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BadugiPhase } from '../../types/phases';

/** Compute the size of the best Badugi subset for a hand. O(2^4) enumeration. */
function badugiSize(cards: readonly Card[]): number {
  if (cards.length === 0) return 0;
  let best = 0;
  for (let mask = 1; mask < 1 << cards.length; mask++) {
    const suits = new Set<CardDesign>();
    const ranks = new Set<number>();
    let ok = true;
    let size = 0;
    for (let i = 0; i < cards.length; i++) {
      if ((mask & (1 << i)) === 0) continue;
      const c = cards[i];
      if (suits.has(c.design) || ranks.has(c.value)) {
        ok = false;
        break;
      }
      suits.add(c.design);
      ranks.add(c.value);
      size++;
    }
    if (ok && size > best) best = size;
  }
  return best;
}

/**
 * Returns a Badugi hint for the human seat. The hint surfaces three levels of
 * guidance: stand pat on a completed Badugi, discard duplicates when the best
 * subset is under-full, and fold on a clearly weak hand facing a bet.
 */
export function getBadugiHint(state: BadugiResponse): HintResult | null {
  const human = state.players?.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn) return null;
  const phase = state.phase;
  const isBet = phase === BadugiPhase.DEAL || phase === BadugiPhase.BET;
  const isDraw = phase === BadugiPhase.DRAW;
  if (!isBet && !isDraw) return null;
  if (state.currentTurn !== human.id) return null;

  const cards = human.cards ?? [];
  const size = badugiSize(cards);

  if (isDraw) {
    if (size >= 4) return { targetAction: 'stand', reason: 'hint.standPat', confidence: 'strong' };
    return { targetAction: 'exchange', reason: 'hint.dropDup', confidence: 'moderate' };
  }

  if (size >= 4) return { targetAction: 'raise', reason: 'hint.bigBet', confidence: 'strong' };
  if (size === 3) return { targetAction: 'call', reason: 'hint.callWithValue', confidence: 'moderate' };
  const hasBetToCall = state.lastBet > human.currentBet;
  if (hasBetToCall) return { targetAction: 'fold', reason: 'hint.foldWeak', confidence: 'moderate' };
  return { targetAction: 'check', reason: 'hint.foldWeak', confidence: 'moderate' };
}
