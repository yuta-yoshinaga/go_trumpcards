import type { Card, RistikontraResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Ristikontra, or null when no
 * suggestion applies.
 *
 * The priority order mirrors what actually wins the game:
 *
 * 1. **Counter.** If a capture is open (`counterRank` non-zero) and the hand
 *    holds that rank, playing it takes the whole bundle away from the capturer.
 *    This is the largest swing on the board and the mechanic the game is named
 *    for, so it is checked first.
 * 2. **Capture.** A card matching the pile top sweeps the pile.
 * 3. **Lay low.** Otherwise play the lowest card and keep the useful ranks.
 *
 * **This is not Pişti.** The clone source treats the Jack as a wild sweep and
 * pays a bonus for capturing a lone card; Ristikontra has neither, so a hint
 * that recommends a Jack "because it sweeps" would be advising a rule this game
 * does not have.
 */
export function getRistikontraHint(state: RistikontraResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== 'play') return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentTurn !== human.id) return null;

  // 1. 打ち返し — 直前の捕獲を束ごと奪える。盤面で一番大きい振れ幅。
  if (state.counterRank > 0) {
    const steal = human.cards.findIndex((c) => c.value === state.counterRank);
    if (steal >= 0) {
      return {
        targetAction: `card-${steal}`,
        reason: 'frontendHint.ristikontraCounter',
        confidence: 'strong',
      };
    }
  }

  // 2. 同ランクで場を総取り。
  const top = state.pileTop;
  if (top) {
    const match = human.cards.findIndex((c) => c.value === top.value);
    if (match >= 0) {
      return {
        targetAction: `card-${match}`,
        reason: 'frontendHint.ristikontraCapture',
        confidence: 'strong',
      };
    }
  }

  // 3. 取れないなら一番小さい札を捨てる。
  return {
    targetAction: `card-${lowestIdx(human.cards)}`,
    reason: 'frontendHint.ristikontraLayLow',
    confidence: 'moderate',
  };
}

/**
 * 手札で一番小さいランクの位置。
 *
 * **ジャックを避ける理由は無い。** クローン元のピシュティではジャックが万能の
 * 捕獲札なので温存していたが、リスティコントラではただの札。
 */
function lowestIdx(hand: Card[]): number {
  let best = 0;
  for (let i = 1; i < hand.length; i++) {
    if (hand[i].value < hand[best].value) best = i;
  }
  return best;
}
