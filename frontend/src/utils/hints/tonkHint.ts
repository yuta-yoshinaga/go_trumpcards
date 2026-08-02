import type { Card, TonkResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TonkPhase } from '../../types/phases';

/**
 * ノックできるデッドウッドの上限 (sync: internal/domain/Tonk.go:17,
 * `TonkKnockThreshold`). サーバはこの値を送ってこないので持ち直している。
 */
const KNOCK_THRESHOLD = 5;

/** メルドは 3 枚以上 (sync: `findAllPossibleMelds`, internal/domain/GinRummy.go)。 */
const MIN_MELD = 3;

/**
 * Returns a frontend {@link HintResult} for Tonk, or null when no suggestion is
 * available.
 *
 * There is no server-side GetHint here, so the deadwood search is done on the
 * client. It is the same search the server runs — sets of three or more of a
 * rank, runs of three or more in a suit, recursively minimised — ported from
 * `FindBestMelds` / `CalcDeadwoodValue` (`internal/domain/GinRummy.go`), which is
 * what `Tonk.PlayerKnock` validates against.
 *
 * **A pair is not a meld.** The first version of this file used the shallow
 * "connects with something" test the run-building rummies use, which counts a
 * pair as safe. That is not conservative here: it *under*-states deadwood, so a
 * hand like 2-2-K-K-Q reads as zero and the hint offers a knock the server
 * rejects with `ErrInvalidPlay`. Nothing shallow is safe in the direction that
 * matters, so the search is exact instead. A Tonk hand is five cards (six while
 * holding a draw), so the recursion is trivially small.
 */
export function getTonkHint(state: TonkResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const hand = human.cards;

  if (state.phase === TonkPhase.DRAW) {
    const top = state.discardTop;
    // 拾って減るかどうかで決める。サーバの `cpuDraw` (Tonk.go:356) と同じ判定。
    const improves = top !== null && deadwoodValue([...hand, top]) < deadwoodValue(hand);
    return improves
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.tonkTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.tonkDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== TonkPhase.DISCARD) return null;

  const best = bestDiscard(hand);

  // ノックは捨てたあとの手札で判定される (`PlayerKnock` は捨て札を除いてから数える)。
  if (best.deadwood <= KNOCK_THRESHOLD) {
    return { targetAction: 'knock', reason: 'frontendHint.tonkKnock', confidence: 'moderate' };
  }
  return { targetAction: `card-${best.index}`, reason: 'frontendHint.tonkDiscardHeavy', confidence: 'moderate' };
}

/** 捨てたあとのデッドウッドが最小になる札。同点なら手前の札。 */
function bestDiscard(hand: Card[]): { index: number; deadwood: number } {
  let index = 0;
  let deadwood = Number.POSITIVE_INFINITY;
  hand.forEach((_, i) => {
    const dw = deadwoodValue(hand.filter((_, j) => j !== i));
    if (dw < deadwood) {
      deadwood = dw;
      index = i;
    }
  });
  return { index, deadwood };
}

/** 札の点数。A は 1、10/J/Q/K は 10 (sync: `GinRummyCardValue`)。 */
function points(c: Card): number {
  return c.value >= 10 ? 10 : c.value;
}

/** メルドに使えなかった札の合計点。最小になる分け方を探す。 */
function deadwoodValue(hand: Card[]): number {
  const melds = possibleMelds(hand);
  let best = hand.reduce((sum, c) => sum + points(c), 0);
  for (const meld of melds) {
    const rest = hand.filter((c) => !meld.includes(c));
    const dw = deadwoodValue(rest);
    if (dw < best) best = dw;
    if (best === 0) break;
  }
  return best;
}

/**
 * 同ランク 3 枚以上のセットと、同スート連続 3 枚以上のラン。
 *
 * **ランの窓は左端固定にしてある。**サーバの `findAllPossibleMelds`
 * (`internal/domain/GinRummy.go:954`) は開始位置を動かして 3-4-5-6 から
 * `[4,5,6]` も作るが、こちらは `[3,4,5]` と `[3,4,5,6]` しか作らない。
 * これは意図的な制限で、ずれた窓は再帰の次段で先頭から作り直されるため
 * 結果は変わらない。
 *
 * 「変わらないはず」で済ませず総当たりで確認した (#4640 のレビュー指摘):
 * 全窓版と左端固定版のデッドウッドを比べて、フルデッキ 52 枚からの 5 枚
 * 2,598,960 通りと、36 枚からの 6 枚 1,947,792 通りで **差 0 件**。
 * 左端固定側から 3 枚ランを外す負のコントロールでは 124,654 件の差が出たので、
 * 比較自体が空振りしていないことも確かめてある。
 *
 * つまりこの制限は「サーバと同じ窓を作るように直す」必要がない。直しても
 * 答えは変わらず、探索が重くなるだけ。
 */
function possibleMelds(hand: Card[]): Card[][] {
  const melds: Card[][] = [];

  const byRank = new Map<number, Card[]>();
  for (const c of hand) byRank.set(c.value, [...(byRank.get(c.value) ?? []), c]);
  for (const group of byRank.values()) {
    if (group.length >= MIN_MELD) melds.push(group.slice(0, MIN_MELD));
    if (group.length > MIN_MELD) melds.push(group);
  }

  const bySuit = new Map<Card['design'], Card[]>();
  for (const c of hand) bySuit.set(c.design, [...(bySuit.get(c.design) ?? []), c]);
  for (const group of bySuit.values()) {
    const sorted = [...group].sort((a, b) => a.value - b.value);
    for (let start = 0; start < sorted.length; start += 1) {
      let end = start + 1;
      while (end < sorted.length && sorted[end].value === sorted[end - 1].value + 1) end += 1;
      for (let len = MIN_MELD; len <= end - start; len += 1) {
        melds.push(sorted.slice(start, start + len));
      }
      // 連続が切れた位置まで飛ばす。1 枚ずつ進めると同じ並びを何度も積む。
      start = end - 1;
    }
  }

  return melds;
}
