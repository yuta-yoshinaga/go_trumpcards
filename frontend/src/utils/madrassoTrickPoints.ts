import type { MadrassoResponse } from '../types/card';

/** Returns the integer card points currently on a Madrasso trick. */
export function madrassoTrickPoints(cards: MadrassoResponse['currentTrick']): number {
  return cards.reduce((total, { card }) => {
    switch (card.value) {
      case 1:
        return total + 11;
      case 3:
        return total + 10;
      case 13:
        return total + 4;
      case 12:
        return total + 3;
      case 11:
        return total + 2;
      default:
        return total;
    }
  }, 0);
}

// Card points match internal/domain/Madrasso.go's madrassoPoints mapping.
