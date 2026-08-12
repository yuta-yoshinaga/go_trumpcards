import type { LingerLongerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Linger Longer, or null when no
 * suggestion is available.
 *
 * Taking a trick when the stock still has cards is the one move whose value is
 * certain — it refills your hand. Every other choice is a judgement call.
 */
export function getLingerLongerHint(state: LingerLongerResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'lingerlongerWinTrick' ? 'strong' : 'moderate',
  };
}
