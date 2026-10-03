import type { PaiGowHint, PaiGowResponse } from '../types/games/paigow';
import { PaiGowPhase } from '../types/phases';

/** Return the server's house-way low-hand indices during the set-hands phase. */
export function paiGowHouseWaySplit(state: PaiGowResponse | null | undefined): readonly [number, number] | null {
  const hint: PaiGowHint | null | undefined = state?.hint;
  if (state?.phase !== PaiGowPhase.SET_HANDS || hint == null) return null;
  return [hint.lowIdx0, hint.lowIdx1];
}
