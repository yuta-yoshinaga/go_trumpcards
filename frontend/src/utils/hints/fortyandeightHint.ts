import type { FortyAndEightResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for FortyAndEight, or null when no suggestion is
 * available.
 *
 * The move is computed by the Go backend and, since #4483, arrives on every
 * response rather than only the `hint` command's. This was a stub returning
 * null on the grounds that the reasoning lived server-side — true, and no
 * longer a reason to discard the answer once the server started sending it.
 */
export function getFortyAndEightHint(state: FortyAndEightResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。ファウンデーション
  // など列を持たないゾーンは -1 で届くので、そこはゾーン名だけにする。
  const target = hint.toCol >= 0 ? `${hint.toZone}-${hint.toCol}` : hint.toZone;
  return { targetAction: target, reason: 'frontendHint.fortyandeightMove', confidence: 'moderate' };
}
