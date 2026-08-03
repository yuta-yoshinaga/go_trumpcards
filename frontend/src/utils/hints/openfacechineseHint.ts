import type { OpenFaceChineseResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 段番号から段名へ。0=フロント / 1=ミドル / 2=バック。 */
const ROW_NAMES: Record<number, string> = { 0: 'front', 1: 'middle', 2: 'back' };

/**
 * Returns a frontend {@link HintResult} for Open Face Chinese Poker, or null
 * when no suggestion is available.
 *
 * The hint is computed entirely by the Go backend and surfaced on the response's
 * `hint` field, which names the row to place the pending card into plus a
 * `reason` i18n suffix. This adapter re-maps it into the frontend HintResult
 * shape so the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 *
 * `balance` is the backend's default branch — it fires when neither "strong card
 * to the back" nor "weak card to the front" applies — so it is reported at a
 * lower confidence than the two cases the server actually reasoned about.
 */
export function getOpenFaceChineseHint(state: OpenFaceChineseResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // **段 0 はフロント。**真偽値で見るとフロントへのヒントだけが消える。
  const row = ROW_NAMES[hint.row];
  if (!row) return null;

  return {
    targetAction: row,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'balance' ? 'moderate' : 'strong',
  };
}
