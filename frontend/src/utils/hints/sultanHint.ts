import type { SultanResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Sultan, or null when no suggestion is
 * available.
 *
 * The move is computed by the Go backend and, since #4483, arrives on every
 * response rather than only the `hint` command's. This was a stub returning
 * null on the grounds that the reasoning lived server-side — true, and no
 * longer a reason to discard the answer once the server started sending it.
 */
export function getSultanHint(state: SultanResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // **ファウンデーション 0 は正当。**真偽値で見ると先頭だけ落ちる。
  return {
    targetAction: `foundation-${hint.toFoundation}`,
    reason: 'frontendHint.sultanMove',
    confidence: 'moderate',
  };
}
