import type { PolignacResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Polignac, or null when no
 * suggestion is available.
 *
 * Confidence tracks how forced the play is: refusing a trick that already has
 * a jack on it, or taking one to break a declared capot, are close to
 * automatic; shedding a jack on a safe trick is a judgement call.
 */
export function getPolignacHint(state: PolignacResponse): HintResult | null {
  const hint = state.hint;
  if (hint?.cardIndex === undefined) return null;

  // 自分の capot 中は全トリック取るしかなく、選択の余地がほぼ無い。
  const forced =
    hint.reason === 'polignacAvoidJack' || hint.reason === 'polignacBlockCapot' || hint.reason === 'polignacWinCapot';
  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: forced ? 'strong' : 'moderate',
  };
}
