import type { RookResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Rook, or null when no suggestion is
 * available.
 *
 * The hint is computed by the Go backend and, since #4483, arrives on every
 * response rather than only the `hint` command's. This was a stub returning
 * null on the grounds that the reasoning lived server-side — true, and no
 * longer a reason to discard the answer once the server started sending it.
 *
 * The shapes are mutually exclusive, one per decision point, and each is tested
 * with `!== undefined` because zero is legal for all of them.
 */
export function getRookHint(state: RookResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  if (hint.cardIndex !== undefined) {
    return { targetAction: `card-${hint.cardIndex}`, reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.discardIndices !== undefined) {
    return { targetAction: 'discard', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.trumpColor !== undefined) {
    return { targetAction: 'trump', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  // `pass` は false で「パスするな」を意味するので `=== true` で見る。
  if (hint.pass === true) {
    return { targetAction: 'pass', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.bid !== undefined) {
    return { targetAction: 'bid', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  return null;
}
