import type { Card } from '../types/card';

/**
 * Matador rank within Quadrille's trump hierarchy:
 * 1 = Spadille (♠A), 2 = Manille (7 of the trump suit), 3 = Basto (♣A).
 * Lower number = stronger card.
 */
export type MatadorRank = 1 | 2 | 3;

/**
 * Maps a trump-suit code (1=♠ 2=♣ 3=♥ 4=♦) to the {@link Card} `design` value
 * that the trump suit's Manille carries. Codes outside 1..4 (e.g. 0 = none,
 * -1 = unset) are absent, which signals "trump not yet decided".
 */
const TRUMP_DESIGN: Record<number, Card['design']> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/**
 * Identifies whether a card is one of Quadrille's three matadors and returns its
 * rank, or `null` when it is an ordinary card or trump is not yet decided.
 *
 * Mirrors the Go domain's `quadrilleCardStrength` (internal/domain/Quadrille.go)
 * exactly:
 * - Spadille = ♠A, always the top trump regardless of the trump suit.
 * - Manille = the **7 of the trump suit**, always (this codebase uses 7 for
 *   every trump colour, black and red alike — it does not switch to the 2 for
 *   black trumps).
 * - Basto = ♣A, always the third trump regardless of the trump suit.
 *
 * @param card - The card to classify.
 * @param trumpSuit - The trump-suit code (1=♠ 2=♣ 3=♥ 4=♦); values outside
 *   1..4 mean trump is undecided and yield `null`.
 * @returns The matador rank (1/2/3) or `null`.
 */
export function matadorRank(card: Card, trumpSuit: number): MatadorRank | null {
  const trumpDesign = TRUMP_DESIGN[trumpSuit];
  if (trumpDesign === undefined) return null; // trump undecided → no badges
  if (card.design === 'SPADE' && card.value === 1) return 1; // Spadille
  if (card.design === trumpDesign && card.value === 7) return 2; // Manille
  if (card.design === 'CLOVER' && card.value === 1) return 3; // Basto
  return null;
}

/** i18n key (under the `quadrille` namespace) for each matador rank's display name. */
export const MATADOR_NAME_KEY: Record<MatadorRank, string> = {
  1: 'matador.spadille',
  2: 'matador.manille',
  3: 'matador.basto',
};
