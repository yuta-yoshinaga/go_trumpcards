import type { GanjifaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Ganjifa, or null when no
 * suggestion is available.
 *
 * Like Préférence, Ganjifa's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix) —
 * which matters more here than elsewhere, because a correct suggestion has to
 * account for the strong/weak suit inversion. This adapter re-maps that server
 * hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 * The `targetAction` is fixed to `play` because Ganjifa has no other action.
 */
export function getGanjifaHint(state: GanjifaResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
