import type { CourtPieceResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Court Piece (Rang), or null when no
 * suggestion is available.
 *
 * Like Twenty-Nine, Court Piece's hint is computed entirely by the Go backend
 * and surfaced on the response's `hint` field (with a `reason` i18n suffix).
 * This adapter re-maps that server hint into the frontend HintResult shape so
 * the shared {@link useGameHint} tooltip can render it. The `targetAction` is
 * `trump` while a trump-suit declaration is suggested (the hint carries a
 * `trumpSuit`), and `play` otherwise (trick-play phase).
 */
export function getCourtPieceHint(state: CourtPieceResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.trumpSuit != null ? 'trump' : 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
