import type { ViraResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Vira, or null when no suggestion is
 * available.
 *
 * The hint is computed by the Go backend (`Vira.GetHint`) and arrives on the
 * response's `hint` field with a `reason` suffix. This adapter re-maps it into
 * the frontend shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 *
 * `targetAction` is fixed to `play`: the server only produces a hint during the
 * trick-play phase (`GetHint` returns nil otherwise), so there is no bid-phase
 * case to distinguish.
 *
 * The reasons themselves carry the one thing that makes Vira's play different —
 * **under Misère the declarer and the defenders want opposite outcomes from the
 * same trick**, so `misere_duck` and `misere_force` are separate keys rather
 * than one "play your lowest" message.
 */
export function getViraHint(state: ViraResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
