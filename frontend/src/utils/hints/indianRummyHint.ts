import type { Card, IndianRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { IndianRummyPhase } from '../../types/phases';

/**
 * Returns a frontend HintResult for Indian Rummy, or null if no suggestion
 * applies (no human seat, empty hand, game over, or not the human's turn).
 */
export function getIndianRummyHint(state: IndianRummyResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;
  const human = state.players[humanIdx];
  if (human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.currentPlayerIdx !== humanIdx) return null;

  if (state.phase === IndianRummyPhase.DRAW) {
    return getDrawHint(human.cards, state.discardTop, state.wildRank);
  }
  if (state.phase === IndianRummyPhase.DISCARD) {
    return getDiscardHint(human.cards, state.wildRank);
  }
  return null;
}

/** Draw phase: take the discard top when it fits, otherwise draw from the stock. */
function getDrawHint(hand: Card[], discardTop: Card | null, wildRank: number): HintResult {
  if (discardTop && !isWild(discardTop, wildRank) && fitsWithHand(discardTop, hand, wildRank)) {
    return { targetAction: 'drawDiscard', reason: 'hint.drawFromDiscard', confidence: 'strong' };
  }
  return { targetAction: 'drawStock', reason: 'hint.drawFromStock', confidence: 'moderate' };
}

/** Discard phase: declare when a single discard clears all deadwood, else drop deadwood. */
function getDiscardHint(hand: Card[], wildRank: number): HintResult {
  let best = Number.POSITIVE_INFINITY;
  for (let i = 0; i < hand.length; i++) {
    const dv = calcDeadwood(
      hand.filter((_, j) => j !== i),
      wildRank,
    );
    if (dv < best) best = dv;
    if (best === 0) break;
  }

  if (best === 0) {
    return { targetAction: 'declare', reason: 'hint.declareNow', confidence: 'strong' };
  }
  return { targetAction: 'discard', reason: 'hint.discardDeadwood', confidence: 'moderate' };
}

/** A card is wild if it is a printed joker or matches the turned-up wild rank. */
export function isWild(card: Card, wildRank: number): boolean {
  return card.design === 'JOKER' || (wildRank > 0 && card.value === wildRank);
}

/** Check if a card fits with the hand to form (or extend) a potential meld. */
function fitsWithHand(card: Card, hand: Card[], wildRank: number): boolean {
  const natural = hand.filter((c) => !isWild(c, wildRank));

  // Set: another card of the same value.
  if (natural.filter((c) => c.value === card.value).length >= 1) return true;

  // Run: same suit within two ranks (adjacent or one-gap).
  for (const c of natural) {
    if (c.design !== card.design) continue;
    const diff = Math.abs(c.value - card.value);
    if (diff === 1 || diff === 2) return true;
  }
  return false;
}

/**
 * Calculate deadwood points for a hand. Wild cards never count as deadwood and
 * each additionally cancels the highest-value unmatched natural card (a wild can
 * complete or extend a meld). Runs/sets are detected among the natural cards.
 */
export function calcDeadwood(hand: Card[], wildRank: number): number {
  const natural = hand.filter((c) => !isWild(c, wildRank));
  const wildCount = hand.length - natural.length;

  const melds = findMelds(natural);
  const inMeld = new Set<number>();
  for (const meld of melds) {
    for (const idx of meld) inMeld.add(idx);
  }

  const leftover: number[] = [];
  for (let i = 0; i < natural.length; i++) {
    if (!inMeld.has(i)) leftover.push(cardPoint(natural[i]));
  }
  // A wild can absorb the highest remaining deadwood cards.
  leftover.sort((a, b) => b - a);
  return leftover.slice(wildCount).reduce((sum, v) => sum + v, 0);
}

/** Point value of a card (Ace = 10, 10/J/Q/K = 10, joker = 0, else pip). Matches the backend indianRummyCardPoints. */
function cardPoint(card: Card): number {
  if (card.design === 'JOKER') return 0;
  if (card.value === 1) return 10; // Ace scores 10, not 1
  if (card.value >= 10) return 10;
  return card.value;
}

/** Find melds (sets and runs); tries sets-first and runs-first, keeps the lower-deadwood one. */
function findMelds(hand: Card[]): number[][] {
  const setsFirst = findMeldsSetsFirst(hand);
  const runsFirst = findMeldsRunsFirst(hand);
  return deadwoodForMelds(hand, setsFirst) <= deadwoodForMelds(hand, runsFirst) ? setsFirst : runsFirst;
}

/** Sum the point value of cards not covered by any meld. */
function deadwoodForMelds(hand: Card[], melds: number[][]): number {
  const inMeld = new Set<number>();
  for (const meld of melds) {
    for (const idx of meld) inMeld.add(idx);
  }
  let dw = 0;
  for (let i = 0; i < hand.length; i++) {
    if (!inMeld.has(i)) dw += cardPoint(hand[i]);
  }
  return dw;
}

/** Detect melds prioritizing sets first, then runs from the remaining cards. */
function findMeldsSetsFirst(hand: Card[]): number[][] {
  const melds: number[][] = [];
  const used = new Set<number>();

  const byValue = new Map<number, number[]>();
  for (let i = 0; i < hand.length; i++) {
    const v = hand[i].value;
    if (!byValue.has(v)) byValue.set(v, []);
    byValue.get(v)?.push(i);
  }
  for (const [, indices] of byValue) {
    if (indices.length >= 3) {
      melds.push(indices);
      for (const idx of indices) used.add(idx);
    }
  }

  findRuns(hand, used, melds);
  return melds;
}

/** Detect melds prioritizing runs first, then sets from the remaining cards. */
function findMeldsRunsFirst(hand: Card[]): number[][] {
  const melds: number[][] = [];
  const used = new Set<number>();

  findRuns(hand, used, melds);

  const byValue = new Map<number, number[]>();
  for (let i = 0; i < hand.length; i++) {
    if (used.has(i)) continue;
    const v = hand[i].value;
    if (!byValue.has(v)) byValue.set(v, []);
    byValue.get(v)?.push(i);
  }
  for (const [, indices] of byValue) {
    if (indices.length >= 3) {
      melds.push(indices);
      for (const idx of indices) used.add(idx);
    }
  }
  return melds;
}

/** Find runs (3+ consecutive cards of the same suit) among unused cards. */
function findRuns(hand: Card[], used: Set<number>, melds: number[][]): void {
  const bySuit = new Map<string, { idx: number; value: number }[]>();
  for (let i = 0; i < hand.length; i++) {
    if (used.has(i)) continue;
    const s = hand[i].design;
    if (!bySuit.has(s)) bySuit.set(s, []);
    bySuit.get(s)?.push({ idx: i, value: hand[i].value });
  }
  for (const [, cards] of bySuit) {
    cards.sort((a, b) => a.value - b.value);
    let run: { idx: number; value: number }[] = [cards[0]];
    for (let i = 1; i < cards.length; i++) {
      if (cards[i].value === run[run.length - 1].value + 1) {
        run.push(cards[i]);
      } else {
        if (run.length >= 3) {
          melds.push(run.map((r) => r.idx));
          for (const r of run) used.add(r.idx);
        }
        run = [cards[i]];
      }
    }
    if (run.length >= 3) {
      melds.push(run.map((r) => r.idx));
      for (const r of run) used.add(r.idx);
    }
  }
}
