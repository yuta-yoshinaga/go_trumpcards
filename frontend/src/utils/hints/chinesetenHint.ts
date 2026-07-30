import type { ChineseTenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Chinese Ten, or null when there is
 * no suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * ChineseTenWebPresenter.buildBase.
 */
export function getChineseTenHint(state: ChineseTenResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('chineseten.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn' || what === 'none') return null;
  return {
    targetAction: what === 'select' ? 'select' : 'play',
    reason: `hint.${what}`,
    confidence: 'moderate',
  };
}
