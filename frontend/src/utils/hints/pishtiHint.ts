import type { Card, PishtiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** ジャックはいつでも場を総取りする。 */
const JACK = 11;

/**
 * Returns a frontend {@link HintResult} for Pişti, or null when no suggestion
 * is available.
 *
 * The scoring move is easy to walk past: capturing while the pile holds
 * **exactly one** card is a Pişti and pays a bonus on top of the pile.
 *
 * The part that is easy to get wrong — and that the first version of this file
 * did get wrong — is what counts as that capture. It is not only a rank match.
 * The domain header (`internal/domain/Pishti.go:20-25`) spells out three cases:
 *
 * - +10 (`PishtiBonusSingle`) for a rank match onto a lone card;
 * - +10 for a **jack** onto a lone non-jack;
 * - +20 (`PishtiBonusJackOnJack`) for a jack onto a **lone jack** — the biggest
 *   single score in the game.
 *
 * Treating the jack as merely "a card that sweeps" hides the best move on the
 * board, so a lone jack showing is checked before anything else.
 *
 * A jack (or a rank match) onto a multi-card pile is an ordinary capture with no
 * bonus, which is why it ranks below both.
 */
export function getPishtiHint(state: PishtiResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== 'play') return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentTurn !== human.id) return null;

  const top = state.pileTop;
  const jack = human.cards.findIndex((c) => c.value === JACK);
  const lonePile = top !== null && state.pileCount === 1;

  // **一番大きい得点はジャックを 1 枚のジャックに重ねること** (+20)。
  if (lonePile && jack >= 0 && top.value === JACK) {
    return { targetAction: `card-${jack}`, reason: 'frontendHint.pishtiJackOnJack', confidence: 'strong' };
  }

  // 1 枚の場を取れば Pişti (+10)。同ランクでもジャックでも成立する。
  if (lonePile) {
    const match = human.cards.findIndex((c) => c.value === top.value);
    if (match >= 0) {
      return { targetAction: `card-${match}`, reason: 'frontendHint.pishtiPisti', confidence: 'strong' };
    }
    if (jack >= 0) {
      return { targetAction: `card-${jack}`, reason: 'frontendHint.pishtiJackPisti', confidence: 'strong' };
    }
  }

  // 複数枚の場は同ランクで取れる。ボーナスは付かない。
  if (top) {
    const match = human.cards.findIndex((c) => c.value === top.value);
    if (match >= 0) {
      return { targetAction: `card-${match}`, reason: 'frontendHint.pishtiCapture', confidence: 'strong' };
    }
  }

  // **ジャックは場があるときに使う。**空の場に出しても取るものがない。
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
