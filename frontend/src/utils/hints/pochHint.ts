import type { PochResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Poch, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * PochWebPresenter.buildBase.
 */
export function getPochHint(state: PochResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('poch.hint.', '');
  if (what === 'game_end' || what === 'deal_end' || what === 'not_your_turn' || what === 'none') {
    return null;
  }
  // Betting, folding and playing are three different controls.
  const targetAction = what === 'play' ? 'play' : what === 'bet' ? 'bet' : 'fold';
  return { targetAction, reason: `hint.${what}`, confidence: 'moderate' };
}
