import type { EcarteResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Écarté, or null when no suggestion
 * is available.
 *
 * Like Bezique, Écarté's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix). This
 * adapter re-maps that server hint into the frontend HintResult shape so the
 * shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The hint carries a
 * `cardIndex` during the Play phase, and an `action` string (e.g. `propose`,
 * `stand`, `accept`, `refuse`, `discard`) during the Exchange phase. The
 * `targetAction` is the exchange action while one is suggested, and `play`
 * otherwise.
 */
export function getEcarteHint(state: EcarteResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.action ?? 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
