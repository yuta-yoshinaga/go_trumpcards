import type { UnsunKarutaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Unsun Karuta, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain on the response's `hint` field with a
 * bare `reason` identifier (`lead_strong`, `follow_play`, `next_trick`,
 * `next_round`); this adapter re-maps it into the shared HintResult shape for
 * {@link hooks/useGameHint.useGameHint | useGameHint}.
 */
export function getUnsunKarutaHint(state: UnsunKarutaResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason || hint.reason === 'none') return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
