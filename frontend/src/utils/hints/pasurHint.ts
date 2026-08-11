import type { PasurResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Pasur, or null when no suggestion
 * is available.
 *
 * Emptying the table doubles that capture, so it is the one move worth calling
 * strong; an ordinary capture and a forced lay-down are both judgement calls.
 */
export function getPasurHint(state: PasurResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  // **取る場札まで含めて 1 つの手。** 札だけ指しても何を取るのか決まらない。
  const target =
    hint.table.length > 0 ? `card-${hint.cardIndex}-take-${hint.table.join('-')}` : `card-${hint.cardIndex}`;

  return {
    targetAction: target,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'pasurSoor' ? 'strong' : 'moderate',
  };
}
