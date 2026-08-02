import type { Card, TrashResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TrashPhase } from '../../types/phases';

/** スロットは 10 個 (sync: `TrashSlotCnt`, internal/domain/Trash.go:24)。 */
const SLOT_COUNT = 10;

/**
 * Returns a frontend {@link HintResult} for Trash, or null when no suggestion is
 * available.
 *
 * There is no server-side GetHint here, and there is not much to decide: drawing
 * resolves itself (`Draw` chains placements automatically via `resolveChain`),
 * so the only real choice is **where to put a wild**.
 *
 * A king or a joker (`isTrashWild`, `Trash.go:404`) can stand in for any
 * position, and the position it fills is the one thing the player picks. The
 * useful answer is the **lowest still-empty slot**: filling from the bottom
 * keeps the longest run of consecutive positions covered, so a later draw of a
 * middling rank is more likely to land somewhere useful. Spending the wild high
 * leaves a gap that only one specific rank can close.
 */
export function getTrashHint(state: TrashResponse): HintResult | null {
  if (state.phase === TrashPhase.GAME_OVER) return null;

  const human = state.players.findIndex((p) => !p.isCpu);
  if (human < 0 || state.current !== human) return null;

  if (state.phase === TrashPhase.AWAIT_WILD) {
    const slot = lowestEmptySlot(state.players[human].slots);
    // **スロット 0 も正当。**真偽値で見ると先頭だけ落ちる。
    if (slot < 0) return null;
    return { targetAction: `slot-${slot}`, reason: 'frontendHint.trashPlaceWildLow', confidence: 'moderate' };
  }

  if (state.phase !== TrashPhase.PLAYER_TURN) return null;
  return { targetAction: 'draw', reason: 'frontendHint.trashDraw', confidence: 'moderate' };
}

/** まだ埋まっていない一番低い位置。全部埋まっていれば -1。 */
function lowestEmptySlot(slots: { card?: Card; faceUp: boolean }[]): number {
  for (let i = 0; i < Math.min(slots.length, SLOT_COUNT); i += 1) {
    if (!slots[i].faceUp) return i;
  }
  return -1;
}
