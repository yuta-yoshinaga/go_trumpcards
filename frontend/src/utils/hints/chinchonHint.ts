import type { ChinchonResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ChinchonPhase } from '../../types/phases';
import { heaviestSpare, isMaterial } from './rummyHintShape';

/**
 * Returns a frontend {@link HintResult} for Chinchón, or null when no
 * suggestion is available.
 *
 * The response does not carry the human's melds — only the knocker's, and only
 * after the knock — so this does **not** try to prove a meld. It uses a shallow
 * "connects with something" test: same rank, or the neighbouring rank in the
 * same suit. That is enough for the two decisions a player makes every turn
 * (take the discard or not, and which card to throw) without re-implementing
 * the domain's meld detection, which would drift from it.
 */
export function getChinchonHint(state: ChinchonResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.eliminated || human.cards.length === 0) return null;
  if (state.currentPlayerIdx !== human.id) return null;

  if (state.phase === ChinchonPhase.DRAW) {
    const top = state.discardTop;
    return top && isMaterial(top, human.cards, chinchonAdjacent)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.chinchonTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.chinchonDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== ChinchonPhase.DISCARD) return null;

  // 点は手に残った札の合計なので、繋がっていない札のうち一番重いものを出す。
  // 全部繋がっているときも、崩すなら一番重い札。
  // **札 0 も捨て札になりうる。**手札が空でないことは上で確かめてあるので、
  // heaviestLoose は必ず 0 以上を返す。`idx < 0` を書くと到達しない分岐が残る。
  const idx = heaviestSpare(human.cards, chinchonAdjacent);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.chinchonDiscardHeavy', confidence: 'moderate' };
}

/**
 * 40 枚デッキ上の「隣接位置」。
 *
 * Chinchón は 8/9/10 を抜いた 40 枚を使うので、ランは A,2,…,7,J,Q,K が連続。
 * **7 と J は隣接する。**生の `value` を引き算すると `|11-7| = 4` になって
 * 繋がっていないと誤判定する (#4614 のレビュー指摘)。ドメインの
 * `chinchonRankPosition` (Chinchon.go:37) と同じ対応。
 */
function rankPosition(value: number): number {
  if (value >= 1 && value <= 7) return value;
  if (value === 11) return 8;
  if (value === 12) return 9;
  if (value === 13) return 10;
  return 0; // 8/9/10/Joker — このデッキには存在しない
}

/**
 * 40 枚デッキでの隣接判定。
 *
 * `rankPosition` に写してから隣り合うかを見る。**7 と J は隣接する** ——
 * 生の値では `|11-7| = 4` になり、繋がっていないと誤判定する。
 * 位置 0 はこのデッキに存在しない札 (8/9/10/Joker) なので繋がらない。
 */
function chinchonAdjacent(a: number, b: number): boolean {
  const pa = rankPosition(a);
  const pb = rankPosition(b);
  return pa > 0 && pb > 0 && Math.abs(pa - pb) === 1;
}
