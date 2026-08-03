import type { PiquetResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PiquetPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Piquet, or null when no suggestion
 * is available.
 *
 * Piquet's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field, which carries **two mutually exclusive shapes**: a
 * `cardIndex` while playing tricks, and a `discardIndices` list while
 * exchanging with the talon. This adapter re-maps whichever one arrived into the
 * frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getPiquetHint(state: PiquetResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // **`cardIndex` は 0 が正当な手。**省略可能な数値なので、真偽値で見ると
  // 「手札の先頭を出せ」というヒントだけが黙って消える。
  if (state.phase === PiquetPhase.PLAY && hint.cardIndex !== undefined) {
    return { targetAction: 'play', reason: 'frontendHint.piquetPlayLowest', confidence: 'moderate' };
  }

  // 交換フェーズのヒントは捨て札の**枚数**が要る。空配列は「捨てるものがない」
  // であって推奨ではないので、ボタンを指しても行き先がない。
  if (state.phase === PiquetPhase.EXCHANGE && (hint.discardIndices?.length ?? 0) > 0) {
    return { targetAction: 'discard', reason: 'frontendHint.piquetExchangeLowest', confidence: 'moderate' };
  }

  return null;
}
