import type { RollingStoneResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Rolling Stone, or null when no
 * suggestion is available.
 *
 * Being unable to follow leaves exactly one move, so that advice is certain;
 * which card to lead or follow with is a judgement call.
 */
export function getRollingStoneHint(state: RollingStoneResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  // 引き取るしかない場面は札を指さない。選択の余地がない。
  if (hint.cardIndex === undefined) {
    return { targetAction: 'pickup', reason: `hint.${hint.reason}`, confidence: 'strong' };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
