import type { Card, GinRummyMeld, GinRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GinRummyPhase } from '../../types/phases';

/** Maximum deadwood points to allow knocking. */
const KNOCK_THRESHOLD = 10;

/** Returns a frontend HintResult for Gin Rummy, or null if no suggestion. */
export function getGinRummyHint(state: GinRummyResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === GinRummyPhase.DRAW) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getDrawHint(human.cards, state.discardTop);
  }

  if (state.phase === GinRummyPhase.DISCARD) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getDiscardHint(human.cards);
  }

  if (state.phase === GinRummyPhase.LAYOFF) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getLayoffHint(human.cards, state.knockerMelds);
  }

  return null;
}

/** Hint for the draw phase: draw from discard if it helps, else draw from stock. */
function getDrawHint(hand: Card[], discardTop: Card | null): HintResult {
  if (discardTop && fitsWithHand(discardTop, hand)) {
    return { targetAction: 'drawDiscard', reason: 'hint.drawFromDiscard', confidence: 'strong' };
  }
  return { targetAction: 'drawStock', reason: 'hint.drawFromStock', confidence: 'moderate' };
}

/** Hint for the discard phase: suggest knock/gin if possible, else discard highest deadwood. */
function getDiscardHint(hand: Card[]): HintResult {
  const deadwood = calcDeadwood(hand);
  const bestDiscard = findBestDiscard(hand);

  const deadwoodAfterDiscard = bestDiscard !== null ? calcDeadwood(hand.filter((_, i) => i !== bestDiscard)) : deadwood;

  if (deadwoodAfterDiscard === 0) {
    return { targetAction: 'knock', reason: 'hint.ginOpportunity', confidence: 'strong' };
  }

  if (deadwoodAfterDiscard <= KNOCK_THRESHOLD) {
    return { targetAction: 'knock', reason: 'hint.knockNow', confidence: 'strong' };
  }

  return { targetAction: 'discard', reason: 'hint.discardDeadwood', confidence: 'moderate' };
}

/** Hint for the layoff phase: suggest laying off if any card fits knocker's melds. */
function getLayoffHint(hand: Card[], knockerMelds: GinRummyMeld[]): HintResult {
  const canLayoff = hand.some((card) => knockerMelds.some((meld) => fitsInMeld(card, meld)));

  if (canLayoff) {
    return { targetAction: 'layoff', reason: 'hint.layoffCards', confidence: 'strong' };
  }

  return { targetAction: 'skipLayoff', reason: 'hint.skipLayoff', confidence: 'moderate' };
}

/** Check if a card fits with the hand to form a potential meld. */
function fitsWithHand(card: Card, hand: Card[]): boolean {
  // Check for set: same value
  const sameValue = hand.filter((c) => c.value === card.value).length;
  if (sameValue >= 2) return true;

  // Check for run: same suit, adjacent values
  const sameSuit = hand.filter((c) => c.design === card.design);
  for (const c of sameSuit) {
    const diff = Math.abs(c.value - card.value);
    // Adjacent (diff === 1) or gap of 1 (diff === 2): partial run worth drawing
    if (diff === 1 || diff === 2) {
      return true;
    }
  }

  return false;
}

/** Check if a card can be laid off onto a meld. */
function fitsInMeld(card: Card, meld: GinRummyMeld): boolean {
  if (meld.cards.length === 0) return false;

  // Set meld: all same value
  if (meld.cards.every((c) => c.value === meld.cards[0].value)) {
    return card.value === meld.cards[0].value;
  }

  // Run meld: same suit, consecutive values
  const meldSuit = meld.cards[0].design;
  if (card.design !== meldSuit) return false;

  const values = meld.cards.map((c) => c.value).sort((a, b) => a - b);
  const minVal = values[0];
  const maxVal = values[values.length - 1];
  return card.value === minVal - 1 || card.value === maxVal + 1;
}

/** Calculate deadwood points for a hand (cards not in any meld). */
export function calcDeadwood(hand: Card[]): number {
  const melds = findMelds(hand);
  const inMeld = new Set<number>();
  for (const meld of melds) {
    for (const idx of meld) {
      inMeld.add(idx);
    }
  }

  let deadwood = 0;
  for (let i = 0; i < hand.length; i++) {
    if (!inMeld.has(i)) {
      deadwood += cardPoint(hand[i]);
    }
  }
  return deadwood;
}

/** Point value of a card for deadwood calculation. */
function cardPoint(card: Card): number {
  if (card.value >= 10) return 10;
  return card.value;
}

/** Find the best card to discard (index of highest-deadwood card not in a meld). */
function findBestDiscard(hand: Card[]): number | null {
  const melds = findMelds(hand);
  const inMeld = new Set<number>();
  for (const meld of melds) {
    for (const idx of meld) {
      inMeld.add(idx);
    }
  }

  let bestIdx: number | null = null;
  let bestPoints = -1;
  for (let i = 0; i < hand.length; i++) {
    if (!inMeld.has(i)) {
      const pts = cardPoint(hand[i]);
      if (pts > bestPoints) {
        bestPoints = pts;
        bestIdx = i;
      }
    }
  }
  return bestIdx;
}

/** Find melds (sets and runs) in a hand. Tries both sets-first and runs-first, returns the configuration that minimizes deadwood. */
function findMelds(hand: Card[]): number[][] {
  const setsFirst = findMeldsSetsFirst(hand);
  const runsFirst = findMeldsRunsFirst(hand);

  const deadwoodSF = calcDeadwoodForMelds(hand, setsFirst);
  const deadwoodRF = calcDeadwoodForMelds(hand, runsFirst);

  return deadwoodSF <= deadwoodRF ? setsFirst : runsFirst;
}

/** Calculate deadwood for a given set of melds. */
function calcDeadwoodForMelds(hand: Card[], melds: number[][]): number {
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

/** Find melds prioritizing sets first, then runs from remaining cards. */
function findMeldsSetsFirst(hand: Card[]): number[][] {
  const melds: number[][] = [];
  const used = new Set<number>();

  // Sets first
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

  // Runs from remaining
  findRuns(hand, used, melds);
  return melds;
}

/** Find melds prioritizing runs first, then sets from remaining cards. */
function findMeldsRunsFirst(hand: Card[]): number[][] {
  const melds: number[][] = [];
  const used = new Set<number>();

  // Runs first
  findRuns(hand, used, melds);

  // Sets from remaining
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

/** Find runs (3+ consecutive cards of same suit) from unused cards. Adds to melds and used set. */
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
