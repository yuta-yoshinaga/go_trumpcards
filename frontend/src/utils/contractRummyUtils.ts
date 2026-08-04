import type { Card, ContractRummyContractSlot } from '../types/card';

/** Contract slot kind enum: 0 = set (same rank), 1 = run (same suit consecutive). */
export const CONTRACT_SLOT_SET = 0;
export const CONTRACT_SLOT_RUN = 1;

/** Result of evaluating a single contract slot against the cards the player has placed into it. */
export interface ContractSlotEvaluation {
  /** Number of cards required. */
  required: number;
  /** Number of cards currently placed. */
  placed: number;
  /** True only when the count matches and the cards form a valid set/run. */
  satisfied: boolean;
  /** True when the cards do not form a valid set/run (wrong combination or too many cards). */
  invalid: boolean;
}

/** Returns true iff cards share the same rank. Contract Rummy requires at least 3 cards. */
export function isContractSet(cards: Card[]): boolean {
  if (cards.length < 3) return false;
  const first = cards[0].value;
  return cards.every((c) => c.value === first);
}

/** Returns true iff cards share a suit and form a consecutive run (Ace may be low or high). */
export function isContractRun(cards: Card[]): boolean {
  if (cards.length < 3) return false;
  const suit = cards[0].design;
  const values: number[] = [];
  const seen = new Set<number>();
  for (const c of cards) {
    if (c.design !== suit) return false;
    if (seen.has(c.value)) return false;
    seen.add(c.value);
    values.push(c.value);
  }
  const sortedLow = [...values].sort((a, b) => a - b);
  if (isStrictlyConsecutive(sortedLow)) return true;
  if (sortedLow[0] === 1) {
    // sortedLow without the leading Ace is already ascending, and 14 is the max possible
    // value, so concatenating preserves order — no resort needed.
    const high = [...sortedLow.slice(1), 14];
    if (isStrictlyConsecutive(high)) return true;
  }
  return false;
}

function isStrictlyConsecutive(sorted: number[]): boolean {
  for (let i = 1; i < sorted.length; i++) {
    if (sorted[i] !== sorted[i - 1] + 1) return false;
  }
  return true;
}

/** Evaluate a single contract slot against the cards the player has placed into it. */
export function evaluateContractSlot(slot: ContractRummyContractSlot, cards: Card[]): ContractSlotEvaluation {
  const placed = cards.length;
  const required = slot.size;
  if (placed === 0) {
    return { required, placed: 0, satisfied: false, invalid: false };
  }
  if (placed < required) {
    return { required, placed, satisfied: false, invalid: false };
  }
  if (placed > required) {
    return { required, placed, satisfied: false, invalid: true };
  }
  const ok = slot.kind === CONTRACT_SLOT_SET ? isContractSet(cards) : isContractRun(cards);
  return { required, placed, satisfied: ok, invalid: !ok };
}

/**
 * Whether the cards form a valid extra meld — a set or a run of at least three.
 * Mirrors `IsContractRummyMeld` in `internal/domain/ContractRummy.go`, which is
 * the same predicate the contract slots are judged by.
 * @param cards - The selected cards.
 * @returns Whether they may be melded.
 */
export function isContractRummyMeld(cards: Card[]): boolean {
  return isContractSet(cards) || isContractRun(cards);
}
