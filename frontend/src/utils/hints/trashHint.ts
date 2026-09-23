import type { TrashResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TrashPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Trash, or null when no suggestion is
 * available.
 *
 * Drawing resolves itself (`Draw` chains placements automatically via
 * `resolveChain`), so the only real choice is **where to put a wild**. The
 * domain supplies the authoritative recommended slot so the Web and CUI agree.
 */
export function getTrashHint(state: TrashResponse): HintResult | null {
  if (state.phase === TrashPhase.GAME_OVER) return null;

  const human = state.players.findIndex((p) => !p.isCpu);
  if (human < 0 || state.current !== human) return null;

  if (state.phase === TrashPhase.AWAIT_WILD) {
    const slot = state.suggestedWildSlot;
    // **スロット 0 も正当。**真偽値で見ると先頭だけ落ちる。
    if (slot < 0) return null;
    return { targetAction: `slot-${slot}`, reason: 'frontendHint.trashPlaceWildScarce', confidence: 'moderate' };
  }

  if (state.phase !== TrashPhase.PLAYER_TURN) return null;
  return { targetAction: 'draw', reason: 'frontendHint.trashDraw', confidence: 'moderate' };
}
