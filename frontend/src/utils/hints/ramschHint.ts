import type { RamschResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { RamschPhase } from '../../types/phases';

/**
 * Returns a Ramsch frontend hint, or null when there is nothing to advise.
 *
 * The advice itself comes from the domain (`state.hint`), which already knows
 * whether it is the human's turn. **The reason keys describe avoiding points,
 * not winning tricks** — this is a game about refusing to take things, and
 * generic trick-taking advice reads backwards here.
 *
 * There is only one phase to advise on: no auction, no contract, no discard.
 */
export function getRamschHint(state: RamschResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (!state.hint) return null;
  if (state.phase !== RamschPhase.PLAY) return null;
  if (state.hint.cardIndex == null) return null;

  return {
    targetAction: 'play',
    // **`hint.<reason>` の形にする。** `ns:key` と書くと i18next の
    // nsSeparator (既定 ':') が効いて名前空間解決になり、キーが見つからず
    // 生の識別子 (`avoid_points`) がそのままツールチップに出る。
    reason: `hint.${state.hint.reason || 'avoid_points'}`,
    confidence: 'moderate',
    targetPos: state.hint.cardIndex,
  };
}
