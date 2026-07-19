import type { Card } from '../types/card';

/**
 * Deadwood penalty for a wild card left unmatched, mirroring
 * `ThreeThirteenWildDeadwoodValue` in internal/domain/ThreeThirteen.go.
 */
export const THREETHIRTEEN_WILD_DEADWOOD_VALUE = 20;

/** Minimum meld size (set or run), mirroring `ThreeThirteenMeldMinSize`. */
const MELD_MIN_SIZE = 3;

/** Whether `card` is this round's wild (its rank equals `wildRank`). */
export function threeThirteenIsWild(card: Card, wildRank: number): boolean {
  return card.value === wildRank;
}

/** Face value of a single card for scoring: A=1, 2-10=face, J/Q/K=10. */
export function threeThirteenCardValue(card: Card): number {
  return card.value >= 10 ? 10 : card.value;
}

/**
 * Total deadwood points of a pile: face value per natural card, but a leftover
 * wild counts as {@link THREETHIRTEEN_WILD_DEADWOOD_VALUE}. Mirrors
 * `threeThirteenDeadwoodValue` in the domain.
 */
export function calcThreeThirteenDeadwoodValue(cards: readonly Card[], wildRank: number): number {
  let total = 0;
  for (const c of cards) {
    total += threeThirteenIsWild(c, wildRank) ? THREETHIRTEEN_WILD_DEADWOOD_VALUE : threeThirteenCardValue(c);
  }
  return total;
}

/** Indexed handle so the meld search compares card identity, not value
 * (duplicate ranks across the two decks would otherwise collide). */
interface IndexedCard {
  idx: number;
  card: Card;
}

/**
 * Minimum deadwood value obtainable by splitting `hand` into melds
 * (wild-aware sets and runs) plus leftover deadwood. Ports
 * `threeThirteenBestMelds` + `threeThirteenDeadwoodValue` from the domain but
 * surfaces only the integer total that the UI needs.
 */
export function bestThreeThirteenDeadwoodValue(hand: readonly Card[], wildRank: number): number {
  if (hand.length === 0) return 0;
  return search(
    hand.map((card, idx) => ({ idx, card })),
    wildRank,
  );
}

/**
 * Minimum deadwood reachable after discarding exactly one card from `hand`,
 * i.e. the best post-discard total. Mirrors `threeThirteenBestDiscardValue`.
 * Returns 0 for an empty hand.
 */
export function bestThreeThirteenDiscardValue(hand: readonly Card[], wildRank: number): number {
  if (hand.length === 0) return 0;
  let best = -1;
  for (let i = 0; i < hand.length; i++) {
    const rest = hand.filter((_, j) => j !== i);
    const v = bestThreeThirteenDeadwoodValue(rest, wildRank);
    if (best < 0 || v < best) best = v;
    if (best === 0) break;
  }
  return best;
}

/** Recursively split `remaining` into melds, returning the minimum deadwood value. */
function search(remaining: IndexedCard[], wildRank: number): number {
  const candidates = enumerateMelds(remaining, wildRank);
  let best = calcThreeThirteenDeadwoodValue(
    remaining.map((r) => r.card),
    wildRank,
  );
  if (candidates.length === 0) return best;
  for (const meld of candidates) {
    const meldIdx = new Set(meld.map((m) => m.idx));
    const rest = remaining.filter((r) => !meldIdx.has(r.idx));
    const sub = search(rest, wildRank);
    if (sub < best) best = sub;
    if (best === 0) break; // perfect meld — optimal
  }
  return best;
}

/** Enumerate candidate melds (3-4 cards) constructible from `cards`, using
 * wilds to substitute in sets or fill run gaps. Ports `threeThirteenAllMelds`. */
function enumerateMelds(cards: IndexedCard[], wildRank: number): IndexedCard[][] {
  const out: IndexedCard[][] = [];
  const wilds: IndexedCard[] = [];
  const naturals: IndexedCard[] = [];
  for (const c of cards) {
    if (threeThirteenIsWild(c.card, wildRank)) wilds.push(c);
    else naturals.push(c);
  }

  // Sets: same-rank naturals topped up with wilds to reach 3-4 cards.
  const byRank = new Map<number, IndexedCard[]>();
  for (const c of naturals) {
    const list = byRank.get(c.card.value) ?? [];
    list.push(c);
    byRank.set(c.card.value, list);
  }
  for (const group of byRank.values()) {
    for (let size = MELD_MIN_SIZE; size <= MELD_MIN_SIZE + 1; size++) {
      const needWild = Math.max(size - group.length, 0);
      const natTake = size - needWild;
      if (natTake < 1 || natTake > group.length || needWild > wilds.length) continue;
      out.push([...group.slice(0, natTake), ...wilds.slice(0, needWild)]);
    }
  }

  // Runs: same-suit consecutive ranks, wilds filling gaps (Ace low or high).
  const bySuit = new Map<string, Map<number, IndexedCard>>();
  for (const c of naturals) {
    let byVal = bySuit.get(c.card.design);
    if (!byVal) {
      byVal = new Map<number, IndexedCard>();
      bySuit.set(c.card.design, byVal);
    }
    if (!byVal.has(c.card.value)) byVal.set(c.card.value, c);
  }
  for (const byVal of bySuit.values()) {
    out.push(...runsInSuit(byVal, wilds));
  }
  return out;
}

/** Enumerate 3-4 card runs within one suit, using wilds to bridge gaps and
 * allowing Ace to play high (14). Ports `threeThirteenRunsIn`. */
function runsInSuit(byVal: Map<number, IndexedCard>, wilds: IndexedCard[]): IndexedCard[][] {
  const out: IndexedCard[][] = [];
  // Ace (1) may also fill the high slot (14).
  const view = new Map<number, IndexedCard>(byVal);
  const ace = byVal.get(1);
  if (ace) view.set(14, ace);

  for (let size = MELD_MIN_SIZE; size <= MELD_MIN_SIZE + 1; size++) {
    for (let start = 1; start + size - 1 <= 14; start++) {
      const picked: IndexedCard[] = [];
      let wildsUsed = 0;
      let ok = true;
      for (let v = start; v < start + size; v++) {
        const found = view.get(v);
        if (found) {
          picked.push(found);
        } else if (wildsUsed < wilds.length) {
          picked.push(wilds[wildsUsed]);
          wildsUsed++;
        } else {
          ok = false;
          break;
        }
      }
      // A run needs at least one natural (all-wild is not a valid run).
      if (ok && wildsUsed < picked.length) out.push(picked);
    }
  }
  return out;
}
