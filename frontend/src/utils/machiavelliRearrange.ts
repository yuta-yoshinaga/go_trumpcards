import type { Card } from '../types/card';

/**
 * Pure client-side mirrors of the Machiavelli meld-validity and table
 * conservation rules from `internal/domain/Machiavelli.go`. Used by the
 * rearrange preview UI to decide, before submitting the `play` power move,
 * whether a proposed table is a legal rearrangement. The backend re-validates
 * on submit; these helpers only gate the preview so the button is enabled
 * exactly when the domain would accept the play.
 */

/** Minimum cards in a Machiavelli meld (set or run). */
export const MACHIAVELLI_MELD_MIN = 3;

/** Suit-name → numeric design index, matching the Go `CardDesign*` constants. */
const SUIT_TO_NUM: Record<string, number> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/** Numeric design index for a card ('SPADE'→1 … 'DIAMOND'→4, 'JOKER'→0). */
export function designToNum(design: string): number {
  return SUIT_TO_NUM[design] ?? 0;
}

/** Multiset key for a card: `design*100 + value`, matching Go `machiavelliCardKey`. */
function cardKey(card: Card): number {
  return designToNum(card.design) * 100 + card.value;
}

/**
 * True if `cards` form a valid set: 3+ cards of the same rank with distinct
 * suits and no joker. Mirrors Go `machiavelliIsSet`.
 */
export function isMachiavelliSet(cards: Card[]): boolean {
  if (cards.length < MACHIAVELLI_MELD_MIN) return false;
  let rank = -1;
  const seenSuit = new Set<string>();
  for (const c of cards) {
    if (c.design === 'JOKER') return false;
    if (rank === -1) rank = c.value;
    else if (c.value !== rank) return false;
    if (seenSuit.has(c.design)) return false;
    seenSuit.add(c.design);
  }
  return true;
}

/** Ace-low and (when an ace is present) ace-high sorted variants. Mirrors Go `aceVariants`. */
function aceVariants(values: number[]): number[][] {
  const low = [...values].sort((a, b) => a - b);
  const out = [low];
  if (low.length > 0 && low[0] === 1) {
    const high = [...low.slice(1), 14].sort((a, b) => a - b);
    out.push(high);
  }
  return out;
}

/** True if a sorted list increases by exactly 1 each step. Mirrors Go `isConsecutive`. */
function isConsecutive(values: number[]): boolean {
  for (let i = 1; i < values.length; i++) {
    if (values[i] !== values[i - 1] + 1) return false;
  }
  return true;
}

/**
 * True if `cards` form a valid run: 3+ same-suit cards with distinct,
 * consecutive values (ace low or high, no wrap) and no joker. Mirrors Go
 * `machiavelliIsRun`.
 */
export function isMachiavelliRun(cards: Card[]): boolean {
  if (cards.length < MACHIAVELLI_MELD_MIN) return false;
  let suit = '';
  const seen = new Set<number>();
  const values: number[] = [];
  for (const c of cards) {
    if (c.design === 'JOKER') return false;
    if (suit === '') suit = c.design;
    else if (c.design !== suit) return false;
    if (seen.has(c.value)) return false;
    seen.add(c.value);
    values.push(c.value);
  }
  return aceVariants(values).some(isConsecutive);
}

/** True if `cards` form a valid meld (set or run of 3+). Mirrors Go `machiavelliIsValidMeld`. */
export function isMachiavelliValidMeld(cards: Card[]): boolean {
  return isMachiavelliSet(cards) || isMachiavelliRun(cards);
}

/** Multiset equality of two card lists by (design, value). */
function sameMultiset(a: Card[], b: Card[]): boolean {
  if (a.length !== b.length) return false;
  const counts = new Map<number, number>();
  for (const c of a) counts.set(cardKey(c), (counts.get(cardKey(c)) ?? 0) + 1);
  for (const c of b) {
    const k = cardKey(c);
    const n = counts.get(k);
    if (!n) return false;
    counts.set(k, n - 1);
  }
  return true;
}

/**
 * True if `proposed` conserves the cards: its multiset equals the old table's
 * cards plus the `played` hand cards. Mirrors Go `machiavelliConserves`.
 */
export function machiavelliConserves(oldTable: Card[][], played: Card[], proposed: Card[][]): boolean {
  return sameMultiset([...oldTable.flat(), ...played], proposed.flat());
}

/** Result of evaluating a staged rearrangement against the domain rules. */
export interface RearrangeEvaluation {
  /** Per-group validity flags, in group order (empty groups are dropped). */
  groupValidity: boolean[];
  /** Whether every non-empty group is a valid meld. */
  allMeldsValid: boolean;
  /** Whether the proposed table conserves exactly the old table + played cards. */
  conserves: boolean;
  /** Whether at least one hand card is played (Machiavelli requires ≥1). */
  playsFromHand: boolean;
  /** Whether the rearrangement is legal and can be submitted. */
  canSubmit: boolean;
}

/**
 * Evaluate a staged rearrangement: `groups` are the proposed table melds (empty
 * groups ignored), `oldTable` the current table melds, and `played` the hand
 * cards being contributed. Returns per-group validity and an overall submit
 * gate mirroring the three domain checks (valid melds, conservation, ≥1 played).
 */
export function evaluateRearrange(groups: Card[][], oldTable: Card[][], played: Card[]): RearrangeEvaluation {
  const nonEmpty = groups.filter((g) => g.length > 0);
  const groupValidity = nonEmpty.map(isMachiavelliValidMeld);
  const allMeldsValid = nonEmpty.length > 0 && groupValidity.every(Boolean);
  const conserves = machiavelliConserves(oldTable, played, nonEmpty);
  const playsFromHand = played.length >= 1;
  return {
    groupValidity,
    allMeldsValid,
    conserves,
    playsFromHand,
    canSubmit: allMeldsValid && conserves && playsFromHand,
  };
}
