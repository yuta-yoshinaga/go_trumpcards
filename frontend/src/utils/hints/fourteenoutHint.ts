import type { FourteenOutResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FourteenOutPhase } from '../../types/phases';

/**
 * Returns a Fourteen Out hint:
 * - "remove-<col1>-<col2>" when two exposed tails sum to 14
 * - null when the game is over or genuinely stuck
 *
 * **"deal" は返さない。**クローン元の Monte Carlo は山札から補充できるので
 * 手が無いときに deal を勧めたが、Fourteen Out は最初に配り切る。組が無ければ
 * それが敗北で、勧める手は存在しない。
 *
 * Reads the server legal-target map and preserves the backend pair scan order.
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
  for (let c1 = 0; c1 < state.columns.length; c1++) {
    const targets = state.legalTargets[`tableau-${c1}`];
    if (targets?.tableau.length) return { c1, c2: targets.tableau[0] };
  }
  return null;
}
