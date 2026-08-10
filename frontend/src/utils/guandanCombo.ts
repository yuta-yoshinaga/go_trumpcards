import type { Card } from '../types/card';

/**
 * Guandan combo kinds, matching `domain.GuandanComboKind` and the numeric keys
 * of `COMBO_KEYS` in `GuandanPage.tsx` — the same numbers the server sends in
 * `lastCombo.kind`, so a previewed combo and the played one read alike.
 */
export const GUANDAN_COMBO = {
  None: 0,
  Single: 1,
  Pair: 2,
  Triple: 3,
  FullHouse: 4,
  Straight: 5,
  Plate: 6,
  Tube: 7,
  Bomb: 8,
  StraightFlush: 9,
  JokerBomb: 10,
} as const;

/** A combo the selected cards form. Mirrors `domain.GuandanCombo`. */
export interface GuandanComboEval {
  /** One of {@link GUANDAN_COMBO}. */
  kind: number;
  /** The rank used to compare against the combo on the table. */
  rank: number;
  /** How many cards the combo uses. */
  size: number;
}

const MIN_LEVEL = 2;
const RANK_ACE = 14;
const RANK_LEVEL = 15;
const RANK_BLACK_JOKER = 16;
const RANK_RED_JOKER = 17;

/** True when the card is either joker. */
function isJoker(c: Card): boolean {
  return c.design === 'JOKER';
}

/** The card's rank ignoring the level, jokers highest. Mirrors `guandanNaturalRank`. */
function naturalRank(c: Card): number {
  if (isJoker(c)) return c.value >= 2 ? RANK_RED_JOKER : RANK_BLACK_JOKER;
  if (c.value === 1) return RANK_ACE;
  return c.value;
}

/** True when the card is of this hand's level rank. Mirrors `GuandanIsLevelCard`. */
function isLevelCard(c: Card, level: number): boolean {
  return !isJoker(c) && naturalRank(c) === level;
}

/**
 * True when the card is wild. **Only the level card in hearts is** — two cards
 * per hand, since the game uses two decks. Mirrors `GuandanIsWild`.
 * @param c - The card to test.
 * @param level - This hand's level rank.
 * @returns Whether the card can stand in for another.
 */
export function guandanIsWild(c: Card, level: number): boolean {
  return isLevelCard(c, level) && c.design === 'HEART';
}

/**
 * The card's strength for this hand: the level card slots above the ace and
 * below the black joker. Mirrors `GuandanRank`.
 */
function rankOf(c: Card, level: number): number {
  if (isJoker(c)) return naturalRank(c);
  if (isLevelCard(c, level)) return RANK_LEVEL;
  return naturalRank(c);
}

/** Ascending ranks present in the tally, so the search is deterministic. */
function sortedRanks(counts: Map<number, number>): number[] {
  return [...counts.keys()].sort((a, b) => a - b);
}

/** Three of a rank plus a pair, wilds filling either. Mirrors `guandanFullHouse`. */
function fullHouse(counts: Map<number, number>, wilds: number, total: number): GuandanComboEval | null {
  if (total !== 5) return null;
  const ranks = sortedRanks(counts);
  for (const three of ranks) {
    const have3 = counts.get(three) ?? 0;
    const need3 = Math.max(3 - have3, 0);
    for (const two of ranks) {
      if (two === three) continue;
      const have2 = counts.get(two) ?? 0;
      const need2 = Math.max(2 - have2, 0);
      if (need3 + need2 <= wilds && have3 + have2 + wilds >= 5) {
        return { kind: GUANDAN_COMBO.FullHouse, rank: three, size: 5 };
      }
    }
  }
  return null;
}

/** True when the ascending values repeat. */
function hasDuplicate(vals: number[]): boolean {
  return vals.some((v, i) => i > 0 && v === vals[i - 1]);
}

/**
 * Five in sequence, **the ace usable at either end**. Level cards count from
 * their natural position here — cutting in above the ace is only about how
 * singles and pairs compare. Mirrors `guandanStraight`.
 */
function straight(fixed: Card[], wilds: number, total: number, _level: number): GuandanComboEval | null {
  if (total !== 5) return null;
  const base: number[] = [];
  let sameSuit = true;
  let suit: string | null = null;
  for (const c of fixed) {
    if (isJoker(c)) return null;
    base.push(c.value === 1 ? 14 : c.value);
    if (suit === null) suit = c.design;
    else if (suit !== c.design) sameSuit = false;
  }
  for (const low of [false, true]) {
    const vals = (low ? base.map((v) => (v === 14 ? 1 : v)) : [...base]).sort((a, b) => a - b);
    if (hasDuplicate(vals) || vals.length === 0) continue;
    const top = vals[vals.length - 1];
    const bottom = vals[0];
    if (top === undefined || bottom === undefined) continue;
    const span = top - bottom + 1;
    if (span > 5) continue;
    if (5 - vals.length <= wilds) {
      const kind = sameSuit && wilds === 0 ? GUANDAN_COMBO.StraightFlush : GUANDAN_COMBO.Straight;
      // **Spare wilds extend upward.** 5-6-7-8 plus a wild reads as either
      // 4-5-6-7-8 or 5-6-7-8-9; settling on the weaker one costs playable hands.
      return { kind, rank: Math.min(top + 5 - span, 14), size: 5 };
    }
  }
  return null;
}

/** `groups` consecutive ranks of `each` card. Mirrors `guandanRun`. */
function run(
  counts: Map<number, number>,
  wilds: number,
  groups: number,
  each: number,
  kind: number,
): GuandanComboEval | null {
  for (let start = MIN_LEVEL; start + groups - 1 <= 14; start++) {
    let need = 0;
    for (let i = 0; i < groups; i++) {
      need += each - Math.min(counts.get(start + i) ?? 0, each);
    }
    if (need <= wilds) {
      return { kind, rank: start + groups - 1, size: groups * each };
    }
  }
  return null;
}

/** The plate (two consecutive triples) and the tube (three consecutive pairs). */
function repeatedRun(counts: Map<number, number>, wilds: number, total: number): GuandanComboEval | null {
  if (total !== 6) return null;
  return run(counts, wilds, 2, 3, GUANDAN_COMBO.Plate) ?? run(counts, wilds, 3, 2, GUANDAN_COMBO.Tube);
}

/**
 * Classify a selection of cards, so the player can see what they hold before
 * the server judges it (#4901).
 *
 * A faithful port of `GuandanEvaluate` in `internal/domain/Guandan.go` — the
 * preview has to agree with the server, or it is worse than showing nothing.
 * @param cards - The selected cards, in any order.
 * @param level - This hand's level rank, from `state.level`.
 * @returns The combo they form, or null when they form none.
 */
export function guandanEvaluate(cards: readonly Card[], level: number): GuandanComboEval | null {
  if (cards.length === 0) return null;
  let wilds = 0;
  const fixed: Card[] = [];
  for (const c of cards) {
    if (guandanIsWild(c, level)) wilds++;
    else fixed.push(c);
  }

  // **Four jokers is the strongest hand there is,** and wilds cannot build it.
  if (cards.length === 4 && wilds === 0 && fixed.every(isJoker)) {
    return { kind: GUANDAN_COMBO.JokerBomb, rank: 100, size: 4 };
  }

  const counts = new Map<number, number>();
  for (const c of fixed) {
    const r = rankOf(c, level);
    counts.set(r, (counts.get(r) ?? 0) + 1);
  }
  // **Sequences count from the natural position.** At level 5, a 4-5-6 run of
  // pairs is legal; ranking the 5s at 15 would punch a hole in the window.
  const naturalCounts = new Map<number, number>();
  for (const c of fixed) {
    const r = naturalRank(c);
    naturalCounts.set(r, (naturalCounts.get(r) ?? 0) + 1);
  }

  if (counts.size <= 1) {
    // Wilds alone rank as low as possible.
    const rank = counts.size === 1 ? (sortedRanks(counts)[0] ?? MIN_LEVEL) : MIN_LEVEL;
    const n = cards.length;
    if (n === 1) return { kind: GUANDAN_COMBO.Single, rank, size: 1 };
    if (n === 2) return { kind: GUANDAN_COMBO.Pair, rank, size: 2 };
    if (n === 3) return { kind: GUANDAN_COMBO.Triple, rank, size: 3 };
    // **Four or more of a rank is a bomb,** and more cards beat fewer.
    return { kind: GUANDAN_COMBO.Bomb, rank, size: n };
  }

  return (
    fullHouse(counts, wilds, cards.length) ??
    straight(fixed, wilds, cards.length, level) ??
    repeatedRun(naturalCounts, wilds, cards.length)
  );
}

/** True when the combo outranks every non-bomb. Mirrors `guandanBombTier > 0`. */
export function guandanIsBomb(kind: number): boolean {
  return kind === GUANDAN_COMBO.Bomb || kind === GUANDAN_COMBO.StraightFlush || kind === GUANDAN_COMBO.JokerBomb;
}
