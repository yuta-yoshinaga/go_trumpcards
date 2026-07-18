import type { Card, ContractRummyContractSlot } from '../types/card';
import { CONTRACT_SLOT_SET, type ContractSlotEvaluation } from './contractRummyUtils';

/** Reason a Carioca contract slot is not yet satisfied. Drives the "what's missing" annotation. */
export type CariocaSlotShortfallCode =
  | 'empty'
  | 'needMoreSet'
  | 'needMoreRun'
  | 'tooMany'
  | 'setRankMismatch'
  | 'runSuitMismatch'
  | 'runNotConsecutive';

/** A single unmet contract-slot requirement: a reason code plus an optional card count. */
export interface CariocaSlotShortfall {
  /** i18n key suffix describing the shortfall. */
  code: CariocaSlotShortfallCode;
  /** Cards still needed (`needMore*`) or excess cards to remove (`tooMany`); absent for combination errors. */
  count?: number;
}

/**
 * Joker-aware contract-slot helpers for Carioca.
 *
 * Carioca uses a 108-card deck (2 standard decks + 4 jokers) where jokers are wild
 * (max 1 per meld). These mirror the Go domain's `cariocaIsSet` / `cariocaIsRun` so the
 * human can submit a joker-wild contract meld that the backend `cariocaValidateContractSlot`
 * accepts. A joker card is identified by `design === 'JOKER'`.
 */

/** Maximum number of jokers permitted inside a single meld. */
const CARIOCA_MAX_JOKERS_PER_MELD = 1;
/** Minimum cards for a Carioca set (trío). */
const CARIOCA_SET_SIZE = 3;
/** Minimum cards for a Carioca run (escala). */
const CARIOCA_RUN_SIZE = 4;

/** Returns true iff the card is a joker (wild). */
function isJoker(card: Card): boolean {
  return card.design === 'JOKER';
}

/**
 * Ace-low and (when an Ace is present) Ace-high value variants, each sorted ascending.
 * Ace (value 1) may act as high (14) so J-Q-K-A runs are recognised.
 */
function aceVariants(values: number[]): number[][] {
  const low = [...values].sort((a, b) => a - b);
  const out: number[][] = [low];
  if (low.length > 0 && low[0] === 1) {
    const high = [...low.slice(1), 14].sort((a, b) => a - b);
    out.push(high);
  }
  return out;
}

/**
 * Returns true iff cards form a valid Carioca set: 3+ cards of the same rank, with at
 * most one joker standing in as a wild.
 */
export function isCariocaSet(cards: Card[]): boolean {
  if (cards.length < CARIOCA_SET_SIZE) return false;
  let jokers = 0;
  let rank: number | null = null;
  for (const c of cards) {
    if (isJoker(c)) {
      jokers++;
      continue;
    }
    if (rank === null) rank = c.value;
    else if (c.value !== rank) return false;
  }
  return jokers <= CARIOCA_MAX_JOKERS_PER_MELD && rank !== null;
}

/**
 * Returns true iff cards form a valid Carioca run: 4+ consecutive cards of the same suit,
 * with at most one joker filling a gap or extending an end. Ace may be low or high; a
 * wrap-around (K-A-2) is not allowed.
 */
export function isCariocaRun(cards: Card[]): boolean {
  if (cards.length < CARIOCA_RUN_SIZE) return false;
  let jokers = 0;
  let suit: Card['design'] | null = null;
  const values: number[] = [];
  const seen = new Set<number>();
  for (const c of cards) {
    if (isJoker(c)) {
      jokers++;
      continue;
    }
    if (suit === null) suit = c.design;
    else if (c.design !== suit) return false;
    if (seen.has(c.value)) return false;
    seen.add(c.value);
    values.push(c.value);
  }
  if (jokers > CARIOCA_MAX_JOKERS_PER_MELD || values.length === 0) return false;
  const total = cards.length;
  // Every real card fits within a `total`-wide consecutive window, so the joker(s) can
  // fill whatever positions are missing.
  for (const variant of aceVariants(values)) {
    if (variant.length === 0) continue;
    const span = variant[variant.length - 1] - variant[0];
    if (span <= total - 1) return true;
  }
  return false;
}

/**
 * Returns true iff `card` can be laid off onto an existing on-table `meldCards` meld.
 * Mirrors the Go domain's `canAddToCariocaMeld`: the meld plus the card must still form a
 * valid set (same rank) or run (same suit, consecutive), with at most one joker across the
 * combined meld. An empty meld never accepts a layoff.
 */
export function canLayoffCariocaMeld(meldCards: Card[], card: Card): boolean {
  if (meldCards.length === 0) return false;
  const combined = [...meldCards, card];
  return isCariocaSet(combined) || isCariocaRun(combined);
}

/**
 * Evaluate a single Carioca contract slot against the cards placed into it, treating a
 * joker as a wildcard. Drop-in replacement for `evaluateContractSlot` (same return shape).
 */
export function evaluateCariocaContractSlot(slot: ContractRummyContractSlot, cards: Card[]): ContractSlotEvaluation {
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
  const ok = slot.kind === CONTRACT_SLOT_SET ? isCariocaSet(cards) : isCariocaRun(cards);
  return { required, placed, satisfied: ok, invalid: !ok };
}

/**
 * Describe what a Carioca contract slot still needs to be satisfied. Returns `null` when the
 * slot already forms a valid set/run at the required size (nothing missing). Otherwise it
 * reports why the slot cannot be submitted yet: not started, too few / too many cards, or a
 * combination error (rank mismatch, suit mismatch, or a broken sequence). Reuses
 * `isCariocaSet` / `isCariocaRun`, so it stays in sync with the Go domain's validation.
 */
export function describeCariocaSlotShortfall(
  slot: ContractRummyContractSlot,
  cards: Card[],
): CariocaSlotShortfall | null {
  const placed = cards.length;
  const required = slot.size;
  const isSet = slot.kind === CONTRACT_SLOT_SET;
  if (placed === 0) return { code: 'empty' };
  if (placed > required) return { code: 'tooMany', count: placed - required };
  if (placed < required) return { code: isSet ? 'needMoreSet' : 'needMoreRun', count: required - placed };
  // placed === required: the count is right, so any failure is a combination error.
  if (isSet ? isCariocaSet(cards) : isCariocaRun(cards)) return null;
  if (isSet) return { code: 'setRankMismatch' };
  const suits = new Set(cards.filter((c) => !isJoker(c)).map((c) => c.design));
  return suits.size > 1 ? { code: 'runSuitMismatch' } : { code: 'runNotConsecutive' };
}
