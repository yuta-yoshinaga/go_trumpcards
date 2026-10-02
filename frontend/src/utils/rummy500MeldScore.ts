import type { Card } from '../types/card';
import { rummy500CardPenalty } from './rummy500HandPenalty';

/**
 * メルドの得点を計算する。internal/domain/Rummy500.go の Rummy500MeldScore (セットは A=1・J/Q/K=10、A と K を含むランでは A=15) と同じ規則。
 */
export function rummy500MeldScore(cards: Card[]): number {
  const isSet = cards.every((card) => card.value === cards[0]?.value);
  const hasAce = cards.some((card) => card.value === 1);
  const hasKing = cards.some((card) => card.value === 13);

  return cards.reduce(
    (total, card) => total + (isSet || card.value !== 1 || !hasAce || !hasKing ? rummy500CardPenalty(card.value) : 15),
    0,
  );
}
