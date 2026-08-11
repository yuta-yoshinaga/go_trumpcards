import type { BhabhiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Bhabhi, or null when no suggestion
 * is available.
 *
 * Every hint names a card — there is nothing else to decide in this game.
 * Dumping a high card when you cannot follow is close to automatic (you take
 * the pile either way, so you may as well be rid of your worst card); leading
 * and ducking are judgement calls.
 */
export function getBhabhiHint(state: BhabhiResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'bhabhiDumpHigh' ? 'strong' : 'moderate',
  };
}
