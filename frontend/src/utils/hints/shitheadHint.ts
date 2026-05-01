import type { ShitheadResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

const MAGIC_TWO = 2;
const MAGIC_TEN = 10;
/** "High card" boundary: at value 7+ the magic-7 lock-in starts to bite,
 * so spending a 10 (burn) here gains the most tempo. Matches Shithead's
 * 7-rule (next play must be ≤ 7) — anything ≥ 7 is considered high. */
const BURN_THRESHOLD = 7;
/** "Pressure" boundary above which a magic 2 (reset) is worth spending. */
const RESET_THRESHOLD = 9;

/** Source identifiers (sync with internal/domain/Shithead.go). */
const SOURCE_HAND = 'hand';
const SOURCE_FACE_DOWN = 'facedown';

/** Returns a Shithead frontend hint or null.
 *
 * Strategy:
 *   - On face-down phase, only blind play is possible — hint to play one.
 *   - When the discard pile is empty, lead with the lowest non-magic card.
 *   - When holding a magic 10 (burn) and pressure is high, suggest burn.
 *   - When 2 (reset) is in hand and the top is high, prefer the 2 as escape.
 *   - Otherwise, prefer the lowest legal card; pickup if no playable cards. */
export function getShitheadHint(state: ShitheadResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human) return null;
  if (state.currentTurn !== human.id) return null;

  if (state.currentSource === SOURCE_FACE_DOWN) {
    return { targetAction: 'play.facedown', reason: 'hint.blindPlay', confidence: 'moderate' };
  }

  const playable = state.currentSource === SOURCE_HAND ? human.handCards : human.faceUpCards;
  if (playable.length === 0) {
    return { targetAction: 'pickup', reason: 'hint.pickup', confidence: 'moderate' };
  }

  const top = state.discardPile.length > 0 ? state.discardPile[state.discardPile.length - 1] : null;
  const has2 = playable.some((c) => c.value === MAGIC_TWO);
  const has10 = playable.some((c) => c.value === MAGIC_TEN);

  if (top === null) {
    return { targetAction: 'play.lowest', reason: 'hint.leadLowest', confidence: 'strong' };
  }
  if (has10 && top.value >= BURN_THRESHOLD) {
    return { targetAction: 'play.ten', reason: 'hint.burnTen', confidence: 'strong' };
  }
  if (has2 && top.value >= RESET_THRESHOLD) {
    return { targetAction: 'play.two', reason: 'hint.resetTwo', confidence: 'strong' };
  }
  return { targetAction: 'play.lowest', reason: 'hint.playLowest', confidence: 'moderate' };
}
