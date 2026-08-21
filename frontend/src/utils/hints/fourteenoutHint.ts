import type { FourteenOutResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FourteenOutPhase } from '../../types/phases';
import { fourteenOutCanRemove } from '../fourteenoutRemovablePairs';

/**
 * Returns a Fourteen Out hint:
 * - "remove-<col1>-<col2>" when two exposed tails sum to 14
 * - null when the game is over or genuinely stuck
 *
 * **"deal" は返さない。**クローン元の Monte Carlo は山札から補充できるので
 * 手が無いときに deal を勧めたが、Fourteen Out は最初に配り切る。組が無ければ
 * それが敗北で、勧める手は存在しない。
 *
 * Mirrors the backend's `forEachRemovablePair` scan (c2 > c1) so the same pair
 * is reported.
 */
export function getFourteenOutHint(state: FourteenOutResponse): HintResult | null {
  if (state.phase !== FourteenOutPhase.PLAYING) return null;

  const pair = findFirstPair(state);
  if (!pair) return null;

  return {
    targetAction: `remove-${pair.c1}-${pair.c2}`,
    reason: 'hint.removePair',
    confidence: 'strong',
  };
}

function findFirstPair(state: FourteenOutResponse): { c1: number; c2: number } | null {
  const cols = state.columns;
  for (let c1 = 0; c1 < cols.length; c1++) {
    for (let c2 = c1 + 1; c2 < cols.length; c2++) {
      if (fourteenOutCanRemove(cols, c1, c2)) return { c1, c2 };
    }
  }
  return null;
}
