import type { PigResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Pig, or null when no suggestion is
 * available.
 *
 * Once a signal is out there is exactly one right move and it is urgent, so
 * that advice is certain; which card to pass is a judgement call.
 */
export function getPigHint(state: PigResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 合図が出ている場面は選択の余地がない。遅れることだけが負け。
  if (hint.cardIndex === undefined) {
    return { targetAction: 'signal', reason: `hint.${hint.reason}`, confidence: 'strong' };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
