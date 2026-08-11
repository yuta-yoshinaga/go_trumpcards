import type { BalootResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Baloot, or null when no
 * suggestion is available.
 *
 * The declaration hint carries no card index — it advises a mode, and when
 * that mode is Hokom the suit matters as much as the mode, so the target
 * action names it. Feeding a winning partner is close to automatic; taking a
 * trick yourself is a judgement call.
 */
export function getBalootHint(state: BalootResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 宣言フェーズは札を指さない。どのモードを選ぶかが助言の対象。
  if (hint.cardIndex === undefined) {
    if (hint.reason === 'balootDeclareHokom') {
      return {
        // **スートまで含めて指す。** Hokom はどのスートを切り札にするかで
        // 序列そのものが変わるので、モードだけでは行動を決められない。
        targetAction: `declare-hokom-${hint.suit}`,
        reason: `hint.${hint.reason}`,
        confidence: 'moderate',
      };
    }
    return {
      targetAction: hint.reason === 'balootDeclareSun' ? 'declare-sun' : 'pass-declare',
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'balootFeedPartner' ? 'strong' : 'moderate',
  };
}
