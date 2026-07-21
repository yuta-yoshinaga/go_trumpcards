import type { Card, CardDesign } from '../types/card';

/**
 * Numeric adjutant-suit code the Napoleon trump-declaration action expects for
 * each card design (JOKER→0, SPADE→1, CLOVER→2, HEART→3, DIAMOND→4). Mirrors
 * the Go domain's `CardDesign*` constants (`internal/domain/Card.go`).
 */
export const ADJUTANT_SUIT_BY_DESIGN: Record<CardDesign, number> = {
  JOKER: 0,
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
};

/** Suit designs shown as the four rows of the adjutant card picker, in display order. */
export const ADJUTANT_SUIT_ROW_DESIGNS: readonly CardDesign[] = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

/**
 * The value code required by the Napoleon domain when the JOKER is named as
 * adjutant. The domain rejects any other value for a joker adjutant, so the
 * picker always submits `1` for the joker option.
 */
export const ADJUTANT_JOKER_VALUE = 1;

/** A selectable adjutant candidate: the card face to render plus the numeric (suit, value) the trump-declaration action submits. */
export interface AdjutantCardOption {
  /** The card face rendered in the grid. */
  card: Card;
  /** Adjutant suit code (0=JOKER, 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND). */
  suit: number;
  /** Adjutant value code (1–13; the JOKER uses 1). */
  value: number;
}

/**
 * Builds the adjutant picker rows: four suit rows (A–K each) followed by a
 * final single-card JOKER row. Every option carries the card face plus the
 * numeric (suit, value) the `trump` action expects, so tapping a card
 * designates that exact adjutant.
 */
export function buildAdjutantCardRows(): AdjutantCardOption[][] {
  const suitRows = ADJUTANT_SUIT_ROW_DESIGNS.map((design) =>
    Array.from({ length: 13 }, (_, i) => {
      const value = i + 1;
      return { card: { design, value }, suit: ADJUTANT_SUIT_BY_DESIGN[design], value };
    }),
  );
  const jokerRow: AdjutantCardOption[] = [
    {
      card: { design: 'JOKER', value: ADJUTANT_JOKER_VALUE },
      suit: ADJUTANT_SUIT_BY_DESIGN.JOKER,
      value: ADJUTANT_JOKER_VALUE,
    },
  ];
  return [...suitRows, jokerRow];
}

/**
 * True when the option's card is present in the given hand. Suit cards match by
 * design + value; the joker matches any held joker (its deck value may differ
 * from the picker's canonical `1`). Used to dim cards the human already holds,
 * since naming one makes the human their own adjutant.
 */
export function isAdjutantCardInHand(option: AdjutantCardOption, hand: readonly Card[]): boolean {
  if (option.card.design === 'JOKER') return hand.some((c) => c.design === 'JOKER');
  return hand.some((c) => c.design === option.card.design && c.value === option.value);
}
