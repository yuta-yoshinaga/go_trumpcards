import type { Card, CariocaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** フェーズ番号 (sync: internal/domain/Carioca.go)。 */
const PHASE_DRAW = 0;
const PHASE_PLAY = 1;

/** 契約スロットの種別 (sync: `CariocaContractSlot.kind`)。 */
const SLOT_SET = 0;

/** 手札に残ったときの失点 (sync: `cariocaCardPenalty`, internal/domain/Carioca.go:884)。 */
const JOKER_PENALTY = 25;
const ACE_PENALTY = 15;
const FACE_PENALTY = 10;
const FACE_FROM = 10;
const ACE = 1;

/**
 * Returns a frontend {@link HintResult} for Carioca, or null when no suggestion
 * is available.
 *
 * A player's `melds` is what they have laid down, not what their hand could
 * make, so the "is this card material" test here is a shallow one: a same-rank
 * partner, or a same-suit neighbour. Unlike Tonk — where the same shallow test
 * produced a hint that recommended an illegal knock — nothing here claims a
 * meld is legal. A pair genuinely is progress toward a trío (three of a rank)
 * and two suited neighbours genuinely are progress toward an escala (four in
 * suit, `CariocaRunSize`), so the shallow answer is the honest one.
 *
 * Two things are specific to this game.
 *
 * The **contract** has to go down whole, in one turn (`contractSlots` lists what
 * the round demands). Until it is met, the heavy cards are material rather than
 * liability, which inverts the usual rummy instinct to shed them.
 *
 * The **joker is the most expensive card to be caught with** — 25 points against
 * 15 for an ace and 10 for a court card (`cariocaCardPenalty`). It is also wild,
 * so it can fill any slot. Both facts point the same way: spend it, never throw
 * it. The discard suggestion therefore skips jokers even when a joker is the
 * costliest card in hand, and it ranks the rest by that same penalty table
 * rather than by rank, because that is what the round is actually scored on.
 */
export function getCariocaHint(state: CariocaResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === PHASE_DRAW) {
    const top = state.discardTop;
    return top && material(top, human.cards)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.cariocaTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.cariocaDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== PHASE_PLAY) return null;

  // **契約は一度に全部そろえて出す。**満たすまでは重い札も材料。
  if (!human.contractMet && state.contractSlots.length > 0) {
    const setsOnly = state.contractSlots.every((s) => s.kind === SLOT_SET);
    return {
      targetAction: 'meld',
      reason: setsOnly ? 'frontendHint.cariocaMeetSets' : 'frontendHint.cariocaMeetContract',
      confidence: 'moderate',
    };
  }

  return {
    targetAction: `card-${costliestSpare(human.cards)}`,
    reason: 'frontendHint.cariocaDiscardCostly',
    confidence: 'moderate',
  };
}

/** ジョーカーはワイルド (sync: `cariocaIsJoker`)。 */
function isJoker(c: Card): boolean {
  return c.design === 'JOKER';
}

/**
 * 同ランクの相方がいるか、同スートで隣のランクがいるか。ジョーカーは常に材料。
 *
 * 相方側でジョーカーを除いているのは飾りではない。ジョーカーは
 * `NewCard(CardDesignJoker, i, false)` で **`value` が 1 と 2** の札として作られる
 * (`TrumpCards.go:51`) ので、除かないと A や 2 と同ランクに見える。
 */
function material(c: Card, hand: Card[]): boolean {
  if (isJoker(c)) return true;
  return hand.some(
    (o) => !isJoker(o) && (o.value === c.value || (o.design === c.design && Math.abs(o.value - c.value) === 1)),
  );
}

/** 手札に残したときの失点 (sync: `cariocaCardPenalty`)。 */
function penalty(c: Card): number {
  if (isJoker(c)) return JOKER_PENALTY;
  if (c.value === ACE) return ACE_PENALTY;
  return c.value >= FACE_FROM ? FACE_PENALTY : c.value;
}

/**
 * 捨てるべき札の位置。どこにも繋がらない札のうち一番高くつくもの。
 *
 * ジョーカーは候補から外す。全部繋がっているなら、繋がっている札の中から
 * 一番高くつくものを出す。
 */
function costliestSpare(hand: Card[]): number {
  const keepable = hand.map((_, i) => i).filter((i) => !isJoker(hand[i]));
  const pool = keepable.length > 0 ? keepable : hand.map((_, i) => i);
  const loose = pool.filter(
    (i) =>
      !material(
        hand[i],
        hand.filter((_, j) => j !== i),
      ),
  );
  const candidates = loose.length > 0 ? loose : pool;
  let best = candidates[0];
  for (const i of candidates) {
    if (penalty(hand[i]) > penalty(hand[best])) best = i;
  }
  return best;
}
