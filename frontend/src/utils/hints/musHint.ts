import type { MusResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Mus, or null when no suggestion is
 * available.
 *
 * Like Sheepshead, Mus computes its hint entirely on the Go backend and surfaces
 * it on the response's `hint` field with a `reason` i18n suffix. This adapter
 * re-maps that server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is derived from
 * the reason family (mus / discard / bet) so the in-page tooltip can annotate
 * the matching control group.
 */
export function getMusHint(state: MusResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;

  let targetAction = 'bet';
  if (hint.reason.startsWith('mus_')) targetAction = 'mus';
  else if (hint.reason.startsWith('discard_')) targetAction = 'discard';

  return {
    targetAction,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
