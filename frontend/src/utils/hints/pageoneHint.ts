import type { Card, PageOneResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PageOnePhase } from '../../types/phases';

/** Face card values (J=11, Q=12, K=13). */
const FACE_VALUES = new Set([11, 12, 13]);
/** Value of 10 — highest scoring non-face card. */
const TEN = 10;

/**
 * Returns a Page One hint for the human player, or null when no suggestion applies.
 *
 * Strategy:
 * - MUST_DECLARE phase: always declare (skipping costs 2 penalty cards).
 * - PLAY phase on the human's turn:
 *   - If a playable card exists, recommend playing the highest-scoring one so
 *     we do not get stuck holding it at round end (face cards and 10 are worth 10/10 points).
 *   - Otherwise recommend drawing.
 */
export function getPageOneHint(state: PageOneResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === PageOnePhase.MUST_DECLARE) {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    if (state.currentPlayerIdx !== humanIdx) return null;
    return { targetAction: 'declare', reason: 'hint.declare', confidence: 'strong' };
  }

  if (state.phase !== PageOnePhase.PLAY) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.currentPlayerIdx !== humanIdx) return null;
  if (!state.discardTop) return null;

  const playable = human.cards.filter((c) => isPageOnePlayable(c, state.discardTop));
  if (playable.length === 0) {
    return { targetAction: 'draw', reason: 'hint.drawCard', confidence: 'strong' };
  }

  const hasHighValue = playable.some((c) => FACE_VALUES.has(c.value) || c.value === TEN);
  return {
    targetAction: 'play',
    reason: hasHighValue ? 'hint.playHighValue' : 'hint.playMatching',
    confidence: hasHighValue ? 'strong' : 'moderate',
  };
}

/**
 * True when `card` matches the discard top by suit or value.
 *
 * Exported so the page highlights exactly what the hint engine (and the CUI's
 * IsValidPlay) consider legal, rather than a second copy free to drift (#4744).
 * @param card - The candidate card.
 * @param top - The current discard top, if any.
 * @returns Whether the card may be played.
 */
export function isPageOnePlayable(card: Card, top: Card | null): boolean {
  // An empty discard accepts anything, matching PageOne.isValidPlay. Start()
  // always seeds the pile so this cannot happen in a live game, but the two
  // implementations are now compared directly and must not disagree.
  if (!top) return true;
  return card.design === top.design || card.value === top.value;
}
