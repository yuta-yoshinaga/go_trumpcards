import type { Card } from '../types/card';

/**
 * Returns whether a card is a trump in standard (normal-game) Doppelkopf.
 *
 * The trump set mirrors the Go domain (`dkIsTrump` in
 * `internal/domain/Doppelkopf.go`): every Diamond, every Queen (value 12),
 * every Jack (value 11), and the ♥10 (the "Dulle"). All other cards are fail
 * (non-trump) cards.
 *
 * @param card - The card to classify.
 * @returns `true` when the card is a trump, `false` otherwise.
 */
export function isDoppelkopfTrump(card: Card): boolean {
  if (card.design === 'DIAMOND' || card.value === 11 || card.value === 12) {
    return true;
  }
  return card.design === 'HEART' && card.value === 10;
}

/**
 * The Doppelkopf trump ordering, strongest first, as display symbols:
 * ♥10 (Dulle) > ♣Q > ♠Q > ♥Q > ♦Q > ♣J > ♠J > ♥J > ♦J > ♦A > ♦10 > ♦K > ♦9.
 *
 * Suit rank inside Queens/Jacks is ♣ > ♠ > ♥ > ♦ (mirrors
 * `dkTrumpSuitOrder`); the Diamond trumps rank A > 10 > K > 9 (`dkFailRank`).
 */
export const DOPPELKOPF_TRUMP_ORDER: readonly string[] = [
  '♥10',
  '♣Q',
  '♠Q',
  '♥Q',
  '♦Q',
  '♣J',
  '♠J',
  '♥J',
  '♦J',
  '♦A',
  '♦10',
  '♦K',
  '♦9',
];
