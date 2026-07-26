import type { GutsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Guts, or null when no suggestion is
 * available.
 *
 * Like Three Card Brag, Guts's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field. The hint carries a `declaration`
 * (0=out/fold, 1=in/stay) mapped to the `targetAction` string, and a `reason`
 * i18n suffix (`strong_hand` / `weak_hand`) re-mapped into the frontend
 * HintResult shape so the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getGutsHint(state: GutsResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.declaration === 1 ? 'in' : 'out',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
