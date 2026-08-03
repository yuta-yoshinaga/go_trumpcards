import type { ZwickerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Zwicker, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * ZwickerWebPresenter.buildBase.
 */
export function getZwickerHint(state: ZwickerResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('zwicker.hint.', '');
  if (what === 'game_end' || what === 'round_end' || what === 'not_your_turn' || what === 'none') {
    return null;
  }
  // Capturing and trailing are two different controls.
  const targetAction = what === 'take' ? 'take' : 'discard';
  return { targetAction, reason: `hint.${what}`, confidence: 'moderate' };
}
