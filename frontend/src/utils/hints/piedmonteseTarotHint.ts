import type { PiedmonteseTarotResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Tarocco Piemontese, or null when no
 * suggestion is available.
 *
 * The suggestion itself is computed by the Go domain and arrives on the
 * response's `hint` field with a bare `reason` identifier (`scarto_weak`,
 * `lead_low`, `follow_play`, `overtrump`, `next_trick`, `next_round`). This
 * adapter re-maps it into the shared HintResult shape so
 * {@link hooks/useGameHint.useGameHint | useGameHint} can render it; the
 * `targetAction` is fixed to `play` because every hint points at a decision.
 */
export function getPiedmonteseTarotHint(state: PiedmonteseTarotResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason || hint.reason === 'none') return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
