import type { BuraResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Bura, or null when no suggestion
 * is available.
 *
 * The suggestion is computed by the Go backend and arrives on the response's
 * `hint` field on EVERY state response, not only on the `hint` command --
 * unlike most games here, whose presenters set it in `HintOutput` alone, which
 * no page calls. See BuraWebPresenter.buildBase.
 *
 * `reason` arrives as `bura.hint.<what>`; the tooltip wants a namespaced key,
 * so the prefix is rewritten rather than the whole string re-derived. A
 * declare or claim suggestion is strong -- both end the round on the spot --
 * while a card suggestion is moderate.
 */
export function getBuraHint(state: BuraResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  const what = hint.reason.replace('bura.hint.', '');
  if (what === 'game_end' || what === 'not_your_turn') return null;
  return {
    targetAction: what === 'declare' || what === 'claim' ? what : 'play',
    reason: `hint.${what}`,
    confidence: what === 'declare' || what === 'claim' ? 'strong' : 'moderate',
  };
}
