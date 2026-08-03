import type { AluetteResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Aluette, or null when no
 * suggestion is available.
 *
 * The hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field with a `reason` i18n suffix. It carries more weight
 * here than in most trick-takers because strength is **per card, not per
 * value**: the 3 of coins is the best card in the deck while the 3 of swords
 * is ordinary, so a player cannot rank a hand by eye until the six luettes are
 * memorised. This adapter re-maps the server hint into the frontend
 * HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 */
export function getAluetteHint(state: AluetteResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
