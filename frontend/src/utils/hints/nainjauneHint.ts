import type { NainJauneResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Le Nain Jaune, or null when there
 * is no suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * NainJauneWebPresenter.buildBase.
 */
export function getNainJauneHint(state: NainJauneResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('nainjaune.hint.', '');
  if (what === 'game_end' || what === 'deal_end' || what === 'not_your_turn' || what === 'none') {
    return null;
  }
  // All three live reasons point at the same control; the reason is what
  // differs — claiming a box is worth calling out separately from merely
  // continuing the run.
  return { targetAction: 'play', reason: `hint.${what}`, confidence: 'moderate' };
}
