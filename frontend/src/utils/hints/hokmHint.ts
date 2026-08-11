import type { HokmResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Hokm, or null when no suggestion
 * is available.
 *
 * The trump hint carries no card index — it names the suit to declare from the
 * hakem's first five cards. Holding high cards back while your partner is
 * already winning the trick is close to automatic; going for a trick yourself
 * is a judgement call.
 */
export function getHokmHint(state: HokmResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 切り札の宣言は札を指さない。どのスートにするかが助言の対象。
  if (hint.cardIndex === undefined) {
    return {
      targetAction: `trump-${hint.suit}`,
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'hokmSaveCards' ? 'strong' : 'moderate',
  };
}
