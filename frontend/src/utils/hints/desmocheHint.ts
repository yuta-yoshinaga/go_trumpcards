import type { DesmocheResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Desmoche, or null when there is no
 * suggestion.
 *
 * The suggestion is computed by the Go backend and arrives on EVERY state
 * response, not only on the `hint` command — unlike most games here, whose
 * presenters set it in `HintOutput` alone, which no page calls. See
 * DesmocheWebPresenter.buildBase.
 */
export function getDesmocheHint(state: DesmocheResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('desmoche.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn' || what === 'none') return null;
  // Drawing, melding and discarding are three different controls.
  const targetAction = what === 'draw' ? 'draw' : what === 'meld' ? 'meld' : 'discard';
  return { targetAction, reason: `hint.${what}`, confidence: 'moderate' };
}
