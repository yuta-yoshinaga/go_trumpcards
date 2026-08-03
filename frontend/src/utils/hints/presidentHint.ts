import type { Card, PresidentResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 通常時の強さ (sync: `PresidentCardStrength`, internal/domain/PresidentEval.go:5)。 */
const ACE_STRENGTH = 14;
const TWO_STRENGTH = 15;

/** 革命時は 18 から引いて反転する (sync: `PresidentCardStrengthRevolution`)。 */
const REVOLUTION_PIVOT = 18;

/**
 * Returns a frontend {@link HintResult} for President, or null when no
 * suggestion is available.
 *
 * There is no server-side GetHint, so this works from the hand.
 *
 * The whole game turns on one flag. **A revolution inverts the order** —
 * `PresidentCardStrengthRevolution` is literally `18 - strength`, so the two
 * that was strongest becomes weakest and the three that was weakest becomes
 * strongest (`PresidentEval.go:16`). A hint that named a card without checking
 * `revolutionActive` would be pointing at the opposite end of the hand half the
 * time, which is worse than saying nothing.
 *
 * Beyond that the advice is the ordinary shedding one: lead the weakest card you
 * hold, because the strong ones are what let you take back the lead later.
 */
export function getPresidentHint(state: PresidentResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isFinished || human.cards.length === 0) return null;
  if (state.currentTurn !== human.id) return null;

  const weakest = weakestIndex(human.cards, state.revolutionActive);

  // 場が空なら好きに出せる。一番弱い札から捌く。
  if (state.tableCards.length === 0) {
    return {
      targetAction: `card-${weakest}`,
      reason: state.revolutionActive ? 'frontendHint.presidentLeadRevolution' : 'frontendHint.presidentLead',
      confidence: 'moderate',
    };
  }

  // 場より強い札のうち一番弱いものを出す。無ければパス。
  const beat = weakestBeating(human.cards, state.tableCards, state.revolutionActive);
  if (beat < 0) {
    return { targetAction: 'pass', reason: 'frontendHint.presidentPass', confidence: 'moderate' };
  }
  return {
    targetAction: `card-${beat}`,
    reason: state.revolutionActive ? 'frontendHint.presidentFollowRevolution' : 'frontendHint.presidentFollow',
    confidence: 'moderate',
  };
}

/** 現在の向きでの強さ。A は 14、2 は 15、革命中は 18 から引く。 */
function strength(value: number, revolution: boolean): number {
  const base = value === 1 ? ACE_STRENGTH : value === 2 ? TWO_STRENGTH : value;
  return revolution ? REVOLUTION_PIVOT - base : base;
}

/** 一番弱い札の位置。 */
function weakestIndex(hand: Card[], revolution: boolean): number {
  let best = 0;
  for (let i = 1; i < hand.length; i += 1) {
    if (strength(hand[i].value, revolution) < strength(hand[best].value, revolution)) best = i;
  }
  return best;
}

/**
 * 場札より強い札のうち一番弱いものの位置。無ければ -1。
 *
 * 場札の枚数（ペアやスリーカード）は見ていない。**枚数を合わせる必要がある**が、
 * どの組を出すかはプレイヤーが選ぶので、ここでは「その強さを超える札がそもそも
 * あるか」だけを答える。無ければパスしかない、という判断は枚数に依らず正しい。
 */
function weakestBeating(hand: Card[], table: Card[], revolution: boolean): number {
  const target = strength(table[0].value, revolution);
  let best = -1;
  for (let i = 0; i < hand.length; i += 1) {
    const s = strength(hand[i].value, revolution);
    if (s <= target) continue;
    if (best < 0 || s < strength(hand[best].value, revolution)) best = i;
  }
  return best;
}
