import type { TrexResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Trex, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * TrexWebPresenter.buildBase.
 */
export function getTrexHint(state: TrexResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('trex.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn' || what === 'none') return null;
  // Choosing a contract targets the contract buttons, not a card.
  const targetAction = what === 'choose' ? 'choose' : what === 'pass' ? 'pass' : 'play';
  return { targetAction, reason: `hint.${what}`, confidence: 'moderate' };
}
