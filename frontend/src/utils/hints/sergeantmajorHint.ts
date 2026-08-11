import type { SergeantMajorResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Sergeant Major, or null when no
 * suggestion is available.
 *
 * The trump hint names a suit and the discard hint names several cards, so
 * neither carries a single card index. During play, taking a trick you still
 * owe is close to automatic; pressing on past your target still scores, but is
 * a judgement call.
 */
export function getSergeantMajorHint(state: SergeantMajorResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 捨て札は複数枚を指すので、先頭を代表として扱う。
  if (hint.indices.length > 0) {
    return {
      targetAction: `discard-${hint.indices.join('-')}`,
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

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
    confidence: hint.reason === 'sergeantmajorWinTrick' ? 'strong' : 'moderate',
  };
}
