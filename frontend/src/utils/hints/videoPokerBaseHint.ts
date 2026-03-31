import type { Card, VideoPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { VideoPokerPhase } from '../../types/phases';

/** Minimum card value to be considered a high card (Jack). */
const HIGH_CARD_THRESHOLD = 11;

/** Minimum same-suit count to suggest holding a flush draw. */
const FLUSH_DRAW_COUNT = 4;

/** Minimum sequential count to suggest holding a straight draw. */
const STRAIGHT_DRAW_COUNT = 4;

/** Returns a frontend HintResult for Video Poker variants, or null if no suggestion available. */
export function getVideoPokerBaseHint(state: VideoPokerResponse, isWild: (card: Card) => boolean): HintResult | null {
  if (state.phase !== VideoPokerPhase.DRAW) return null;

  const hand = state.hand;
  if (!hand || hand.length === 0) return null;

  const wildIndices = hand.map((c, i) => (isWild(c) ? i : -1)).filter((i) => i >= 0);
  if (wildIndices.length > 0) {
    return getWildAwareHint(hand, wildIndices, isWild);
  }

  return getStandardHint(hand);
}

/** Hint when wild cards are present: always hold wilds, then analyze remaining. */
function getWildAwareHint(hand: Card[], wildIndices: number[], isWild: (card: Card) => boolean): HintResult {
  const pairResult = findPairHold(hand, isWild);
  if (pairResult) {
    return {
      targetAction: formatHoldAction(wildIndices, pairResult.indices),
      reason: 'hint.holdWild',
      confidence: 'strong',
    };
  }
  return { targetAction: formatHoldAction(wildIndices, []), reason: 'hint.holdWild', confidence: 'strong' };
}

/** Hint for standard (no wild) video poker hands. */
function getStandardHint(hand: Card[]): HintResult {
  // Check for existing pairs/trips/quads
  const groups = groupByValue(hand);
  const best = getBestGroup(groups);

  if (best.count >= 4) {
    return { targetAction: formatHoldAction(best.indices, []), reason: 'hint.holdQuads', confidence: 'strong' };
  }
  if (best.count >= 3) {
    return { targetAction: formatHoldAction(best.indices, []), reason: 'hint.holdTrips', confidence: 'strong' };
  }
  if (best.count >= 2) {
    return { targetAction: formatHoldAction(best.indices, []), reason: 'hint.holdPair', confidence: 'moderate' };
  }

  // Check flush draw (4+ same suit)
  const flushDraw = findFlushDraw(hand);
  if (flushDraw) {
    return { targetAction: formatHoldAction(flushDraw, []), reason: 'hint.holdFlushDraw', confidence: 'moderate' };
  }

  // Check straight draw (4+ sequential)
  const straightDraw = findStraightDraw(hand);
  if (straightDraw) {
    return {
      targetAction: formatHoldAction(straightDraw, []),
      reason: 'hint.holdStraightDraw',
      confidence: 'moderate',
    };
  }

  // Hold high cards
  const highCards = hand.map((c, i) => (c.value >= HIGH_CARD_THRESHOLD ? i : -1)).filter((i) => i >= 0);
  if (highCards.length > 0) {
    return { targetAction: formatHoldAction(highCards, []), reason: 'hint.holdHighCards', confidence: 'moderate' };
  }

  return { targetAction: 'draw-all', reason: 'hint.drawAll', confidence: 'moderate' };
}

/** Group cards by value, returning indices for each group. */
function groupByValue(hand: Card[]): Map<number, number[]> {
  const groups = new Map<number, number[]>();
  for (let i = 0; i < hand.length; i++) {
    const v = hand[i].value;
    const arr = groups.get(v) ?? [];
    arr.push(i);
    groups.set(v, arr);
  }
  return groups;
}

/** Find the largest group of same-value cards. */
function getBestGroup(groups: Map<number, number[]>): { count: number; indices: number[] } {
  let best = { count: 0, indices: [] as number[] };
  for (const indices of groups.values()) {
    if (indices.length > best.count) {
      best = { count: indices.length, indices };
    }
  }
  return best;
}

/** Find 4+ cards of the same suit (flush draw). */
function findFlushDraw(hand: Card[]): number[] | null {
  const suits = new Map<string, number[]>();
  for (let i = 0; i < hand.length; i++) {
    const d = hand[i].design;
    const arr = suits.get(d) ?? [];
    arr.push(i);
    suits.set(d, arr);
  }
  for (const indices of suits.values()) {
    if (indices.length >= FLUSH_DRAW_COUNT) return indices;
  }
  return null;
}

/** Find 4+ sequential card values (straight draw). */
function findStraightDraw(hand: Card[]): number[] | null {
  const sorted = hand.map((c, i) => ({ value: c.value, index: i })).sort((a, b) => a.value - b.value);
  for (let start = 0; start <= sorted.length - STRAIGHT_DRAW_COUNT; start++) {
    const seq = [sorted[start]];
    for (let j = start + 1; j < sorted.length && seq.length < STRAIGHT_DRAW_COUNT; j++) {
      if (sorted[j].value === seq[seq.length - 1].value + 1) {
        seq.push(sorted[j]);
      }
    }
    if (seq.length >= STRAIGHT_DRAW_COUNT) {
      return seq.map((s) => s.index);
    }
  }
  return null;
}

/** Find pairs among non-wild cards and return their original indices. */
function findPairHold(hand: Card[], isWild: (card: Card) => boolean): { indices: number[] } | null {
  const groups = new Map<number, number[]>();
  for (let i = 0; i < hand.length; i++) {
    if (isWild(hand[i])) continue;
    const v = hand[i].value;
    const arr = groups.get(v) ?? [];
    arr.push(i);
    groups.set(v, arr);
  }
  let best: number[] = [];
  for (const indices of groups.values()) {
    if (indices.length > best.length) best = indices;
  }
  return best.length >= 2 ? { indices: best } : null;
}

/** Format hold action as "hold:0,1,3" string. */
function formatHoldAction(primary: number[], secondary: number[]): string {
  const all = [...new Set([...primary, ...secondary])].sort((a, b) => a - b);
  return `hold:${all.join(',')}`;
}
