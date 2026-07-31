import type { PopeJoanResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Pope Joan, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * PopeJoanWebPresenter.buildBase.
 */
export function getPopeJoanHint(state: PopeJoanResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('popejoan.hint.', '');
  if (what === 'game_end' || what === 'deal_end' || what === 'not_your_turn' || what === 'none') {
    return null;
  }
  // Leading and following are both the same control here, but the reason
  // differs -- a stopped run takes your lowest card, a live one takes the
  // next higher of the suit.
  return { targetAction: 'play', reason: `hint.${what}`, confidence: 'moderate' };
}
