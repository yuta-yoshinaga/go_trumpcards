import type { Card } from '../types/common';

function cardKey(card: Card): string {
  return `${card.design}:${card.value}:${card.label ?? ''}:${card.glyph ?? ''}`;
}

/** Finds one card whose count increased between two hands. */
export function findDrawnCard(before: Card[], after: Card[]): Card | undefined {
  const counts = new Map<string, number>();
  for (const card of before) {
    const key = cardKey(card);
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }
  for (const card of after) {
    const key = cardKey(card);
    const count = counts.get(key) ?? 0;
    if (count === 0) return card;
    counts.set(key, count - 1);
  }
  return undefined;
}
