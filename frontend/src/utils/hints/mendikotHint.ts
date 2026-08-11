import type { MendikotResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Mendikot, or null when no
 * suggestion is available.
 *
 * Every Mendikot hint names a card — there is no trump to declare, because the
 * suit is fixed by whichever card the first player who cannot follow plays.
 * Chasing a ten that is already on the table is the one near-automatic move;
 * ducking and feeding your partner are judgement calls.
 */
export function getMendikotHint(state: MendikotResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'mendikotChaseTen' ? 'strong' : 'moderate',
  };
}
