import type { TarabishResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Tarabish, or null when no
 * suggestion is available.
 *
 * The bidding hint carries no card index — it advises whether to take the
 * turned suit — so its `targetAction` names the decision rather than a card.
 * Feeding a winning partner is close to automatic; taking a trick yourself is
 * a judgement call.
 */
export function getTarabishHint(state: TarabishResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 入札フェーズは札を指さない。引き受けるかどうかが助言の対象。
  if (hint.cardIndex === undefined) {
    return {
      targetAction: hint.reason === 'tarabishTakeTrump' ? 'take-trump' : 'pass-trump',
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'tarabishFeedPartner' ? 'strong' : 'moderate',
  };
}
