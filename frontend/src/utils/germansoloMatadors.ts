import type { Card } from '../types/card';

/**
 * Matador rank within German Solo's trump hierarchy:
 * 1 = Spadille (♣Q), 2 = Manille (7 of the trump suit), 3 = Basta (♠Q).
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

/** Card value of a queen — the rank both fixed matadors carry. */
const QUEEN = 12;

/** Card value of the Manille, the 7 of whichever suit is trumps. */
const MANILLE = 7;

/**
 * Identifies whether a card is one of German Solo's three matadors and returns
 * its rank, or `null` when it is an ordinary card or trump is not yet decided.
 *
 * Mirrors the Go domain's `germanSoloCardStrength`
 * (internal/domain/GermanSolo.go) exactly:
 * - Spadille = **♣Q**, always the top trump regardless of the trump suit.
 * - Manille = the **7 of the trump suit**, always.
 * - Basta = **♠Q**, always the third trump regardless of the trump suit.
 *
 * The two black queens being permanent trumps is what makes this game's ranking
 * differ from `quadrille`'s (which elevates the two black aces instead), so the
 * order of these checks matters: ♣Q is Spadille even when clubs are trumps, and
 * the trump 7 outranks ♠Q.
 *
 * @param card - The card to classify.
 * @param trumpSuit - The trump-suit code (1=♠ 2=♣ 3=♥ 4=♦); values outside
 *   1..4 mean trump is undecided and yield `null`.
 * @returns The matador rank (1/2/3) or `null`.
 */
export function matadorRank(card: Card, trumpSuit: number): MatadorRank | null {
  const trumpDesign = TRUMP_DESIGN[trumpSuit];
  if (trumpDesign === undefined) return null; // trump undecided → no badges
  if (card.design === 'CLOVER' && card.value === QUEEN) return 1; // Spadille
  if (card.design === trumpDesign && card.value === MANILLE) return 2; // Manille
  if (card.design === 'SPADE' && card.value === QUEEN) return 3; // Basta
  return null;
}

/** i18n key (under the `germansolo` namespace) for each matador rank's display name. */
export const MATADOR_NAME_KEY: Record<MatadorRank, string> = {
  1: 'matador.spadille',
  2: 'matador.manille',
  3: 'matador.basta',
};
