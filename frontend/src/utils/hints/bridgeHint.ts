import type { BridgeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * 手が「決まっている」理由。リードスートに従える札が 1 枚しかない場面や、
 * 切り札で取りに行く場面は、選択の余地が小さいので確度を上げる。
 */
const DECISIVE = new Set(['trump_cut', 'lead_trump']);

/**
 * Returns a frontend {@link HintResult} for Bridge, or null when no suggestion
 * is available.
 *
 * Bridge's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field, which carries **two mutually exclusive shapes**: a
 * `cardIndex` during play, and a `bidType` / `bidLevel` / `bidSuit` triple
 * during the auction. This adapter re-maps whichever one arrived into the
 * frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 *
 * The reason keys live under `hintReason.` rather than this directory's usual
 * `frontendHint.` because the page's own hint line already reads them there.
 */
export function getBridgeHint(state: BridgeResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  const confidence = DECISIVE.has(hint.reason) ? 'strong' : 'moderate';

  // **札 0 は正当な手。**省略可能な数値なので、真偽値で見ると「手札の先頭を
  // 出せ」というヒントだけが黙って消える。パスの bidLevel 0 も同じ。
  if (hint.cardIndex !== undefined) {
    return { targetAction: `card-${hint.cardIndex}`, reason: `hintReason.${hint.reason}`, confidence };
  }
  if (hint.bidType !== undefined) {
    return { targetAction: 'bid', reason: `hintReason.${hint.reason}`, confidence };
  }
  return null;
}
