import type { Card } from '../types/card';

/**
 * Dou Dizhu play categories, mirroring the Go domain `DoudizhuComboType`
 * (single / pair / trio / trio+kicker / straight / consecutive-pairs /
 * airplane / airplane+wings / bomb / rocket).
 */
export type DoudizhuComboType =
  | 'single'
  | 'pair'
  | 'trio'
  | 'trioSingle'
  | 'trioPair'
  | 'straight'
  | 'consecutivePair'
  | 'airplane'
  | 'airplaneSingle'
  | 'airplanePair'
  | 'bomb'
  | 'rocket';

/** A classified Dou Dizhu combo: its type plus the comparison rank and chain length. */
export interface DoudizhuCombo {
  type: DoudizhuComboType;
  /** Primary comparison rank (trio value, straight start, etc.). */
  rank: number;
  /** Chain length for straights / consecutive pairs / airplanes; 1 otherwise. */
  length: number;
}

const STRENGTH_SMALL_JOKER = 16;
const STRENGTH_BIG_JOKER = 17;

/**
 * Card strength, mirroring the Go domain `DoudizhuCardStrength`:
 * 3..K = 3..13, A = 14, 2 = 15, small joker = 16, big joker = 17.
 */
export function doudizhuCardStrength(card: Card): number {
  if (card.design === 'JOKER') {
    return card.value >= 2 ? STRENGTH_BIG_JOKER : STRENGTH_SMALL_JOKER;
  }
  if (card.value === 1) return 14;
  if (card.value === 2) return 15;
  return card.value;
}

interface RankCount {
  strength: number;
  count: number;
}

/** Frequency of each strength, sorted by count desc then strength asc (mirrors the domain). */
function rankCounts(cards: Card[]): RankCount[] {
  const freq = new Map<number, number>();
  for (const c of cards) {
    const s = doudizhuCardStrength(c);
    freq.set(s, (freq.get(s) ?? 0) + 1);
  }
  const result: RankCount[] = [];
  for (const [strength, count] of freq) result.push({ strength, count });
  result.sort((a, b) => (a.count !== b.count ? b.count - a.count : a.strength - b.strength));
  return result;
}

/** Both cards jokers = rocket (火箭). */
function isRocket(cards: Card[]): boolean {
  return cards.length === 2 && cards.every((c) => c.design === 'JOKER');
}

/** Chainable ranks are 3..A (strength 3..14); 2 and jokers cannot form chains. */
function isChainable(strength: number): boolean {
  return strength >= 3 && strength <= 14;
}

function classifyTrioKicker(ranks: RankCount[], n: number): DoudizhuCombo | null {
  const trio = ranks.find((r) => r.count === 3);
  if (!trio) return null;
  const kickerCount = n - 3;
  if (kickerCount === 1) return { type: 'trioSingle', rank: trio.strength, length: 1 };
  if (kickerCount === 2 && ranks.some((r) => r.strength !== trio.strength && r.count === 2)) {
    return { type: 'trioPair', rank: trio.strength, length: 1 };
  }
  return null;
}

function classifyStraight(ranks: RankCount[], n: number): DoudizhuCombo | null {
  if (n < 5) return null;
  if (!ranks.every((r) => r.count === 1 && isChainable(r.strength))) return null;
  const strengths = ranks.map((r) => r.strength).sort((a, b) => a - b);
  if (strengths.length !== n) return null;
  for (let i = 1; i < strengths.length; i++) {
    if (strengths[i] - strengths[i - 1] !== 1) return null;
  }
  return { type: 'straight', rank: strengths[0], length: n };
}

function classifyConsecutivePair(ranks: RankCount[], n: number): DoudizhuCombo | null {
  if (n < 6 || n % 2 !== 0) return null;
  const pairCount = n / 2;
  if (pairCount < 3) return null;
  if (!ranks.every((r) => r.count === 2 && isChainable(r.strength))) return null;
  const strengths = ranks.map((r) => r.strength).sort((a, b) => a - b);
  if (strengths.length !== pairCount) return null;
  for (let i = 1; i < strengths.length; i++) {
    if (strengths[i] - strengths[i - 1] !== 1) return null;
  }
  return { type: 'consecutivePair', rank: strengths[0], length: pairCount };
}

function findLongestConsecutive(sorted: number[]): number[] {
  if (sorted.length === 0) return [];
  let best = [sorted[0]];
  let current = [sorted[0]];
  for (let i = 1; i < sorted.length; i++) {
    current = sorted[i] === sorted[i - 1] + 1 ? [...current, sorted[i]] : [sorted[i]];
    if (current.length > best.length) best = [...current];
  }
  return best;
}

function classifyAirplane(ranks: RankCount[], n: number): DoudizhuCombo | null {
  const trios = ranks
    .filter((r) => r.count === 3 && isChainable(r.strength))
    .map((r) => r.strength)
    .sort((a, b) => a - b);
  if (trios.length < 2) return null;

  const chain = findLongestConsecutive(trios);
  if (chain.length < 2) return null;
  const chainLen = chain.length;

  if (n === chainLen * 3) return { type: 'airplane', rank: chain[0], length: chainLen };
  if (n === chainLen * 4) return { type: 'airplaneSingle', rank: chain[0], length: chainLen };
  if (n === chainLen * 5) {
    const inChain = (strength: number) => chain.includes(strength);
    for (const r of ranks) {
      const rem = inChain(r.strength) ? r.count - 3 : r.count;
      if (rem % 2 !== 0) return null;
    }
    return { type: 'airplanePair', rank: chain[0], length: chainLen };
  }
  return null;
}

/**
 * Classifies a selection of cards into its Dou Dizhu combo, mirroring the Go
 * domain `DoudizhuClassifyCombo`. Returns `null` for an invalid selection.
 */
export function classifyDoudizhuCombo(cards: Card[]): DoudizhuCombo | null {
  const n = cards.length;
  if (n === 0) return null;

  if (n === 2 && isRocket(cards)) return { type: 'rocket', rank: STRENGTH_BIG_JOKER, length: 1 };

  const ranks = rankCounts(cards);

  if (n === 1) return { type: 'single', rank: doudizhuCardStrength(cards[0]), length: 1 };
  if (n === 2) return ranks[0].count === 2 ? { type: 'pair', rank: ranks[0].strength, length: 1 } : null;
  if (n === 3 && ranks[0].count === 3) return { type: 'trio', rank: ranks[0].strength, length: 1 };
  if (n === 4) {
    if (ranks.length === 1 && ranks[0].count === 4) return { type: 'bomb', rank: ranks[0].strength, length: 1 };
    const combo = classifyTrioKicker(ranks, n);
    if (combo) return combo;
  }
  if (n === 5) {
    const combo = classifyTrioKicker(ranks, n);
    if (combo) return combo;
  }

  return classifyStraight(ranks, n) ?? classifyConsecutivePair(ranks, n) ?? classifyAirplane(ranks, n);
}

/**
 * Whether `play` beats `table`, mirroring the Go domain `DoudizhuCanBeat`.
 * Rocket beats everything; a bomb beats any non-bomb/non-rocket; otherwise the
 * types and chain lengths must match and `play` must outrank `table`.
 */
export function canBeatDoudizhu(play: DoudizhuCombo, table: DoudizhuCombo): boolean {
  if (play.type === 'rocket') return true;
  if (play.type === 'bomb') {
    if (table.type === 'bomb') return play.rank > table.rank;
    if (table.type === 'rocket') return false;
    return true;
  }
  if (play.type !== table.type) return false;
  if (play.length !== table.length) return false;
  return play.rank > table.rank;
}

/** Why a Dou Dizhu selection cannot be played, or `null` when it is playable. */
export type DoudizhuInvalidReason = 'notCombo' | 'noBeat';

/**
 * Pre-validates a selection against the current table before submitting.
 * Returns `'notCombo'` when the cards form no legal combo, `'noBeat'` when they
 * form a valid combo that cannot beat the table, or `null` when playable
 * (either a fresh lead or a beating play). Mirrors the Go domain gate in
 * `Doudizhu.PlayerPlay`.
 */
export function doudizhuInvalidReason(selected: Card[], tableCards: Card[]): DoudizhuInvalidReason | null {
  const combo = classifyDoudizhuCombo(selected);
  if (!combo) return 'notCombo';
  if (tableCards.length === 0) return null;
  const table = classifyDoudizhuCombo(tableCards);
  if (!table) return null;
  return canBeatDoudizhu(combo, table) ? null : 'noBeat';
}
