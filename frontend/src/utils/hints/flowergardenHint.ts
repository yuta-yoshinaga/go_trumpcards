import type { FlowerGardenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FlowerGardenPhase } from '../../types/phases';

/** Returns a Flower Garden frontend hint or null.
 *
 * Flower Garden is a high-difficulty solitaire where every card sits face-up
 * from the start (six open flower-bed fans plus a sixteen-card one-way bouquet
 * reserve), so the strongest tactical advice is "always advance to the
 * foundation if legal" because freeing a fan top or emptying a bouquet card
 * opens future moves. The backend supplies the next-move hint via `state.hint`;
 * we surface it unchanged. */
export function getFlowergardenHint(state: FlowerGardenResponse): HintResult | null {
  if (state.phase !== FlowerGardenPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  const target = state.hint.toZone === 'foundation' ? 'play.foundation' : 'play.tableau';
  const reason = state.hint.toZone === 'foundation' ? 'hintReason.toFoundation' : 'hintReason.toTableau';
  return { targetAction: target, reason, confidence: 'strong' };
}
