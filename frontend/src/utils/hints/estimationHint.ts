import type { EstimationResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Estimation, or null when no
 * suggestion is available.
 *
 * Neither the trump hint nor the call hint names a card, so both put the
 * recommended value into `targetAction` — the suit to make trump, or the
 * number to call. Ducking when you already have your call is close to
 * automatic; chasing a trick you still need is a judgement call.
 */
export function getEstimationHint(state: EstimationResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 切り札選択と宣言は札を指さない。値そのものが助言の対象。
  if (hint.cardIndex === undefined) {
    if (hint.reason === 'estimationSelectTrump') {
      return {
        targetAction: `trump-${hint.value}`,
        reason: `hint.${hint.reason}`,
        confidence: 'moderate',
      };
    }
    return {
      targetAction: `bid-${hint.value}`,
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'estimationDuck' ? 'strong' : 'moderate',
  };
}
