import type { TeenDoPaanchResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for 3-2-5, or null when no suggestion
 * is available.
 *
 * The trump hint carries no card index — it names the suit to declare from the
 * five cards the 5-target seat can see. During play, going for a trick you
 * still owe is close to automatic; ducking once you have your number is the
 * judgement call, since **extra tricks score nothing**.
 */
export function getTeenDoPaanchHint(state: TeenDoPaanchResponse): HintResult | null {
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
    confidence: hint.reason === 'teendopaanchWinTrick' ? 'strong' : 'moderate',
  };
}
