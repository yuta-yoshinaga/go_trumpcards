import type { StHelenaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for StHelena, or null when no
 * suggestion is available.
 *
 * The move is computed by the Go backend and, since #4483, arrives on every
 * response rather than only the `hint` command's. This was a stub returning
 * null on the grounds that the reasoning lived server-side — true, and no
 * longer a reason to discard the answer once the server started sending it.
 */
export function getStHelenaHint(state: StHelenaResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 再配りは盤面の手ではないので、行き先を名指ししない。
  if (hint.redeal) {
    return { targetAction: 'redeal', reason: 'frontendHint.stHelenaRedeal', confidence: 'moderate' };
  }

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。
  return {
    targetAction: `${hint.toZone}-${hint.toCol}`,
    reason: 'frontendHint.stHelenaMove',
    confidence: 'moderate',
  };
}
