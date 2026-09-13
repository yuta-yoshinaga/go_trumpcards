import type { FiveCardStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getFiveCardStudHint } from './fivecardstudHint';

/**
 * Returns a Soko frontend hint.
 *
 * Soko's betting decisions are Five Card Stud's — the streets, the bring-in and
 * the pot odds are identical — so the stud hint applies unchanged. What differs
 * is the showdown ranking, and the server has already resolved that into
 * `handName`/`handRank` before the page sees it, so there is nothing extra for
 * the frontend to compute. This delegates rather than forking a near-identical
 * copy that would drift.
 */
export function getSokoHint(state: FiveCardStudResponse): HintResult | null {
  const standard = getFiveCardStudHint(state);
  if (!standard) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human) return standard;
  const cards = [...human.holeCards, ...human.doorCards];
  if (hasFourCardHand(cards)) {
    return { targetAction: 'raise', reason: 'frontendHint.sokoRaiseFourCard', confidence: 'strong' };
  }
  return standard;
}

function hasFourCardHand(cards: FiveCardStudResponse['players'][number]['holeCards']): boolean {
  for (let a = 0; a < cards.length; a++) {
    for (let b = a + 1; b < cards.length; b++) {
      for (let c = b + 1; c < cards.length; c++) {
        for (let d = c + 1; d < cards.length; d++) {
          const four = [cards[a], cards[b], cards[c], cards[d]];
          if (new Set(four.map((card) => card.design)).size === 1) return true;
          const ranks = new Set(four.flatMap((card) => (card.value === 1 ? [1, 14] : [card.value])));
          if ([...Array(11)].some((_, i) => [0, 1, 2, 3].every((offset) => ranks.has(i + 1 + offset)))) return true;
        }
      }
    }
  }
  return false;
}
