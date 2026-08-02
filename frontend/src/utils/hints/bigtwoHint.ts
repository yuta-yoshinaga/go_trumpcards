import type { BigTwoResponse, Card } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 値の強さ: 3 が最弱、2 が最強 (sync: `tienLenValueStrength`, internal/domain/BigTwoEval.go:12)。 */
const TWO_STRENGTH = 12;
const ACE_STRENGTH = 11;

/**
 * スート順 **♦ < ♣ < ♥ < ♠** (sync: `bigTwoSuitStrength`, internal/domain/BigTwoEval.go:23)。
 *
 * **Tiến Lên とは逆向き。**あちらは ♠ が最弱で ♥ が最強 (`TienLenEval.go:23`)。
 * 同じ「値 * 4 + スート」の形なので、写して直し忘れると自信を持って逆を勧める。
 */
const SUIT_STRENGTH: Record<string, number> = { DIAMOND: 0, CLOVER: 1, HEART: 2, SPADE: 3 };

/** 合成の強さは `値 * 4 + スート` (sync: BigTwoEval.go:9)。 */
const SUIT_SPAN = 4;

/**
 * Returns a frontend {@link HintResult} for Big Two, or null when no suggestion
 * is available.
 *
 * There is no server-side GetHint, so the ordering is ported here.
 *
 * As in Tiến Lên the **suit is part of the rank** — strength is `value * 4 + suit`
 * — but the suit order is the **opposite way round**: ♦ < ♣ < ♥ < ♠ here
 * (`BigTwoEval.go:9`) against ♠ < ♣ < ♦ < ♥ there. So the weakest card in this
 * deck is the three of diamonds and the strongest the two of spades, exactly
 * inverting the two games' extremes. Copying one file to the other and leaving
 * the suit table alone would produce a hint that is confidently backwards.
 *
 * The advice itself is the ordinary shedding one: lead the weakest card, and
 * when following play the cheapest thing that still beats the table.
 */
export function getBigTwoHint(state: BigTwoResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentTurn !== human.id) return null;

  if (state.tableCards.length === 0) {
    return {
      targetAction: `card-${weakestIndex(human.cards)}`,
      reason: 'frontendHint.bigtwoLead',
      confidence: 'moderate',
    };
  }

  const beat = weakestBeating(human.cards, state.tableCards);
  if (beat < 0) {
    return { targetAction: 'pass', reason: 'frontendHint.bigtwoPass', confidence: 'moderate' };
  }
  return { targetAction: `card-${beat}`, reason: 'frontendHint.bigtwoFollow', confidence: 'moderate' };
}

/** 値とスートを合わせた強さ。 */
function strength(c: Card): number {
  const v = c.value === 2 ? TWO_STRENGTH : c.value === 1 ? ACE_STRENGTH : c.value - 3;
  return v * SUIT_SPAN + (SUIT_STRENGTH[c.design] ?? 0);
}

/** 一番弱い札の位置。 */
function weakestIndex(hand: Card[]): number {
  let best = 0;
  for (let i = 1; i < hand.length; i += 1) {
    if (strength(hand[i]) < strength(hand[best])) best = i;
  }
  return best;
}

/**
 * 場札より強い札のうち一番弱いものの位置。無ければ -1。
 *
 * 場札の枚数（ペアや連番）は見ていない。どの組を出すかはプレイヤーが選ぶので、
 * ここでは「その強さを超える札がそもそもあるか」だけを答える。無ければパスしか
 * ない、という判断は枚数に依らず正しい。
 */
function weakestBeating(hand: Card[], table: Card[]): number {
  // 場の代表は一番強い札。組の強さはそれで決まる。
  const target = Math.max(...table.map(strength));
  let best = -1;
  for (let i = 0; i < hand.length; i += 1) {
    if (strength(hand[i]) <= target) continue;
    if (best < 0 || strength(hand[i]) < strength(hand[best])) best = i;
  }
  return best;
}
