import type { Card, PishtiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** ジャックはいつでも場を総取りする。 */
const JACK = 11;

/**
 * Returns a frontend {@link HintResult} for Pişti, or null when no suggestion
 * is available.
 *
 * The scoring move is easy to walk past: matching the pile's top card takes the
 * pile, and doing it while **exactly one** card is showing is a Pişti and worth
 * a bonus on top. A jack sweeps the pile whenever it is played, so it is the
 * card to hold back rather than the card to lead.
 */
export function getPishtiHint(state: PishtiResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== 'play') return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentTurn !== human.id) return null;

  const top = state.pileTop;

  // **1 枚の場に同ランクを重ねると Pişti。**ボーナスが付く唯一の形。
  if (top && state.pileCount === 1) {
    const idx = human.cards.findIndex((c) => c.value === top.value);
    if (idx >= 0) {
      return { targetAction: `card-${idx}`, reason: 'frontendHint.pishtiPisti', confidence: 'strong' };
    }
  }

  // 同ランクなら場を取れる。
  if (top) {
    const idx = human.cards.findIndex((c) => c.value === top.value);
    if (idx >= 0) {
      return { targetAction: `card-${idx}`, reason: 'frontendHint.pishtiCapture', confidence: 'strong' };
    }
  }

  // **ジャックは場があるときに使う。**空の場に出しても取るものがない。
  const jack = human.cards.findIndex((c) => c.value === JACK);
  if (jack >= 0 && state.pileCount > 0) {
    return { targetAction: `card-${jack}`, reason: 'frontendHint.pishtiJackSweep', confidence: 'moderate' };
  }

  return {
    targetAction: `card-${lowestIdx(human.cards)}`,
    reason: 'frontendHint.pishtiLayLow',
    confidence: 'moderate',
  };
}

/** 手札で一番小さいランクの位置。ジャックは避ける。 */
function lowestIdx(hand: Card[]): number {
  const pool = hand.map((_, i) => i).filter((i) => hand[i].value !== JACK);
  const idxs = pool.length > 0 ? pool : hand.map((_, i) => i);
  let best = idxs[0];
  for (const i of idxs) {
    if (hand[i].value < hand[best].value) best = i;
  }
  return best;
}
