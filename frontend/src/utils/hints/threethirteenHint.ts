import type { Card, ThreeThirteenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ThreeThirteenPhase } from '../../types/phases';
import { aceHighAdjacent, heaviestSpare, isMaterial } from './rummyHintShape';

/**
 * Returns a frontend {@link HintResult} for Three Thirteen, or null when no
 * suggestion is available.
 *
 * The response carries `deadwood` per player but not the melds behind it, so
 * this does not try to prove a meld. It uses the same shallow
 * "connects with something" test as the other rummies — same rank, or the
 * neighbouring rank in the same suit.
 *
 * The round's `wildRank` is the piece specific to this game: a wild card
 * substitutes for anything, so it is never the card to throw even when nothing
 * in hand happens to touch it.
 */
export function getThreeThirteenHint(state: ThreeThirteenResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === ThreeThirteenPhase.DRAW) {
    const top = state.discardTop;
    return top && (top.value === state.wildRank || isMaterial(top, human.cards, aceHighAdjacent))
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.threethirteenTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.threethirteenDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== ThreeThirteenPhase.DISCARD) return null;

  const idx = heaviestSpare(human.cards, aceHighAdjacent, (c: Card) => c.value === state.wildRank);
  // **捨てられる札が無い。**全部がワイルドか繋がっている手では黙る。
  if (idx < 0) return null;
  return { targetAction: `card-${idx}`, reason: 'frontendHint.threethirteenDiscardHeavy', confidence: 'moderate' };
}
