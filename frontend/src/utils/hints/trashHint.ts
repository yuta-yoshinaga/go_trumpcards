import type { TrashResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — Trash has no decisional hint beyond wild placement, which is
 * driven directly by the UI affordance (clicking a face-down slot). Keeping an
 * explicit stub so the game is registered in {@link hooks/useGameHint.useGameHint | useGameHint} and the hint
 * toggle UI remains consistent across the suite.
 */
export function getTrashHint(_state: TrashResponse): HintResult | null {
  return null;
}
