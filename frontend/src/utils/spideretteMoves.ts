import type { SpideretteTableauCard } from '../types/card';

/**
 * Reports whether the cards from `cardIndex` to the end of a Spiderette
 * tableau column form a movable same-suit descending sequence.
 * Mirrors `Spiderette.isValidSequence` and its face-up source guard.
 */
export function spideretteCanSelectSource(column: readonly SpideretteTableauCard[], cardIndex: number): boolean {
  if (cardIndex < 0 || cardIndex >= column.length) return false;
  const source = column[cardIndex];
  if (!source.faceUp || !source.card) return false;
  for (let i = cardIndex + 1; i < column.length; i++) {
    const previous = column[i - 1];
    const current = column[i];
    if (!current.faceUp || !previous.card || !current.card) return false;
    if (current.card.design !== previous.card.design) return false;
    if (current.card.value !== previous.card.value - 1) return false;
  }
  return true;
}
