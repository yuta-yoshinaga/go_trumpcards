import type { SheepsheadResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Sheepshead, or null when no
 * suggestion is available.
 *
 * Unlike most games, Sheepshead's hint is computed entirely by the Go backend
 * and surfaced on the response's `hint` field (with a `reason` i18n suffix).
 * This adapter simply re-maps that server hint into the frontend HintResult
 * shape so the shared {@link useGameHint} tooltip can render it. The
 * `targetAction` is fixed to `play` because the in-page tooltip only annotates
 * the play step; the richer per-phase hint (pick/call) is shown via the
 * dedicated server-hint banner on the page.
 */
export function getSheepsheadHint(state: SheepsheadResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
