import type { MinibridgeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Minibridge, or null when no
 * suggestion is available.
 *
 * The contract hint names a contract rather than a card, so it carries no card
 * index. While playing, taking a trick the contract needs is close to
 * automatic; a card in the dummy's hand is flagged separately because the
 * index points into a different hand than the one below your seat.
 */
export function getMinibridgeHint(state: MinibridgeResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 契約の助言は札を指さない。どのレベル・種別を選ぶかが対象。
  if (hint.cardIndex === undefined) {
    return {
      targetAction: `contract-${hint.level}-${hint.suit}`,
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    // **ダミーの札は別の手札を指す。** 同じ index でも掴む場所が違う。
    targetAction: hint.reason === 'minibridgeDummy' ? `dummy-${hint.cardIndex}` : `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'minibridgeWinTrick' ? 'strong' : 'moderate',
  };
}
