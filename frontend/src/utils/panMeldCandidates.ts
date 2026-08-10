import type { Card } from '../types/card';

/**
 * A candidate meld found in a Panguingue (Pan) hand — a minimal (3-card) legal
 * set or rope/run that the player could lay down. `indices` are positions into
 * the hand's `cards` array, sorted ascending so they map directly onto the
 * server's `meld` command payload.
 */
export interface PanMeldCandidate {
  kind: 'set' | 'run';
  indices: number[];
}

/** Minimum cards in a Pan meld (set or run). Mirrors the domain's `PanMeldMin`. */
const PAN_MELD_MIN = 3;

/**
 * Pan's reduced-deck rank order (A,2,3,4,5,6,7,J,Q,K — the 8s, 9s and 10s are
 * removed), mapping a card value to its position so that 7 and J are adjacent.
 * A run's consecutiveness is judged on these indices, not raw values. Mirrors
 * the domain's `panRankOrder`.
 */
const PAN_RANK_ORDER: Readonly<Record<number, number>> = {
  1: 0,
  2: 1,
  3: 2,
  4: 3,
  5: 4,
  6: 5,
  7: 6,
  11: 7,
  12: 8,
  13: 9,
};

/** The Ace's rank index when it ranks above the King (Q-K-A). Mirrors `panAceHighRank`. */
const PAN_ACE_HIGH_RANK = 10;

/** Returns the Pan rank index for a card value, or -1 for a non-Pan rank (8/9/10). */
function panRankIndex(value: number): number {
  const idx = PAN_RANK_ORDER[value];
  return idx === undefined ? -1 : idx;
}

/** True when every value equals its predecessor + 1 (input must be sorted). Mirrors `isConsecutive`. */
function isConsecutive(values: number[]): boolean {
  for (let i = 1; i < values.length; i++) {
    if (values[i] !== values[i - 1] + 1) return false;
  }
  return true;
}

/**
 * True when `meld` is a valid Pan set: 3+ cards of the same rank (duplicate
 * suits allowed on the multi-deck). Mirrors the domain's `panIsValidSet`.
 */
function panIsValidSet(meld: Card[]): boolean {
  if (meld.length < PAN_MELD_MIN) return false;
  return meld.every((c) => c.value === meld[0].value);
}

/**
 * True when `meld` is a valid Pan run/rope: 3+ same-suit cards consecutive by
 * reduced-deck rank index (Ace low or high, no wrap, no duplicate ranks).
 * Mirrors the domain's `panIsValidRun`.
 */
function panIsValidRun(meld: Card[]): boolean {
  if (meld.length < PAN_MELD_MIN) return false;
  const suit = meld[0].design;
  if (!meld.every((c) => c.design === suit)) return false;
  const idx: number[] = [];
  for (const c of meld) {
    const ri = panRankIndex(c.value);
    if (ri < 0) return false;
    idx.push(ri);
  }
  idx.sort((a, b) => a - b);
  for (let i = 1; i < idx.length; i++) {
    if (idx[i] === idx[i - 1]) return false; // duplicate ranks are not a run
  }
  if (isConsecutive(idx)) return true;
  // Re-evaluate the Ace as high (above the King), e.g. Q-K-A.
  if (idx[0] === 0) {
    const high = [...idx];
    high[0] = PAN_ACE_HIGH_RANK;
    high.sort((a, b) => a - b);
    if (isConsecutive(high)) return true;
  }
  return false;
}

/** Adds each length-3 consecutive window of a same-suit rank-index sequence as a run candidate. */
function collectRunWindows(
  seq: { rankIdx: number; cardIdx: number }[],
  out: PanMeldCandidate[],
  seen: Set<string>,
): void {
  // Dedup by rank index (a run cannot repeat a rank), keeping the first card at each rank.
  const byRank = new Map<number, number>();
  for (const { rankIdx, cardIdx } of seq) {
    if (!byRank.has(rankIdx)) byRank.set(rankIdx, cardIdx);
  }
  const sorted = [...byRank.entries()].sort((a, b) => a[0] - b[0]); // [rankIdx, cardIdx]
  for (let i = 0; i + PAN_MELD_MIN <= sorted.length; i++) {
    let consecutive = true;
    for (let k = 1; k < PAN_MELD_MIN; k++) {
      if (sorted[i + k][0] !== sorted[i + k - 1][0] + 1) {
        consecutive = false;
        break;
      }
    }
    if (!consecutive) continue;
    const indices = sorted.slice(i, i + PAN_MELD_MIN).map(([, cardIdx]) => cardIdx);
    indices.sort((a, b) => a - b);
    const key = `run:${indices.join(',')}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ kind: 'run', indices });
  }
}

/**
 * Enumerates the minimal (3-card) legal meld candidates in a Pan hand, so the
 * UI can pre-present them before the player melds. Each candidate is guaranteed
 * valid under the domain's meld rules (a set of same-rank cards, or a same-suit
 * run consecutive by the reduced-deck rank order with the Ace low or high).
 *
 * This mirrors the domain's `PanIsValidMeld`; it is intentionally a "which
 * cards form a legal meld" enumeration for hinting, not an optimal hand
 * partition. Only minimal 3-card melds are surfaced so every suggestion is
 * unambiguously legal.
 */
export function panMeldCandidates(cards: Card[]): PanMeldCandidate[] {
  const out: PanMeldCandidate[] = [];
  const seen = new Set<string>();

  // Sets: 3+ cards of the same rank.
  const byValue = new Map<number, number[]>();
  for (let i = 0; i < cards.length; i++) {
    const group = byValue.get(cards[i].value);
    if (group) group.push(i);
    else byValue.set(cards[i].value, [i]);
  }
  for (const group of byValue.values()) {
    if (group.length < PAN_MELD_MIN) continue;
    const indices = group.slice(0, PAN_MELD_MIN);
    const key = `set:${indices.join(',')}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ kind: 'set', indices });
  }

  // Runs: 3+ same-suit cards consecutive by reduced-deck rank index.
  const bySuit = new Map<string, number[]>();
  for (let i = 0; i < cards.length; i++) {
    if (panRankIndex(cards[i].value) < 0) continue; // skip non-Pan ranks defensively
    const group = bySuit.get(cards[i].design);
    if (group) group.push(i);
    else bySuit.set(cards[i].design, [i]);
  }
  for (const group of bySuit.values()) {
    if (group.length < PAN_MELD_MIN) continue;
    const low = group.map((cardIdx) => ({ rankIdx: panRankIndex(cards[cardIdx].value), cardIdx }));
    collectRunWindows(low, out, seen);
    // Ace-high alternative (Q-K-A) when the suit holds an Ace.
    if (group.some((cardIdx) => cards[cardIdx].value === 1)) {
      const high = group.map((cardIdx) => ({
        rankIdx: cards[cardIdx].value === 1 ? PAN_ACE_HIGH_RANK : panRankIndex(cards[cardIdx].value),
        cardIdx,
      }));
      collectRunWindows(high, out, seen);
    }
  }

  return out;
}

/**
 * Returns the set of hand indices that can be laid off onto at least one of the
 * given table melds. Mirrors the domain's `PanCanLayoff`: a card extends a set
 * if it keeps the set same-rank, or extends a run if the run stays a valid
 * same-suit sequence.
 */
export function panLayoffIndices(hand: Card[], tableMelds: Card[][]): Set<number> {
  const result = new Set<number>();
  for (let i = 0; i < hand.length; i++) {
    for (const meld of tableMelds) {
      if (panCanLayoff(meld, hand[i])) {
        result.add(i);
        break;
      }
    }
  }
  return result;
}

/** True when `card` can be laid off onto `meld`. Mirrors the domain's `PanCanLayoff`. */
function panCanLayoff(meld: Card[], card: Card): boolean {
  if (meld.length === 0) return false;
  const candidate = [...meld, card];
  if (panIsValidSet(meld)) return panIsValidSet(candidate);
  return panIsValidRun(candidate);
}

/**
 * Ranks whose sets pay a chip to every player — a "valle" (バジェ). Mirrors the
 * domain's `panValleRanks`.
 */
const PAN_VALLE_RANKS: readonly number[] = [3, 5, 7];

/**
 * True when the cards form a valle: a set (3+ of the same rank) of 3s, 5s or 7s.
 * Mirrors the domain's `PanIsValleMeld`. Laying one moves every player's chip
 * count, and neither UI said which meld caused it (#4853).
 */
export function isPanValleMeld(cards: readonly Card[]): boolean {
  if (cards.length < PAN_MELD_MIN) return false;
  const value = cards[0].value;
  if (!PAN_VALLE_RANKS.includes(value)) return false;
  return cards.every((c) => c.value === value);
}
