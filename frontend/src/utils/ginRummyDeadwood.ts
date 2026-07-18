import type { Card } from '../types/card';
import { suitSymbol } from './cardAlt';
import { valueName } from './cardUtils';

/** Backend's GinRummyKnockThreshold (10). */
export const GIN_RUMMY_KNOCK_THRESHOLD = 10;

/**
 * Short type label for a Gin Rummy meld, derived from its cards (no API change):
 * a same-rank set yields the rank (e.g. `7`, `K`); a same-suit run yields the
 * suit symbol with its low–high range (e.g. `♠ 3-5`). Returns `''` for an empty
 * meld. Used as a header badge so players can spot layoff targets quickly.
 */
export function ginRummyMeldLabel(cards: readonly Card[]): string {
  if (cards.length === 0) return '';
  const first = cards[0];
  const sameRank = cards.every((c) => c.value === first.value);
  if (sameRank) return valueName(first.value); // set
  const sorted = cards.map((c) => c.value).sort((a, b) => a - b); // run
  return `${suitSymbol(first.design)} ${valueName(sorted[0])}-${valueName(sorted[sorted.length - 1])}`;
}

/** Gin Rummy card value (A=1, 2-9=face, 10/J/Q/K=10). */
export function ginRummyCardValue(card: Card): number {
  if (card.value === 1) return 1;
  if (card.value >= 10) return 10;
  return card.value;
}

/** Sum of card values in a deadwood pile. */
export function calcDeadwoodValue(cards: readonly Card[]): number {
  let total = 0;
  for (const c of cards) total += ginRummyCardValue(c);
  return total;
}

/** Indexed handle so the meld-finder can compare card identity rather than
 * value (suits matter for runs and identical-rank duplicates would otherwise
 * collide). */
interface IndexedCard {
  idx: number;
  card: Card;
}

/** Result of splitting a hand into melds + deadwood: which original hand
 * indices belong to some meld in the best split, plus the deadwood total. */
export interface MeldSplit {
  /** Original hand indices that are part of a meld in the best split. */
  meldedIndices: ReadonlySet<number>;
  /** Sum of card values of the unmelded (deadwood) cards. */
  deadwoodValue: number;
}

/** Find the minimum-deadwood split of `hand`, returning which original hand
 * indices are melded and the deadwood total of the rest. Mirrors
 * `FindBestMelds` in internal/domain/GinRummy.go. Shares the same search as
 * {@link bestDeadwoodValue}, so the two stay consistent. */
export function bestMeldSplit(hand: readonly Card[]): MeldSplit {
  if (hand.length === 0) return { meldedIndices: new Set(), deadwoodValue: 0 };
  const indexed: IndexedCard[] = hand.map((card, idx) => ({ idx, card }));
  const { value, melded } = search(indexed);
  return { meldedIndices: melded, deadwoodValue: value };
}

/** Find the minimum-deadwood split of `hand` into melds + deadwood.
 * Mirrors `FindBestMelds` in internal/domain/GinRummy.go but the only
 * value we surface back is the integer deadwood total. */
export function bestDeadwoodValue(hand: readonly Card[]): number {
  return bestMeldSplit(hand).deadwoodValue;
}

/** Minimum-deadwood split of `remaining`, returning the deadwood value and the
 * set of original indices that end up melded in that best split. */
function search(remaining: IndexedCard[]): { value: number; melded: Set<number> } {
  const candidates = enumerateMelds(remaining);
  // Baseline: take no meld here → every remaining card is deadwood.
  let best = { value: calcDeadwoodValue(remaining.map((r) => r.card)), melded: new Set<number>() };
  if (candidates.length === 0) return best;
  for (const meld of candidates) {
    const meldIdx = new Set(meld.map((m) => m.idx));
    const rest = remaining.filter((r) => !meldIdx.has(r.idx));
    const sub = search(rest);
    if (sub.value < best.value) {
      const melded = new Set<number>(meldIdx);
      for (const i of sub.melded) melded.add(i);
      best = { value: sub.value, melded };
    }
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

  // Runs: 3+ consecutive in the same suit
  const bySuit = new Map<string, IndexedCard[]>();
  for (const c of cards) {
    const list = bySuit.get(c.card.design) ?? [];
    list.push(c);
    bySuit.set(c.card.design, list);
  }
  for (const group of bySuit.values()) {
    if (group.length < 3) continue;
    const sorted = [...group].sort((a, b) => a.card.value - b.card.value);
    for (let i = 0; i < sorted.length; i++) {
      const run: IndexedCard[] = [sorted[i]];
      for (let j = i + 1; j < sorted.length; j++) {
        if (sorted[j].card.value === run[run.length - 1].card.value + 1) {
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
