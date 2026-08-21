import type { RankAndFileResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for RankAndFile, or null when no suggestion is
 * available.
 *
 * The move is computed by the Go backend and, since #4483, arrives on every
 * response rather than only the `hint` command's. This was a stub returning
 * null on the grounds that the reasoning lived server-side — true, and no
 * longer a reason to discard the answer once the server started sending it.
 */
export function getRankAndFileHint(state: RankAndFileResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 盤上に手が無くストックだけ残っている局面 (#5525)。行き詰まりではないので
  // 「引け」と言う。移動の体裁に落とすと列を持たない waste--1 が出てしまう。
  if (hint.fromZone === 'stock') {
    return { targetAction: 'draw', reason: 'frontendHint.rankandfileDraw', confidence: 'moderate' };
  }

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。ファウンデーション
  // など列を持たないゾーンは -1 で届くので、そこはゾーン名だけにする。
  const target = hint.toCol >= 0 ? `${hint.toZone}-${hint.toCol}` : hint.toZone;
  return { targetAction: target, reason: 'frontendHint.rankandfileMove', confidence: 'moderate' };
}
