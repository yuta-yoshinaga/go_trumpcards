import type { HasenpfefferResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Hasenpfeffer, or null when no
 * suggestion is available.
 *
 * The bidding hint carries no card index — it names a number, or a pass. The
 * discard hint names a card but also a suit, since the declarer chooses both
 * at once. Taking a trick you still need is close to automatic; holding back
 * while your partner is winning is the judgement call.
 */
export function getHasenpfefferHint(state: HasenpfefferResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 競りの助言は札を指さない。何トリック宣言するかが助言の対象。
  if (hint.cardIndex === undefined) {
    return {
      targetAction: hint.value > 0 ? `bid-${hint.value}` : 'pass',
      reason: `hint.${hint.reason}`,
      // **親が降りられない場面は選択肢が無い。** 迷う余地がないので強く出す。
      confidence: hint.reason === 'hasenpfefferMustBid' ? 'strong' : 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'hasenpfefferWinTrick' ? 'strong' : 'moderate',
  };
}
