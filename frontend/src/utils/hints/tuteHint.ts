import type { TuteResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Tute, or null when no suggestion is
 * available.
 *
 * Like Doppelkopf and Sheepshead, Tute's hint is computed entirely by the Go
 * backend and surfaced on the response's `hint` field (with a `reason` i18n
 * suffix). This adapter re-maps that server hint into the frontend HintResult
 * shape so the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The
 * `targetAction` is fixed to `play` because the in-page tooltip only annotates
 * the play step.
 */
export function getTuteHint(state: TuteResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
