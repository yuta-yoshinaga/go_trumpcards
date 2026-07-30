import type { MushiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Mushi, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * MushiWebPresenter.buildBase.
 *
 * `reason` arrives as `mushi.hint.<what>`; only the prefix is rewritten so the
 * tooltip key stays in step with the server rather than being re-derived.
 */
export function getMushiHint(state: MushiResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('mushi.hint.', '');
  if (what === 'game_end' || what === 'round_end' || what === 'not_your_turn' || what === 'none') {
    return null;
  }
  return {
    targetAction: what === 'select' ? 'select' : 'play',
    reason: `hint.${what}`,
    confidence: 'moderate',
  };
}
