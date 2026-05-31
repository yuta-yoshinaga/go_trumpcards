import type { BarbuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Trumps contract id (positive — try to win tricks). */
const CONTRACT_TRUMPS = 5;
/** Dominoes contract id (lay cards / pass). */
const CONTRACT_DOMINOES = 6;

/**
 * Barbu hint heuristic. Only advises during the play phase for the human's
 * turn; contract selection is left to the player.
 *
 * - Trumps: win tricks → recommend playing.
 * - Dominoes: place a card if any is playable, otherwise pass.
 * - Negative contracts (No Tricks / Hearts / Queens / Barbu / Last Trick):
 *   avoid taking the trick → recommend playing a low/safe card.
 *
 * Returns null when there is no state, the game has ended, it is not the human's
 * turn, or the game is not in the play phase.
 */
export function getBarbuHint(state: BarbuResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;
  if (state.phase !== 'play') return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.id !== state.currentTurn) return null;
  if (human.cards.length === 0) return null;

  if (state.currentContract === CONTRACT_DOMINOES) {
    if (state.dominoPlayable.length === 0) {
      return { targetAction: 'pass', reason: 'barbu.hint.pass', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'barbu.hint.placeDomino', confidence: 'moderate' };
  }

  if (state.currentContract === CONTRACT_TRUMPS) {
    return { targetAction: 'play', reason: 'barbu.hint.win', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'barbu.hint.avoid', confidence: 'moderate' };
}
