import type { GuandanResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (1 = Play)。 */
const PLAY_PHASE = 1;

/**
 * Returns a frontend {@link HintResult} for Guandan, or null when no suggestion
 * is available.
 *
 * Guandan sends no list of legal plays — a combo is chosen, not picked from a
 * list — so this does not try to name a card. What it can say without guessing
 * at combination rules is the two things the type comments single out:
 *
 *  - the hand's **level rank beats aces** and its two hearts are wild, so those
 *    cards are worth far more than their face value suggests;
 *  - a partner who has already gone out means the pair is playing for **how
 *    far** they climb, not whether they win the hand.
 */
export function getGuandanHint(state: GuandanResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  // **相方が既に上がっている。**勝敗ではなく順位で得られるレベルが変わる。
  const partner = state.players.find((p) => p.team === human.team && p.id !== human.id);
  if (partner && partner.finishedRank > 0) {
    return { targetAction: 'play', reason: 'frontendHint.guandanPartnerOut', confidence: 'moderate' };
  }

  // **レベル札はエースより強く、そのうち赤 2 枚はワイルド。**
  if (human.cards.some((c) => c.value === state.level)) {
    return { targetAction: 'play', reason: 'frontendHint.guandanLevelCards', confidence: 'moderate' };
  }

  // 場に何も出ていないならリードできる。
  return state.lastCombo === null
    ? { targetAction: 'play', reason: 'frontendHint.guandanFreeLead', confidence: 'moderate' }
    : { targetAction: 'pass', reason: 'frontendHint.guandanConsiderPass', confidence: 'moderate' };
}
