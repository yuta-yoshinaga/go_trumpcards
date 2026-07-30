import type { LaughAndLieDownResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Laugh and Lie Down, or null when
 * there is no suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * LaughAndLieDownWebPresenter.buildBase.
 */
export function getLaughAndLieDownHint(state: LaughAndLieDownResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('laughandliedown.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn') return null;
  // must_lie_down is not a move the player chooses, but it IS worth saying:
  // the whole hand is about to go to the table.
  return { targetAction: 'play', reason: `hint.${what}`, confidence: 'moderate' };
}
