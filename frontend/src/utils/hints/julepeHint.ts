import type { JulepeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Julepe, or null when no suggestion is
 * available.
 *
 * The decision-phase hint carries no card index — it advises whether to enter
 * the round at all — so its `targetAction` names the decision rather than a
 * card. Banking that first trick is the one thing that removes the extra
 * payment, so it is reported as the stronger call.
 */
export function getJulepeHint(state: JulepeResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 選択フェーズは札を指さない。行動そのものが助言の対象。
  if (hint.cardIndex === undefined) {
    return {
      targetAction: hint.reason === 'julepePlayIn' ? 'play-in' : 'pass-out',
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'julepeTakeTrick' ? 'strong' : 'moderate',
  };
}
