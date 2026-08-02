import type { Card, ConquianResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ConquianPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Conquian, or null when no
 * suggestion is available.
 *
 * The response does not carry the human's melds — only the knocker's, and only
 * after the knock — so this does **not** try to prove a meld. It uses a shallow
 * "connects with something" test: same rank, or the neighbouring rank in the
 * same suit. That is enough for the two decisions a player makes every turn
 * (take the discard or not, and which card to throw) without re-implementing
 * the domain's meld detection, which would drift from it.
 */
export function getConquianHint(state: ConquianResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;
  if (state.currentPlayerIdx !== human.id) return null;

  if (state.phase === ConquianPhase.DRAW) {
    const top = state.discardTop;
    return top && connects(top, human.cards)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.conquianTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.conquianDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== ConquianPhase.MELD) return null;

  // 点は手に残った札の合計なので、繋がっていない札のうち一番重いものを出す。
  // 全部繋がっているときも、崩すなら一番重い札。
  // **札 0 も捨て札になりうる。**手札が空でないことは上で確かめてあるので、
  // heaviestLoose は必ず 0 以上を返す。`idx < 0` を書くと到達しない分岐が残る。
  const idx = heaviestLoose(human.cards);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.conquianDiscardHeavy', confidence: 'moderate' };
}

/**
 * 40 枚デッキ上の「隣接位置」。
 *
 * Conquian は 8/9/10 を抜いた 40 枚を使うので、ランは A,2,…,7,J,Q,K が連続。
 * **7 と J は隣接する。**生の `value` を引き算すると `|11-7| = 4` になって
 * 繋がっていないと誤判定する (#4614 と同じ形 のレビュー指摘)。ドメインの
 * `conquianRankPosition` (Conquian.go:38) と同じ対応。
 */
function rankPosition(value: number): number {
  if (value >= 1 && value <= 7) return value;
  if (value === 11) return 8;
  if (value === 12) return 9;
  if (value === 13) return 10;
  return 0; // 8/9/10/Joker — このデッキには存在しない
}

/** 同じランクがあるか、同じスートで隣のランクがあるか。メルドの証明ではない。 */
function connects(c: Card, hand: Card[]): boolean {
  const pos = rankPosition(c.value);
  return hand.some(
    (o) =>
      o.value === c.value ||
      (o.design === c.design && pos > 0 && rankPosition(o.value) > 0 && Math.abs(rankPosition(o.value) - pos) === 1),
  );
}

/**
 * 繋がっていない札のうち一番重いものの位置。全部繋がっていれば一番重い札。
 *
 * 呼び出し側が手札の非空を確かめているので、必ず有効な位置を返す。
 */
function heaviestLoose(hand: Card[]): number {
  const loose: number[] = [];
  for (let i = 0; i < hand.length; i += 1) {
    const rest = hand.filter((_, j) => j !== i);
    if (!connects(hand[i], rest)) loose.push(i);
  }
  const pool = loose.length > 0 ? loose : hand.map((_, i) => i);
  let best = pool[0];
  for (const i of pool) {
    if (hand[i].value > hand[best].value) best = i;
  }
  return best;
}
