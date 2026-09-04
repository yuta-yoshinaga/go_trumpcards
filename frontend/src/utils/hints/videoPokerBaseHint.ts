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
export function getVideoPokerBaseHint(
  state: VideoPokerResponse,
  isWild: (card: Card) => boolean,
  /**
   * Lowest rank whose pair actually pays in this variant, or null when no pair
   * pays at all. Deuces Wild stops at three of a kind, so recommending a high
   * pair there is a hold worth nothing -- and it outranks a four-card royal
   * (#6301). The caller supplies it because only it knows the paytable.
   */
  payingPairRank: number | null = HIGH_CARD_THRESHOLD,
): HintResult | null {
  if (state.phase !== VideoPokerPhase.DRAW) return null;

  const hand = state.hand;
  if (!hand || hand.length === 0) return null;

  const wildIndices = hand.map((c, i) => (isWild(c) ? i : -1)).filter((i) => i >= 0);
  if (wildIndices.length > 0) {
    return getWildAwareHint(hand, wildIndices, isWild);
  }

  return getStandardHint(hand, payingPairRank);
}

/** Hint when wild cards are present: always hold wilds, then analyze remaining. */
function getWildAwareHint(hand: Card[], wildIndices: number[], isWild: (card: Card) => boolean): HintResult {
  const pairResult = findPairHold(hand, isWild);
  if (pairResult) {
    return {
      targetAction: formatHoldAction(wildIndices, pairResult.indices),
      reason: 'hint.holdWildAndPair',
      confidence: 'strong',
    };
  }
  return { targetAction: formatHoldAction(wildIndices, []), reason: 'hint.holdWild', confidence: 'strong' };
}

/** Hint for standard (no wild) video poker hands. */
function getStandardHint(hand: Card[], payingPairRank: number | null): HintResult {
  // Check for existing pairs/trips/quads — hold all groups of 2+
  const groups = groupByValue(hand);
  const allGroupIndices = Array.from(groups.values())
    .filter((indices) => indices.length >= 2)
    .flat();

  const maxCount = allGroupIndices.length > 0 ? Math.max(...Array.from(groups.values()).map((g) => g.length)) : 0;
  // **配当のつくペアかどうかで扱いが変わる。**Jacks or Better では J 未満の
  // ペア単体には配当が無い。以前はこの分岐が無条件に最初へ来ていたため、
  // 4枚ロイヤルや4枚フラッシュが同居していても常に弱いペアを勧めていた (#4691)。
  const isPayingGroup = maxCount >= 3 || (maxCount === 2 && hasPayingPair(groups, payingPairRank));

  if (allGroupIndices.length > 0 && isPayingGroup) {
    const reason = maxCount >= 4 ? 'hint.holdQuads' : maxCount >= 3 ? 'hint.holdTrips' : 'hint.holdPair';
    return {
      targetAction: formatHoldAction(allGroupIndices, []),
      reason,
      confidence: maxCount >= 3 ? 'strong' : 'moderate',
    };
  }

  // **4枚ロイヤルは低ペアより遥かに強い。**標準戦略ではフルハウスより上に来る。
  const royalDraw = findRoyalDraw(hand);
  if (royalDraw) {
    return { targetAction: formatHoldAction(royalDraw, []), reason: 'hint.holdRoyalDraw', confidence: 'strong' };
  }

  // Check flush draw (4+ same suit) -- 標準戦略で低ペアより上。
  const flushDraw = findFlushDraw(hand);
  if (flushDraw) {
    return { targetAction: formatHoldAction(flushDraw, []), reason: 'hint.holdFlushDraw', confidence: 'moderate' };
  }

  // **低ペアは4枚ストレートより上。**ここを「ドロー優先」で一括りにすると、
  // 標準戦略から外れる方向に壊れる。
  if (allGroupIndices.length > 0) {
    return {
      targetAction: formatHoldAction(allGroupIndices, []),
      reason: 'hint.holdPair',
      confidence: 'moderate',
    };
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

/** Find 4+ cards of the same suit (flush draw). */
/**
 * 配当のつくペア (J 以上、またはエース) を含むかどうか。
 *
 * `groupByValue` はカードの値をキーにするので、キーをそのまま見れば足りる。
 * エースは実装によって 1 と 14 のどちらでも来るため両方を受ける。
 */
function hasPayingPair(groups: Map<number, number[]>, payingPairRank: number | null): boolean {
  if (payingPairRank === null) return false;
  for (const [value, indices] of groups) {
    if (indices.length >= 2 && (value >= payingPairRank || value === 1 || value === ACE_HIGH)) {
      return true;
    }
  }
  return false;
}

/** 同一スートで 10-J-Q-K-A のうち 4 枚そろっているか。 */
function findRoyalDraw(hand: Card[]): number[] | null {
  const bySuit = new Map<string, number[]>();
  for (let i = 0; i < hand.length; i++) {
    const v = hand[i].value;
    const isRoyalRank = v >= 10 || v === 1 || v === ACE_HIGH;
    if (!isRoyalRank) continue;
    const arr = bySuit.get(hand[i].design) ?? [];
    arr.push(i);
    bySuit.set(hand[i].design, arr);
  }
  for (const indices of bySuit.values()) {
    // 同じランクの重複は royal を構成しない。
    const uniq = new Map<number, number>();
    for (const i of indices) uniq.set(hand[i].value, i);
    if (uniq.size >= FLUSH_DRAW_COUNT) return Array.from(uniq.values()).sort((a, b) => a - b);
  }
  return null;
}

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

/** Ace value used for high straights. */
const ACE_HIGH = 14;

/** Find 4+ sequential card values (straight draw), including Ace-low wheel (A-2-3-4-5). */
function findStraightDraw(hand: Card[]): number[] | null {
  const entries = hand.map((c, i) => ({ value: c.value, index: i }));

  // Add Ace as low (value 1) for wheel detection
  for (const e of hand) {
    if (e.value === ACE_HIGH) {
      entries.push({ value: 1, index: hand.indexOf(e) });
    }
  }

  const sorted = entries.sort((a, b) => a.value - b.value);
  let bestSeq: typeof sorted = [];

  for (let start = 0; start < sorted.length; start++) {
    const seq = [sorted[start]];
    for (let j = start + 1; j < sorted.length; j++) {
      if (sorted[j].value === seq[seq.length - 1].value + 1) {
        seq.push(sorted[j]);
      } else if (sorted[j].value !== seq[seq.length - 1].value) {
        break;
      }
    }
    if (seq.length > bestSeq.length) {
      bestSeq = seq;
    }
  }

  if (bestSeq.length >= STRAIGHT_DRAW_COUNT) {
    const indices = [...new Set(bestSeq.map((s) => s.index))];
    return indices;
  }
  return null;
}

/** Find all pairs/trips/quads among non-wild cards and return their indices. */
function findPairHold(hand: Card[], isWild: (card: Card) => boolean): { indices: number[] } | null {
  const groups = new Map<number, number[]>();
  for (let i = 0; i < hand.length; i++) {
    if (isWild(hand[i])) continue;
    const v = hand[i].value;
    const arr = groups.get(v) ?? [];
    arr.push(i);
    groups.set(v, arr);
  }
  const allIndices = Array.from(groups.values())
    .filter((indices) => indices.length >= 2)
    .flat();
  return allIndices.length >= 2 ? { indices: allIndices } : null;
}

/** Format hold action as "hold:0,1,3" string. */
function formatHoldAction(primary: number[], secondary: number[]): string {
  const all = [...new Set([...primary, ...secondary])].sort((a, b) => a - b);
  return `hold:${all.join(',')}`;
}
