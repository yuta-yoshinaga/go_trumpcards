import type { BeziqueResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Bezique, or null when no
 * suggestion is available.
 *
 * Like Court Piece, Bezique's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix). This
 * adapter re-maps that server hint into the frontend HintResult shape so the
 * shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The hint carries a
 * `meldIndex` during the Meld phase (the special value -1 means "skip the
 * meld"), and a `cardIndex` during the Play phase. The `targetAction` is `meld`
 * while a meld declaration / skip is suggested and `play` otherwise.
 */
export function getBeziqueHint(state: BeziqueResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.meldIndex != null ? 'meld' : 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
