import type { SevenTwentySevenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for SevenTwentySeven, or null when there
 * is nothing to advise.
 *
 * The advice is computed by the Go backend and surfaced on `state.hint`. The
 * **reason names which side it is chasing** — "draw" on its own is not advice
 * when 7 and 27 pull in opposite directions, so the reason key carries the
 * whole point of the suggestion.
 */
export function getSevenTwentySevenHint(state: SevenTwentySevenResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.draw ? 'card' : 'stand',
    // **`hint.<reason>` の形にする。** `ns:key` と書くと i18next の
    // nsSeparator (既定 ':') が効いて名前空間解決になり、キーが見つからず
    // 生の識別子 (`chase_seven`) がそのままツールチップに出る。
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
