import type { Card } from '../types/card';

/**
 * The Tichu play categories, mirroring the domain `TichuComboType`
 * (internal/domain/TichuEval.go). `invalid` covers any selection that does not
 * form a legal combo.
 */
export type TichuComboType =
  | 'single'
  | 'pair'
  | 'triple'
  | 'fullHouse'
  | 'straight'
  | 'stairs'
  | 'bomb'
  | 'straightFlush'
  | 'dog'
  | 'invalid';

/** A classified Tichu selection: its combo type and, for length-bearing combos, the card count. */
export interface TichuComboResult {
  type: TichuComboType;
  /** Card count for straight / stairs / straight-flush / bomb combos; 0 otherwise. */
  length: number;
}

// Tichu special cards use the JOKER design; the value identifies the special.
// Mirrors internal/domain/TichuEval.go (Mahjong=1, Dog=2, Phoenix=3, Dragon=4).
const TICHU_MAHJONG = 1;
const TICHU_DOG = 2;
const TICHU_PHOENIX = 3;
const TICHU_DRAGON = 4;

const TICHU_MAHJONG_RANK = 1;
const TICHU_DRAGON_RANK = 15;

/** Special-card kind for a card (0 for a normal suited card). */
function specialKind(c: Card): number {
  return c.design === 'JOKER' ? c.value : 0;
}

/**
 * Tichu rank of a card used inside combos. Ace (value 1) is high (14); Mahjong=1,
 * Dragon=15, Phoenix/Dog=-1. Mirrors `tichuRank` in TichuEval.go.
 */
function tichuRank(c: Card): number {
  switch (specialKind(c)) {
    case TICHU_MAHJONG:
      return TICHU_MAHJONG_RANK;
    case TICHU_DRAGON:
      return TICHU_DRAGON_RANK;
    case TICHU_PHOENIX:
    case TICHU_DOG:
      return -1;
    default:
      return c.value === 1 ? 14 : c.value;
  }
}

function phoenixCount(cards: readonly Card[]): number {
  return cards.filter((c) => specialKind(c) === TICHU_PHOENIX).length;
}

function nonPhoenix(cards: readonly Card[]): Card[] {
  return cards.filter((c) => specialKind(c) !== TICHU_PHOENIX);
}

/** Rank -> count map over the non-phoenix cards. */
function rankCounts(cards: readonly Card[]): Map<number, number> {
  const m = new Map<number, number>();
  for (const c of cards) {
    const r = tichuRank(c);
    m.set(r, (m.get(r) ?? 0) + 1);
  }
  return m;
}

/** Ascending distinct ranks from a rank-count map. */
function sortedDistinctRanks(counts: Map<number, number>): number[] {
  return [...counts.keys()].sort((a, b) => a - b);
}

const INVALID: TichuComboResult = { type: 'invalid', length: 0 };

function tryPair(cards: readonly Card[], pcount: number): TichuComboResult | null {
  const rest = nonPhoenix(cards);
  if (pcount === 1) {
    // Phoenix + one normal card; cannot pair with the Mahjong (rank 1).
    const r = tichuRank(rest[0]);
    if (r < TICHU_MAHJONG_RANK + 1) return null;
    return { type: 'pair', length: 0 };
  }
  if (tichuRank(rest[0]) === tichuRank(rest[1]) && tichuRank(rest[0]) >= 2) {
    return { type: 'pair', length: 0 };
  }
  return null;
}

function tryTriple(cards: readonly Card[], pcount: number): TichuComboResult | null {
  const rest = nonPhoenix(cards);
  const counts = rankCounts(rest);
  if (pcount === 1) {
    if (counts.size === 1 && tichuRank(rest[0]) >= 2) return { type: 'triple', length: 0 };
    return null;
  }
  if (counts.size === 1 && tichuRank(rest[0]) >= 2) return { type: 'triple', length: 0 };
  return null;
}

function tryBomb4(cards: readonly Card[], pcount: number): TichuComboResult | null {
  if (pcount !== 0 || cards.length !== 4) return null;
  const counts = rankCounts(cards);
  if (counts.size === 1) {
    const r = tichuRank(cards[0]);
    if (r >= 2 && r <= 14) return { type: 'bomb', length: 4 };
  }
  return null;
}

function tryStraightFlush(cards: readonly Card[], pcount: number): TichuComboResult | null {
  if (pcount !== 0 || cards.length < 5) return null;
  let design = '';
  for (const c of cards) {
    if (c.design === 'JOKER') return null; // specials have no suit
    if (design === '') design = c.design;
    else if (c.design !== design) return null;
  }
  const counts = rankCounts(cards);
  const ranks = sortedDistinctRanks(counts);
  if (ranks.length !== cards.length) return null; // duplicate ranks
  if (ranks[ranks.length - 1] - ranks[0] !== ranks.length - 1) return null; // not consecutive
  return { type: 'straightFlush', length: cards.length };
}

function tryStraight(cards: readonly Card[], pcount: number): TichuComboResult | null {
  if (cards.length < 5) return null;
  const rest = nonPhoenix(cards);
  const counts = rankCounts(rest);
  const ranks = sortedDistinctRanks(counts);
  if (ranks.length !== rest.length) return null; // duplicate ranks (a pair snuck in)
  for (const r of ranks) {
    if (r < 1 || r > 14) return null; // Dragon (15) cannot join a straight
  }
  if (pcount === 0) {
    if (ranks[ranks.length - 1] - ranks[0] === ranks.length - 1) {
      return { type: 'straight', length: cards.length };
    }
    return null;
  }
  // One phoenix: extend an end when gapless, or fill a single gap.
  const lo = ranks[0];
  const hi = ranks[ranks.length - 1];
  const span = hi - lo;
  if (span === ranks.length - 1) return { type: 'straight', length: cards.length }; // gapless -> extend
  if (span === ranks.length) return { type: 'straight', length: cards.length }; // one gap -> fill
  return null;
}

function tryStairs(cards: readonly Card[], pcount: number): TichuComboResult | null {
  if (cards.length < 4 || cards.length % 2 !== 0) return null;
  const rest = nonPhoenix(cards);
  const counts = rankCounts(rest);
  for (const r of counts.keys()) {
    if (r < 2 || r > 14) return null; // Mahjong(1)/Dragon(15) cannot pair
  }
  const ranks = sortedDistinctRanks(counts);
  if (ranks[ranks.length - 1] - ranks[0] !== ranks.length - 1) return null; // ranks not consecutive
  let singles = 0;
  for (const r of ranks) {
    const c = counts.get(r) ?? 0;
    if (c === 2) continue;
    if (c === 1) singles++;
    else return null;
  }
  if (pcount === 0) {
    return singles === 0 ? { type: 'stairs', length: cards.length } : null;
  }
  // One phoenix: exactly one rank must be a lone single.
  return singles === 1 ? { type: 'stairs', length: cards.length } : null;
}

function tryFullHouse(cards: readonly Card[], pcount: number): TichuComboResult | null {
  if (cards.length !== 5) return null;
  const rest = nonPhoenix(cards);
  const counts = rankCounts(rest);
  for (const r of counts.keys()) {
    if (r < 2 || r > 14) return null;
  }
  if (pcount === 0) {
    let trip = 0;
    let pair = 0;
    for (const [r, c] of counts) {
      if (c === 3) trip = r;
      else if (c === 2) pair = r;
      else return null;
    }
    return trip !== 0 && pair !== 0 ? { type: 'fullHouse', length: 0 } : null;
  }
  // One phoenix (four normal cards).
  const ranks = sortedDistinctRanks(counts);
  if (ranks.length !== 2) return null;
  const [r0, r1] = ranks;
  const c0 = counts.get(r0) ?? 0;
  const c1 = counts.get(r1) ?? 0;
  if (c0 === 3 && c1 === 1) return { type: 'fullHouse', length: 0 };
  if (c0 === 1 && c1 === 3) return { type: 'fullHouse', length: 0 };
  if (c0 === 2 && c1 === 2) return { type: 'fullHouse', length: 0 };
  return null;
}

/**
 * Classifies a selection of cards into its Tichu combo type, faithfully mirroring
 * the domain `ClassifyTichu` (internal/domain/TichuEval.go). Returns `invalid`
 * for any selection that does not form a legal combo, and `null`-equivalent
 * `invalid` for an empty selection. This is an ADDITIVE preview only — the
 * backend remains the source of truth and rejects any truly-illegal play.
 */
export function classifyTichuCombo(cards: readonly Card[]): TichuComboResult {
  const n = cards.length;
  if (n === 0) return INVALID;

  let hasDog = false;
  let hasDragon = false;
  for (const c of cards) {
    const kind = specialKind(c);
    if (kind === TICHU_DOG) hasDog = true;
    else if (kind === TICHU_DRAGON) hasDragon = true;
  }
  const pcount = phoenixCount(cards);

  // Dog: single lead only.
  if (hasDog) return n === 1 ? { type: 'dog', length: 0 } : INVALID;
  // Dragon: single only.
  if (hasDragon) return n === 1 ? { type: 'single', length: 0 } : INVALID;

  if (n === 1) return { type: 'single', length: 0 };
  if (pcount > 1) return INVALID; // at most one phoenix

  let result: TichuComboResult | null = null;
  if (n === 2) result = tryPair(cards, pcount);
  else if (n === 3) result = tryTriple(cards, pcount);
  else {
    result =
      tryBomb4(cards, pcount) ??
      tryStraightFlush(cards, pcount) ??
      tryStraight(cards, pcount) ??
      tryStairs(cards, pcount) ??
      tryFullHouse(cards, pcount);
  }
  return result ?? INVALID;
}
