import type { Card } from '../types/card';
import type { KlaverjasTrickCard } from '../types/games/klaverjas';

/** The reason a Klaverjas card cannot be played. */
export type KlaverjasPlayRestriction = 'followSuit' | 'mustTrump' | 'mustOvertrump';

/**
 * Returns the Klaverjas rule that rejects a card, if any.
 *
 * This mirrors `validatePlay` in `internal/domain/Klaverjas.go:307-331`; when
 * changing the rules, update both implementations together.
 */
export function getKlaverjasPlayRestriction(
  hand: readonly Card[],
  currentTrick: readonly KlaverjasTrickCard[],
  trumpSuit: number,
  cardIndex: number,
): KlaverjasPlayRestriction | undefined {
  const card = hand[cardIndex];
  if (!card || currentTrick.length === 0) return undefined;

  const leadSuit = currentTrick[0].card.design;
  const hasLeadSuit = hand.some((handCard) => handCard.design === leadSuit);

  if (hasLeadSuit && card.design !== leadSuit) return 'followSuit';
  if (
    !hasLeadSuit &&
    hand.some((handCard) => handCard.design === trumpSuitName(trumpSuit)) &&
    card.design !== trumpSuitName(trumpSuit)
  ) {
    return 'mustTrump';
  }

  const highestTrump = currentTrick.reduce(
    (highest, trickCard) =>
      trickCard.card.design === trumpSuitName(trumpSuit)
        ? Math.max(highest, trumpStrength(trickCard.card.value))
        : highest,
    -1,
  );
  if (
    highestTrump >= 0 &&
    card.design === trumpSuitName(trumpSuit) &&
    trumpStrength(card.value) <= highestTrump &&
    hand.some(
      (handCard) => handCard.design === trumpSuitName(trumpSuit) && trumpStrength(handCard.value) > highestTrump,
    )
  ) {
    return 'mustOvertrump';
  }
  return undefined;
}

function trumpSuitName(trumpSuit: number): Card['design'] {
  return ['', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'][trumpSuit] as Card['design'];
}

// Keep these values aligned with Klaverjas.trumpStrength in the Go domain.
function trumpStrength(value: number): number {
  switch (value) {
    case 11:
      return 8;
    case 9:
      return 7;
    case 1:
      return 6;
    case 10:
      return 5;
    case 13:
      return 4;
    case 12:
      return 3;
    case 8:
      return 2;
    default:
      return 1;
  }
}
