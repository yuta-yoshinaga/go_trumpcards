import type { Card } from '../types/card';

/**
 * Client-side port of Marriage declaration validation, mirroring
 * `MarriageValidateDeclaration`, `MarriageHasPureSequence`, and
 * `MarriageDeadwoodScore` in `internal/domain/Marriage.go`. Used to preview,
 * before the player declares, whether their arranged 21-card hand forms a valid
 * declaration (>=3 pure sequences, every card melded). The backend
 * remains the source of truth; this only powers a non-blocking warning.
 */

/** Marriage hand size validated on declaration (21, after the finish card). */
export const MARRIAGE_HAND_SIZE = 21;
/** Minimum length of a sequence (run). Mirrors `MarriageSeqMin`. */
const SEQ_MIN = 3;
/** Full deadwood penalty; an invalid declaration scores this flat. Mirrors `MarriageDeadwoodCap`. */
export const MARRIAGE_DEADWOOD_CAP = 80;
/** DFS iteration ceiling. Mirrors `marriageSearchCap`. */
const SEARCH_CAP = 2_000_000;
/** Cartesian product ceiling. Mirrors the domain's `maxCartesian`. */
const CARTESIAN_CAP = 256;

/** A candidate meld: original hand indices plus its sequence / pure flags. */
interface Meld {
  idx: number[];
  seq: boolean;
  pure: boolean;
}

/** Whether a card is tiplu: the same rank and suit as the turned-up tiplu. */
export function marriageIsTiplu(card: Card, tiplu: Card): boolean {
  return card.design === tiplu.design && card.value === tiplu.value;
}

/** Whether a card is poplu: the rank above tiplu, wrapping K to A. */
export function marriageIsPoplu(card: Card, tiplu: Card): boolean {
  return card.value === (tiplu.value % 13) + 1;
}

/** Whether a card is jhiplu: the rank below tiplu, wrapping A to K. */
export function marriageIsJhiplu(card: Card, tiplu: Card): boolean {
  return card.value === ((tiplu.value + 11) % 13) + 1;
}

/** Whether a card is alter: another suit of tiplu's rank. */
export function marriageIsAlter(card: Card, tiplu: Card): boolean {
  return card.value === tiplu.value && card.design !== tiplu.design;
}

/** Whether a card is wild: a printed joker, tiplu, poplu, jhiplu, or alter. */
export function marriageIsWild(card: Card, wildRank: number): boolean {
  if (card.design === 'JOKER') return true;
  if (wildRank === 0) return false;
  return card.value === wildRank || card.value === (wildRank % 13) + 1 || card.value === ((wildRank + 11) % 13) + 1;
}

/** Deadwood points for a card (wild=0, A=10, 2-9=pip, 10/J/Q/K=10). */
export function marriageCardPoints(card: Card, wildRank: number): number {
  if (marriageIsWild(card, wildRank)) return 0;
  const v = card.value;
  if (v === 1) return 10; // Ace
  if (v >= 10) return 10; // 10, J, Q, K
  return v;
}

/** Enumerate index combinations of size `k` from `idxs` with all-distinct suits. */
function distinctSuitCombos(idxs: number[], cards: readonly Card[], k: number): number[][] {
  const bySuit = new Map<string, number[]>();
  const order: string[] = [];
  for (const i of idxs) {
    const s = cards[i].design;
    if (!bySuit.has(s)) order.push(s);
    const list = bySuit.get(s) ?? [];
    list.push(i);
    bySuit.set(s, list);
  }
  const res: number[][] = [];
  const rec = (pos: number, cur: number[]): void => {
    if (cur.length === k) {
      res.push([...cur]);
      return;
    }
    if (pos >= order.length || order.length - pos < k - cur.length) return;
    // Skip this suit.
    rec(pos + 1, cur);
    // Use one card of this suit.
    for (const i of bySuit.get(order[pos]) ?? []) {
      cur.push(i);
      rec(pos + 1, cur);
      cur.pop();
    }
  };
  rec(0, []);
  return res;
}

/** Enumerate set candidates (same rank, distinct suits, 3-4 cards, <=1 wild). */
function generateSets(cards: readonly Card[], wildRank: number, wildIdxs: number[]): Meld[] {
  const byRank = new Map<number, number[]>();
  for (let i = 0; i < cards.length; i++) {
    if (marriageIsWild(cards[i], wildRank)) continue;
    const list = byRank.get(cards[i].value) ?? [];
    list.push(i);
    byRank.set(cards[i].value, list);
  }
  const melds: Meld[] = [];
  for (const idxs of byRank.values()) {
    // Pure sets (no wild), 3 and 4 cards.
    for (const k of [3, 4]) {
      for (const combo of distinctSuitCombos(idxs, cards, k)) {
        melds.push({ idx: combo, seq: false, pure: false });
      }
    }
    // Sets including one wild (3 and 4 total).
    for (const total of [3, 4]) {
      const naturals = total - 1;
      for (const combo of distinctSuitCombos(idxs, cards, naturals)) {
        for (const w of wildIdxs) {
          melds.push({ idx: [...combo, w], seq: false, pure: false });
        }
      }
    }
  }
  return melds;
}

/** Build run candidates for one suit over the window `[start, start+length-1]`. */
function buildRunWindow(
  byVal: Map<number, number[]>,
  start: number,
  length: number,
  wildIdxs: number[],
): Meld[] | null {
  const present: number[][] = [];
  let missing = 0;
  const seen = new Set<number>();
  for (let v = start; v < start + length; v++) {
    let lv = v;
    if (lv === 14) lv = 1; // Ace-high
    if (seen.has(lv)) return null; // window referencing a rank twice (wrap-around) is invalid
    seen.add(lv);
    const opts = byVal.get(lv) ?? [];
    if (opts.length === 0) {
      missing++;
      present.push([]);
    } else {
      present.push(opts);
    }
  }
  if (missing === 0) {
    return cartesian(present).map((combo) => ({ idx: combo, seq: true, pure: true }));
  }
  if (missing === 1 && wildIdxs.length > 0) {
    const base = present.filter((opts) => opts.length > 0);
    const combos = cartesian(base);
    const out: Meld[] = [];
    for (const combo of combos) {
      for (const w of wildIdxs) {
        out.push({ idx: [...combo, w], seq: true, pure: false });
      }
    }
    return out;
  }
  return null;
}

/** Enumerate run candidates (same suit, consecutive 3+, <=1 wild; Ace low & high). */
function generateRuns(cards: readonly Card[], wildRank: number, wildIdxs: number[]): Meld[] {
  const bySuit = new Map<string, Map<number, number[]>>();
  for (let i = 0; i < cards.length; i++) {
    if (marriageIsWild(cards[i], wildRank)) continue;
    const s = cards[i].design;
    let byVal = bySuit.get(s);
    if (!byVal) {
      byVal = new Map();
      bySuit.set(s, byVal);
    }
    const list = byVal.get(cards[i].value) ?? [];
    list.push(i);
    byVal.set(cards[i].value, list);
  }
  const melds: Meld[] = [];
  for (const byVal of bySuit.values()) {
    for (let start = 1; start <= 13; start++) {
      for (let length = SEQ_MIN; start + length - 1 <= 14; length++) {
        const window = buildRunWindow(byVal, start, length, wildIdxs);
        if (window) melds.push(...window);
      }
    }
  }
  return melds;
}

/** Cartesian product picking one index per list, capped to avoid blow-up. */
function cartesian(lists: number[][]): number[][] {
  let res: number[][] = [[]];
  for (const opts of lists) {
    if (opts.length === 0) continue;
    const next: number[][] = [];
    for (const prefix of res) {
      for (const o of opts) next.push([...prefix, o]);
    }
    res = next.length > CARTESIAN_CAP ? next.slice(0, CARTESIAN_CAP) : next;
  }
  return res;
}

/** Enumerate every valid set / sequence candidate from `cards`. */
function generateMelds(cards: readonly Card[], wildRank: number): Meld[] {
  const wildIdxs: number[] = [];
  for (let i = 0; i < cards.length; i++) {
    if (marriageIsWild(cards[i], wildRank)) wildIdxs.push(i);
  }
  return [...generateSets(cards, wildRank, wildIdxs), ...generateRuns(cards, wildRank, wildIdxs)];
}

/** For each card index, the list of meld indices that cover it. */
function covering(n: number, melds: Meld[]): number[][] {
  const cov: number[][] = Array.from({ length: n }, () => []);
  for (let mi = 0; mi < melds.length; mi++) {
    for (const ci of melds[mi].idx) cov[ci].push(mi);
  }
  return cov;
}

/**
 * Whether `cards` (21 cards) form a valid declaration: every card melded, with
 * at least three pure sequences. Mirrors
 * `MarriageValidateDeclaration`.
 */
export function marriageValidateDeclaration(cards: readonly Card[], wildRank: number): boolean {
  const n = cards.length;
  if (n !== MARRIAGE_HAND_SIZE) return false;
  const melds = generateMelds(cards, wildRank);
  const cov = covering(n, melds);
  const decided = new Array<boolean>(n).fill(false);
  let iter = 0;

  const firstUndecided = (): number => {
    for (let i = 0; i < n; i++) if (!decided[i]) return i;
    return -1;
  };
  const allUndecided = (idx: number[]): boolean => idx.every((ci) => !decided[ci]);
  const setDecided = (idx: number[], v: boolean): void => {
    for (const ci of idx) decided[ci] = v;
  };

  const dfs = (seq: number, pure: number): boolean => {
    iter++;
    if (iter > SEARCH_CAP) return false;
    const i = firstUndecided();
    if (i === -1) return pure >= 3;
    for (const mi of cov[i]) {
      const m = melds[mi];
      if (!allUndecided(m.idx)) continue;
      setDecided(m.idx, true);
      let si = 0;
      let pi = 0;
      if (m.seq) {
        si = 1;
        if (m.pure) pi = 1;
      }
      if (dfs(seq + si, pure + pi)) return true;
      setDecided(m.idx, false);
    }
    return false;
  };
  return dfs(0, 0);
}

/** Whether `cards` contain a pure sequence (no wild, same-suit run of 3+). */
export function marriageHasPureSequence(cards: readonly Card[], wildRank: number): boolean {
  return generateMelds(cards, wildRank).some((m) => m.seq && m.pure);
}

/** A minimum-deadwood split: total deadwood points plus the melded indices. */
interface DeadwoodSplit {
  points: number;
  melded: Set<number>;
}

/** Minimum-deadwood partition of `cards` into disjoint melds. Mirrors `marriageMinDeadwood`. */
function minDeadwoodSplit(cards: readonly Card[], wildRank: number): DeadwoodSplit {
  const n = cards.length;
  if (n === 0) return { points: 0, melded: new Set() };
  const melds = generateMelds(cards, wildRank);
  const cov = covering(n, melds);
  const points = cards.map((c) => marriageCardPoints(c, wildRank));
  const decided = new Array<boolean>(n).fill(false);
  let iter = 0;

  const firstUndecided = (): number => {
    for (let i = 0; i < n; i++) if (!decided[i]) return i;
    return -1;
  };
  const allUndecided = (idx: number[]): boolean => idx.every((ci) => !decided[ci]);
  const setDecided = (idx: number[], v: boolean): void => {
    for (const ci of idx) decided[ci] = v;
  };

  const dfs = (): DeadwoodSplit => {
    iter++;
    if (iter > SEARCH_CAP) {
      let s = 0;
      for (let k = 0; k < n; k++) if (!decided[k]) s += points[k];
      return { points: s, melded: new Set() };
    }
    const i = firstUndecided();
    if (i === -1) return { points: 0, melded: new Set() };
    // Option A: card i is deadwood.
    decided[i] = true;
    const subA = dfs();
    decided[i] = false;
    let best: DeadwoodSplit = { points: points[i] + subA.points, melded: subA.melded };
    // Option B: cover i with a meld.
    for (const mi of cov[i]) {
      const m = melds[mi];
      if (!allUndecided(m.idx)) continue;
      setDecided(m.idx, true);
      const sub = dfs();
      setDecided(m.idx, false);
      if (sub.points < best.points) {
        const melded = new Set<number>(sub.melded);
        for (const ci of m.idx) melded.add(ci);
        best = { points: sub.points, melded };
      }
    }
    return best;
  };
  return dfs();
}

/**
 * Deadwood score for `cards`: 80 (full cap) if there is no pure sequence;
 * otherwise the minimum deadwood capped at 80. Mirrors `MarriageDeadwoodScore`.
 */
export function marriageDeadwoodScore(cards: readonly Card[], wildRank: number): number {
  if (!marriageHasPureSequence(cards, wildRank)) return MARRIAGE_DEADWOOD_CAP;
  const dw = minDeadwoodSplit(cards, wildRank).points;
  return dw > MARRIAGE_DEADWOOD_CAP ? MARRIAGE_DEADWOOD_CAP : dw;
}

/** Preview of a Marriage declaration for a 21-card arranged hand. */
export interface MarriageDeclarePreview {
  /** Whether the hand is a valid declaration (backend still re-validates). */
  valid: boolean;
  /** Whether the hand contains at least one pure sequence. */
  hasPureSequence: boolean;
  /** Number of cards left unmelded in the best (minimum-deadwood) split. */
  unmeldedCount: number;
  /** Deadwood points of the unmelded cards in that best split. */
  unmeldedPoints: number;
  /** Penalty the player would take on this declaration (0 if valid, else 80). */
  penalty: number;
}

/**
 * Evaluate whether the given arranged 21-card hand forms a valid Marriage
 * declaration, returning both the validity verdict and human-friendly context
 * (missing pure sequence, unmelded card count / points, projected penalty).
 */
export function evaluateMarriageDeclare(cards: readonly Card[], wildRank: number): MarriageDeclarePreview {
  const valid = marriageValidateDeclaration(cards, wildRank);
  const hasPureSequence = marriageHasPureSequence(cards, wildRank);
  const split = minDeadwoodSplit(cards, wildRank);
  return {
    valid,
    hasPureSequence,
    unmeldedCount: cards.length - split.melded.size,
    unmeldedPoints: split.points,
    penalty: valid ? 0 : MARRIAGE_DEADWOOD_CAP,
  };
}
