import type { Card } from '../types/card';
import { suitSymbol } from './cardAlt';
import { valueName } from './cardUtils';

/** Backend's default Chinchón knock threshold (config.knockThreshold, default 5). */
export const CHINCHON_DEFAULT_KNOCK_THRESHOLD = 5;

/**
 * Adjacency position of a rank in the 40-card Latin deck (8/9/10 removed).
 * Runs treat A,2,3,4,5,6,7,J,Q,K as consecutive, so 7 and J are adjacent.
 * Returns A=1,2=2,...,7=7,J=8,Q=9,K=10; ranks absent from the deck return 0.
 * Mirrors `chinchonRankPosition` in internal/domain/Chinchon.go.
 */
export function chinchonRankPosition(value: number): number {
  if (value >= 1 && value <= 7) return value;
  if (value === 11) return 8; // J
  if (value === 12) return 9; // Q
  if (value === 13) return 10; // K
  return 0;
}

/**
 * Short type label for a Chinchón meld, derived from its cards (no API change):
 * a same-rank set yields the rank (e.g. `7`, `K`); a same-suit run yields the
 * suit symbol with its low–high range (e.g. `♠ 7-K`). Returns `''` for an empty
 * meld. Used as a header badge so players can spot layoff targets quickly.
 */
export function chinchonMeldLabel(cards: readonly Card[]): string {
  if (cards.length === 0) return '';
  const first = cards[0];
  const sameRank = cards.every((c) => c.value === first.value);
  if (sameRank) return valueName(first.value); // set
  const sorted = [...cards].sort((a, b) => chinchonRankPosition(a.value) - chinchonRankPosition(b.value)); // run
  return `${suitSymbol(first.design)} ${valueName(sorted[0].value)}-${valueName(sorted[sorted.length - 1].value)}`;
}

/** Chinchón card value (A=1, 2-7=face, J/Q/K=10). */
export function chinchonCardValue(card: Card): number {
  if (card.value === 1) return 1;
  if (card.value >= 10) return 10;
  return card.value;
}

/** Sum of card values in a deadwood pile. */
export function calcChinchonDeadwoodValue(cards: readonly Card[]): number {
  let total = 0;
  for (const c of cards) total += chinchonCardValue(c);
  return total;
}

/** Indexed handle so the meld-finder compares card identity, not value
 * (suits matter for runs and identical-rank duplicates would otherwise collide). */
interface IndexedCard {
  idx: number;
  card: Card;
}

/** Find the minimum-deadwood split of `hand` into melds + deadwood.
 * Mirrors `FindBestMelds` in internal/domain/Chinchon.go but the only value
 * surfaced back to React is the integer deadwood total. */
export function bestChinchonDeadwoodValue(hand: readonly Card[]): number {
  if (hand.length === 0) return 0;
  return search(hand.map((card, idx) => ({ idx, card }))).value;
}

/** The deadwood cards (and their values) left over by the best meld split. */
export interface ChinchonDeadwoodBreakdown {
  /** Cards not absorbed by any meld in the minimum-deadwood split. */
  cards: Card[];
  /** Each card's Chinchón point value, parallel to `cards`. */
  values: number[];
  /** Sum of `values` (the deadwood score). */
  total: number;
}

/** Breakdown of which cards remain as deadwood in the best meld split, for a
 * per-card hint like "5 + 3 + 2 = 10". */
export function chinchonDeadwoodBreakdown(hand: readonly Card[]): ChinchonDeadwoodBreakdown {
  const best = search(hand.map((card, idx) => ({ idx, card })));
  const cards = best.deadwood.map((d) => d.card);
  return { cards, values: cards.map(chinchonCardValue), total: best.value };
}

/** The best meld split of a hand: which card indices are absorbed by melds. */
export interface ChinchonMeldSplit {
  /** Indices (into `hand`) of cards that belong to a meld in the minimum-deadwood split. */
  meldedIndices: ReadonlySet<number>;
  /** Deadwood total of the split. */
  deadwoodValue: number;
}

/**
 * Find the minimum-deadwood split and report which hand indices are melded, so
 * the UI can color-code melded vs. deadwood cards (mirrors Gin Rummy's
 * `bestMeldSplit`). Shares the same search as the deadwood breakdown, so the
 * coloring and the score always agree. Empty hand yields an empty set.
 */
export function bestChinchonMeldSplit(hand: readonly Card[]): ChinchonMeldSplit {
  if (hand.length === 0) return { meldedIndices: new Set(), deadwoodValue: 0 };
  const best = search(hand.map((card, idx) => ({ idx, card })));
  const deadwoodIdx = new Set(best.deadwood.map((d) => d.idx));
  const melded = new Set<number>();
  for (let i = 0; i < hand.length; i++) if (!deadwoodIdx.has(i)) melded.add(i);
  return { meldedIndices: melded, deadwoodValue: best.value };
}

function search(remaining: IndexedCard[]): { value: number; deadwood: IndexedCard[] } {
  const candidates = enumerateMelds(remaining);
  let best = { value: calcChinchonDeadwoodValue(remaining.map((r) => r.card)), deadwood: remaining };
  if (candidates.length === 0) return best;
  for (const meld of candidates) {
    const meldIdx = new Set(meld.map((m) => m.idx));
    const rest = remaining.filter((r) => !meldIdx.has(r.idx));
    const sub = search(rest);
    if (sub.value < best.value) best = sub;
    if (best.value === 0) break;
  }
  return best;
}

function enumerateMelds(cards: IndexedCard[]): IndexedCard[][] {
  const out: IndexedCard[][] = [];

  // Sets: 3+ same value across any suit
  const byValue = new Map<number, IndexedCard[]>();
  for (const c of cards) {
    const list = byValue.get(c.card.value) ?? [];
    list.push(c);
    byValue.set(c.card.value, list);
  }
  for (const group of byValue.values()) {
    if (group.length >= 3) {
      out.push(group.slice(0, 3));
      if (group.length >= 4) out.push(group.slice(0, 4));
    }
  }

  // Runs: 3+ consecutive (by rank position) in the same suit
  const bySuit = new Map<string, IndexedCard[]>();
  for (const c of cards) {
    if (chinchonRankPosition(c.card.value) === 0) continue;
    const list = bySuit.get(c.card.design) ?? [];
    list.push(c);
    bySuit.set(c.card.design, list);
  }
  for (const group of bySuit.values()) {
    if (group.length < 3) continue;
    const sorted = [...group].sort((a, b) => chinchonRankPosition(a.card.value) - chinchonRankPosition(b.card.value));
    for (let i = 0; i < sorted.length; i++) {
      const run: IndexedCard[] = [sorted[i]];
      for (let j = i + 1; j < sorted.length; j++) {
        if (chinchonRankPosition(sorted[j].card.value) === chinchonRankPosition(run[run.length - 1].card.value) + 1) {
          run.push(sorted[j]);
          if (run.length >= 3) out.push([...run]);
        } else {
          break;
        }
      }
    }
  }
  return out;
}
