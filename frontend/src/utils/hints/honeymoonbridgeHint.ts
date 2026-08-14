import type { HoneymoonBridgeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Honeymoon Bridge, or null when no
 * suggestion is available.
 *
 * The auction hint names a contract rather than a card, so it carries no card
 * index. Taking a trick you need for the contract is close to automatic;
 * anything in the draw phase is a judgement call, because those tricks do not
 * score and only decide who draws first.
 */
export function getHoneymoonBridgeHint(state: HoneymoonBridgeResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 競りの助言は札を指さない。どの契約を宣言するかが対象。
  if (hint.cardIndex === undefined) {
    return {
      targetAction: hint.level === 0 ? 'pass' : `bid-${hint.level}-${hint.suit}`,
      reason: `hint.${hint.reason}`,
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'honeymoonbridgeWinTrick' ? 'strong' : 'moderate',
  };
}
