import type { Card, ThreeThirteenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ThreeThirteenPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Three Thirteen, or null when no
 * suggestion is available.
 *
 * The response carries `deadwood` per player but not the melds behind it, so
 * this does not try to prove a meld. It uses the same shallow
 * "connects with something" test as the other rummies — same rank, or the
 * neighbouring rank in the same suit.
 *
 * The round's `wildRank` is the piece specific to this game: a wild card
 * substitutes for anything, so it is never the card to throw even when nothing
 * in hand happens to touch it.
 */
/** A(1) と K(13) の差。ランの端では隣り合う。 */
const ACE_TO_KING_GAP = 12;

export function getThreeThirteenHint(state: ThreeThirteenResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === ThreeThirteenPhase.DRAW) {
    const top = state.discardTop;
    return top && (top.value === state.wildRank || connects(top, human.cards))
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.threethirteenTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.threethirteenDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== ThreeThirteenPhase.DISCARD) return null;

  const idx = heaviestLoose(human.cards, state.wildRank);
  // **捨てられる札が無い。**全部がワイルドか繋がっている手では黙る。
  if (idx < 0) return null;
  return { targetAction: `card-${idx}`, reason: 'frontendHint.threethirteenDiscardHeavy', confidence: 'moderate' };
}

/**
 * 同じランクがあるか、同じスートで隣のランクがあるか。メルドの証明ではない。
 *
 * **A は 2 の隣であると同時に K の隣でもある** (internal/domain/ThreeThirteen.go:719)。生の値で
 * 引き算すると A(1) と K(13) が 12 離れて見え、A を持っている手で K を拾う
 * 理由を見落とす。
 */
function connects(c: Card, hand: Card[]): boolean {
  return hand.some((o) => o.value === c.value || (o.design === c.design && adjacent(o.value, c.value)));
}

/** 隣のランクか。A(1) は 2 の隣であると同時に K(13) の隣でもある。 */
function adjacent(a: number, b: number): boolean {
  const gap = Math.abs(a - b);
  return gap === 1 || gap === ACE_TO_KING_GAP;
}

/**
 * 捨てるべき札の位置。ワイルドは何にでもなるので候補から外す。
 *
 * 繋がっていない札があればその中で一番重いもの、無ければワイルド以外で
 * 一番重いもの。候補がまったく無ければ -1。
 */
function heaviestLoose(hand: Card[], wildRank: number): number {
  const candidates: number[] = [];
  for (let i = 0; i < hand.length; i += 1) {
    if (hand[i].value === wildRank) continue;
    candidates.push(i);
  }
  if (candidates.length === 0) return -1;

  const loose = candidates.filter(
    (i) =>
      !connects(
        hand[i],
        hand.filter((_, j) => j !== i),
      ),
  );
  const pool = loose.length > 0 ? loose : candidates;
  let best = pool[0];
  for (const i of pool) {
    if (hand[i].value > hand[best].value) best = i;
  }
  return best;
}
