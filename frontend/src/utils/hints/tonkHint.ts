import type { Card, TonkResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TonkPhase } from '../../types/phases';

/**
 * ノックできるデッドウッドの上限 (sync: internal/domain/Tonk.go:17,
 * `TonkKnockThreshold`). サーバはこの値を送ってこないので持ち直している。
 */
const KNOCK_THRESHOLD = 5;

/** 絵札は 10 点。A は 1 点。 */
const FACE_VALUE = 10;

/**
 * Returns a frontend {@link HintResult} for Tonk, or null when no suggestion is
 * available.
 *
 * There is no server-side GetHint here, so this works from the hand. It uses the
 * same shallow connects-with-something test as the other rummies rather than
 * proving a meld — `knockerMelds` describes the settlement, not what the hand
 * could make.
 *
 * The specific rule is the knock: it is refused above five points of deadwood,
 * so the hint only offers it when a shallow count already clears the bar. That
 * count can only over-estimate — a card it calls loose might belong to a meld it
 * cannot see — so it never suggests a knock the server would reject.
 */
export function getTonkHint(state: TonkResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === TonkPhase.DRAW) {
    const top = state.discardTop;
    return top && connects(top, human.cards)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.tonkTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.tonkDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== TonkPhase.DISCARD) return null;

  // **ノックは 5 点以下でしか通らない。**浅い数え方は多めに出るので、
  // これを超えて勧めることはない。
  if (looseValue(human.cards) <= KNOCK_THRESHOLD) {
    return { targetAction: 'knock', reason: 'frontendHint.tonkKnock', confidence: 'moderate' };
  }

  const idx = heaviestLoose(human.cards);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.tonkDiscardHeavy', confidence: 'moderate' };
}

/** 同じランクがあるか、同じスートで隣のランクがあるか。メルドの証明ではない。 */
function connects(c: Card, hand: Card[]): boolean {
  return hand.some((o) => o.value === c.value || (o.design === c.design && Math.abs(o.value - c.value) === 1));
}

/** 札の点数。絵札は 10、それ以外は数字どおり。 */
function points(c: Card): number {
  return c.value >= FACE_VALUE ? FACE_VALUE : c.value;
}

/** 繋がっていない札の合計点。実際のデッドウッド以上の値になる。 */
function looseValue(hand: Card[]): number {
  return hand.reduce((sum, _c, i) => {
    const rest = hand.filter((_, j) => j !== i);
    return connects(hand[i], rest) ? sum : sum + points(hand[i]);
  }, 0);
}

/**
 * 繋がっていない札のうち一番重いものの位置。
 *
 * 端札が一枚も無い場合は考えなくてよい。そのとき `looseValue` は 0 で、
 * 上の分岐が必ずノックを返してここまで来ないため、「全部繋がっていたら」の
 * 逃げ道を書くと本番で決して通らない枝になる。
 */
function heaviestLoose(hand: Card[]): number {
  const loose = hand
    .map((_, i) => i)
    .filter(
      (i) =>
        !connects(
          hand[i],
          hand.filter((_, j) => j !== i),
        ),
    );
  let best = loose[0];
  for (const i of loose) {
    if (points(hand[i]) > points(hand[best])) best = i;
  }
  return best;
}
