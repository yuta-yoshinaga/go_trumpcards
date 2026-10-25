import type { SimpleSimonResponse } from '../types/card';
import { SimpleSimonPhase } from '../types/phases';

/** Reports whether the given card index starts a movable run. */
export function canSimpleSimonPickUp(state: SimpleSimonResponse | null | undefined, col: number, idx: number): boolean {
  return state?.phase === SimpleSimonPhase.PLAYING && (state.runStarts[col]?.includes(idx) ?? false);
}

/** Reports whether a selected run can be placed on the destination column. */
export function isSimpleSimonLegalDestination(
  state: SimpleSimonResponse | null | undefined,
  fromCol: number,
  idx: number,
  toCol: number,
): boolean {
  if (!state || fromCol === toCol) return false;
  const destination = state.columns[toCol];
  const movingHead = state.columns[fromCol]?.[idx];
  if (!destination || !movingHead) return false;
  if (destination.length === 0) return true;
  const destinationTop = destination[destination.length - 1];
  return destinationTop.value === movingHead.value + 1;
}

/** Reports whether the current game state allows undo. */
export function canSimpleSimonUndo(state: SimpleSimonResponse | null | undefined): boolean {
  return Boolean(state?.canUndo);
}
