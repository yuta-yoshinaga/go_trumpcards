import type { TarocchiniResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Tarocchini, or null when no
 * suggestion is available.
 *
 * The hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field with a `reason` i18n suffix. That matters more here
 * than in most games because a correct suggestion has to account for the papi
 * ranking equal and the later-played one winning — advice that reads backwards
 * if you assume the usual "strictly stronger card wins". This adapter re-maps
 * the server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getTarocchiniHint(state: TarocchiniResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
