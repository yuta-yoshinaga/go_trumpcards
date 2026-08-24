import type { CoincheResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Coinche, or null when no
 * suggestion is available.
 *
 * **The advice comes from the backend, not from a second heuristic here.**
 * The clone carried a frontend copy of Belote's pick-up thresholds, which
 * only stays right for as long as nobody touches the Go side — and Coinche's
 * bidding is nothing like Belote's turn-up anyway. `Coinche.GetHint()`
 * already decides whether it is the human's turn and what to advise, so this
 * adapter only re-maps its `reason` into the shared tooltip shape.
 */
export function getCoincheHint(state: CoincheResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    // The in-page tooltip annotates the play step; the bidding advice is
    // shown through the dedicated server-hint banner.
    targetAction: 'play',
    // **理由キーの置き場は 1 つ。** ページのサーバヒント帯も同じ
    // `hintReason.*` を引くので、`hint.*` を別に持つと同じ文言が 2 か所に
    // 増えて必ず片方だけ直される。
    reason: `hintReason.${hint.reason}`,
    confidence: 'moderate',
  };
}
