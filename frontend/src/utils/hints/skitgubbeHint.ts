import type { SkitgubbeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Skitgubbe, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * SkitgubbeWebPresenter.buildBase.
 */
export function getSkitgubbeHint(state: SkitgubbeResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('skitgubbe.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn' || what === 'none') return null;
  return {
    targetAction: what === 'pickup' ? 'pickup' : 'play',
    reason: `hint.${what}`,
    confidence: 'moderate',
  };
}
