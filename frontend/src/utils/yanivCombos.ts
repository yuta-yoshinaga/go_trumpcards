import type { Card } from '../types/card';

/**
 * Yaniv discard category. Mirrors the Go domain `YanivValidCombo`:
 * a legal discard is a single card, a same-rank set of 2+, or a same-suit
 * run of 3+ consecutive ranks. Anything else is `invalid`.
 */
export type YanivDiscardKind = 'single' | 'set' | 'run' | 'invalid';

/** Result of classifying a Yaniv discard selection. */
export interface YanivDiscardResult {
  /** The category the selection falls into. */
  kind: YanivDiscardKind;
  /**
   * i18n key (under the `yaniv` namespace) explaining why the selection is not
   * a legal discard. Present only when `kind === 'invalid'`.
   */
  reasonKey?: string;
}

/** True when the card is a joker (jokers may only be discarded on their own). */
function isJoker(card: Card): boolean {
  return card.design === 'JOKER';
}

/**
 * Same-rank set: every card shares the first card's value and none is a joker.
 * Mirrors the domain `yanivIsSameValueSet` (jokers cannot form a value set).
 */
function isSameValueSet(cards: Card[]): boolean {
  const v = cards[0].value;
  return cards.every((c) => !isJoker(c) && c.value === v);
}

/**
 * Same-suit run of 3+ consecutive ranks with no joker. Mirrors the domain
 * `yanivIsRun` (all cards share the first card's suit; jokers are never wild).
 */
function isRun(cards: Card[]): boolean {
  if (cards.length < 3) return false;
  const suit = cards[0].design;
  if (suit === 'JOKER') return false;
  if (!cards.every((c) => c.design === suit)) return false;
  const values = cards.map((c) => c.value).sort((a, b) => a - b);
  for (let i = 1; i < values.length; i++) {
    if (values[i] !== values[i - 1] + 1) return false;
  }
  return true;
}

/**
 * Classifies a selection of cards into its Yaniv discard type, mirroring the Go
 * domain `YanivValidCombo`. When the selection is not a legal discard, the
 * result carries an i18n `reasonKey` describing the problem.
 */
export function classifyYanivDiscard(cards: Card[]): YanivDiscardResult {
  if (cards.length === 0) return { kind: 'invalid', reasonKey: 'discardWarn.empty' };
  if (cards.length === 1) return { kind: 'single' };
  if (isSameValueSet(cards)) return { kind: 'set' };
  if (isRun(cards)) return { kind: 'run' };
  if (cards.some(isJoker)) return { kind: 'invalid', reasonKey: 'discardWarn.joker' };
  if (cards.length === 2) return { kind: 'invalid', reasonKey: 'discardWarn.pair' };
  return { kind: 'invalid', reasonKey: 'discardWarn.general' };
}

/** True when the selected cards form any legal Yaniv discard. */
export function isValidYanivDiscard(cards: Card[]): boolean {
  return classifyYanivDiscard(cards).kind !== 'invalid';
}
