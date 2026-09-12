import type { Card } from '../types/card';

/** Reports whether one card can be added to an existing Contract Rummy meld. */
export function contractRummyCanAddToMeld(meld: readonly Card[], card: Card | null | undefined): boolean {
  if (meld.length === 0 || !card) return false;

  const first = meld[0];
  const isSet = meld.length >= 2 && meld.every((m) => m.value === first.value);
  if (isSet) return card.value === first.value;

  if (card.design !== first.design) return false;
  let minValue = first.value;
  let maxValue = first.value;
  for (const m of meld.slice(1)) {
    minValue = Math.min(minValue, m.value);
    maxValue = Math.max(maxValue, m.value);
  }
  return card.value === minValue - 1 || card.value === maxValue + 1 || (card.value === 1 && maxValue === 13);
}
