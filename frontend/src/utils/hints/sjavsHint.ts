import type { SjavsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Sjavs, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * SjavsWebPresenter.buildBase.
 */
export function getSjavsHint(state: SjavsResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('sjavs.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn' || what === 'none') return null;
  return {
    // Bidding and passing both target the bid controls, not a card.
    targetAction: what === 'bid' || what === 'pass' ? 'bid' : 'play',
    reason: `hint.${what}`,
    confidence: 'moderate',
  };
}
