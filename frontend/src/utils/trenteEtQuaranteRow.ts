import type { Card } from '../types/card';

/**
 * Lower bound at which a Trente et Quarante row stops being dealt. The dealer
 * keeps adding cards until the running total reaches this value (31–40).
 */
export const TRENTE_ET_QUARANTE_TARGET = 31;

/**
 * Pip value of a single card in Trente et Quarante: A = 1, 2–10 = face value,
 * J/Q/K = 10. Mirrors `trenteEtQuaranteCardValue` in the Go domain.
 */
export function trenteEtQuaranteCardValue(value: number): number {
  return value > 10 ? 10 : value;
}

/** One dealt card paired with the running row total after it was played. */
export interface TrenteEtQuaranteRowStep {
  /** The dealt card. */
  card: Card;
  /** Cumulative pip total of the row up to and including this card. */
  cumulative: number;
  /** True for the card that first pushed the row total to 31 or more. */
  crossing: boolean;
}

/**
 * Build the running-total breakdown for a row of dealt cards. Each step carries
 * the cumulative total after that card, and the first card whose cumulative
 * total reaches {@link TRENTE_ET_QUARANTE_TARGET} is flagged as the crossing
 * card (the point at which the row's total is finalized).
 */
export function buildTrenteEtQuaranteRow(cards: Card[]): TrenteEtQuaranteRowStep[] {
  let cumulative = 0;
  let crossed = false;
  return cards.map((card) => {
    cumulative += trenteEtQuaranteCardValue(card.value);
    const crossing = !crossed && cumulative >= TRENTE_ET_QUARANTE_TARGET;
    if (crossing) crossed = true;
    return { card, cumulative, crossing };
  });
}
