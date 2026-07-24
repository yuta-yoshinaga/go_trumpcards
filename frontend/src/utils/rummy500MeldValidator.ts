import type { Card } from '../types/card';

/**
 * The kind of a valid Rummy 500 meld.
 *
 * - `set`: 3 or 4 cards of the same rank, each a distinct suit.
 * - `run`: 3+ cards of the same suit in consecutive rank order.
 */
export type Rummy500MeldKind = 'set' | 'run';

/** Result of classifying a selection of cards against the Rummy 500 meld rules. */
export interface Rummy500MeldResult {
  /** Whether the selection forms a valid set or run. */
  valid: boolean;
  /** The meld kind when valid; `null` when the selection is not a valid meld. */
  kind: Rummy500MeldKind | null;
}

/**
 * Whether a sorted list of integers increases by exactly 1 at each step.
 * Mirrors the backend `isConsecutive` helper.
 */
function isConsecutive(values: number[]): boolean {
  for (let i = 1; i < values.length; i++) {
    if (values[i] !== values[i - 1] + 1) return false;
  }
  return true;
}

/**
 * A valid set: 3 or 4 cards of the same rank, each a distinct suit.
 * Mirrors the backend `rummy500IsValidSet`.
 */
function isValidSet(cards: Card[]): boolean {
  if (cards.length < 3 || cards.length > 4) return false;
  const rank = cards[0].value;
  const seenSuits = new Set<string>();
  for (const c of cards) {
    if (c.value !== rank) return false;
    if (seenSuits.has(c.design)) return false;
    seenSuits.add(c.design);
  }
  return true;
}

/**
 * A valid run: 3+ cards of the same suit with distinct, consecutive ranks.
 * The Ace (value 1) may be low (A-2-3) or high (Q-K-A, treated as 14).
 * Mirrors the backend `rummy500IsValidRun`.
 */
function isValidRun(cards: Card[]): boolean {
  if (cards.length < 3) return false;
  const suit = cards[0].design;
  if (cards.some((c) => c.design !== suit)) return false;

  const values = cards.map((c) => c.value).sort((a, b) => a - b);
  // Duplicates are not allowed in a run.
  for (let i = 1; i < values.length; i++) {
    if (values[i] === values[i - 1]) return false;
  }
  if (isConsecutive(values)) return true;
  // Re-evaluate with the Ace as high (14) when one is present.
  if (values[0] === 1) {
    const high = [...values.slice(1), 14].sort((a, b) => a - b);
    if (isConsecutive(high)) return true;
  }
  return false;
}

/**
 * Classifies a card selection against the Rummy 500 meld rules (set or run),
 * mirroring the backend `Rummy500IsValidMeld`. Fewer than 3 cards is always
 * invalid. Used to pre-warn the player before submitting a meld to the server.
 */
export function classifyRummy500Meld(cards: Card[]): Rummy500MeldResult {
  if (cards.length < 3) return { valid: false, kind: null };
  if (isValidSet(cards)) return { valid: true, kind: 'set' };
  if (isValidRun(cards)) return { valid: true, kind: 'run' };
  return { valid: false, kind: null };
}
