import type { ToepenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Toepen, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * ToepenWebPresenter.buildBase.
 *
 * Folding is reported as strong: declining a raise costs the stake BEFORE it,
 * so it is the cheap way out of a hand you cannot win.
 */
export function getToepenHint(state: ToepenResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('toepen.hint.', '');
  if (what === 'game_end' || what === 'hand_end' || what === 'not_your_turn' || what === 'none') {
    return null;
  }
  if (what === 'fold' || what === 'stay') {
    return { targetAction: what, reason: `hint.${what}`, confidence: what === 'fold' ? 'strong' : 'moderate' };
  }
  return { targetAction: 'play', reason: `hint.${what}`, confidence: 'moderate' };
}
