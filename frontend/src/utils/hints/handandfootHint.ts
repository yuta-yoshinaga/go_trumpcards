import type { Card, HandAndFootResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HandAndFootPhase } from '../../types/phases';

/** 山を取るのに要るナチュラル札の枚数 (sync: `naturalPairIndices`)。 */
const NATURAL_PAIR = 2;

/**
 * Returns a frontend {@link HintResult} for Hand and Foot, or null when no
 * suggestion is available.
 *
 * The response carries no melds for the human, so the discard advice is a
 * shallow one — but **shallow in the same shape the game actually melds in**.
 * `validateNewMeld` (`internal/domain/HandAndFoot.go:1070`) requires every
 * natural card in a meld to share one rank; there is no run anywhere in this
 * game. An earlier version of this file reused the same-suit-adjacency test the
 * run-building rummies use (Gin Rummy, Kalooki, Chinchón), where it is a fair
 * proxy. Here it is simply wrong: it calls 6♠-7♠ connected when they can never
 * be melded together, and the advice built on it was wrong in both directions.
 *
 * Taking the discard pile is more restricted than "it connects":
 * `PlayerDrawFromDiscard` (line 350) refuses outright when the top card is a
 * **black three** or a **wild** (`CanastaIsWild` — a Joker or any two), and it
 * requires exactly two **natural** cards of the top card's rank from hand in
 * every case. It is not a freeze-only requirement: `naturalPairIndices` is
 * validated on every call, so the pair is the rule and `isFrozen` does not
 * change what is legal. The flag only records that a wild was discarded onto the
 * pile; taking it clears the flag (line 388).
 *
 * The other thing specific here is the **foot**: a player still on their first
 * hand has one waiting, so emptying the hand is the goal and going out is not.
 */
export function getHandAndFootHint(state: HandAndFootResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === HandAndFootPhase.DRAW) {
    const top = state.discardTop;
    if (!top) {
      return { targetAction: 'drawStock', reason: 'frontendHint.handandfootDrawStock', confidence: 'moderate' };
    }
    // **黒 3 とワイルドがトップなら山は取れない。**凍結とは無関係に拒否される。
    if (isBlackThree(top) || isWild(top)) {
      return { targetAction: 'drawStock', reason: 'frontendHint.handandfootPileBlocked', confidence: 'strong' };
    }
    // 取るには手札から同ランクのナチュラル 2 枚が要る。これも常に必要。
    const naturals = state.players
      .filter((p) => p.isHuman)
      .flatMap((p) => p.cards)
      .filter((c) => !isWild(c) && c.value === top.value).length;
    if (naturals < NATURAL_PAIR) {
      return { targetAction: 'drawStock', reason: 'frontendHint.handandfootNoPair', confidence: 'strong' };
    }
    return state.isFrozen
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.handandfootTakeFrozen', confidence: 'moderate' }
      : { targetAction: 'takeDiscard', reason: 'frontendHint.handandfootTakeDiscard', confidence: 'moderate' };
  }

  if (state.phase !== HandAndFootPhase.DISCARD) return null;

  // **まだフットを持っていない。**上がるのではなく手を空けるのが目的。
  if (!human.inFoot && human.footCount > 0) {
    return { targetAction: 'discard', reason: 'frontendHint.handandfootReachFoot', confidence: 'moderate' };
  }

  const idx = heaviestLoose(human.cards);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.handandfootDiscardHeavy', confidence: 'moderate' };
}

/** 同じランクの札が他にあるか。メルドの証明ではないが、**組める形ではある**。 */
function sameRankElsewhere(c: Card, hand: Card[]): boolean {
  return hand.some((o) => o.value === c.value);
}

/** ジョーカーと 2 はワイルド (sync: `CanastaIsWild`, internal/domain/Canasta.go:1314)。 */
function isWild(c: Card): boolean {
  return c.design === 'JOKER' || c.value === 2;
}

/** 黒 3 (sync: `CanastaIsBlack3`, internal/domain/Canasta.go:1324)。 */
function isBlackThree(c: Card): boolean {
  return c.value === 3 && (c.design === 'SPADE' || c.design === 'CLOVER');
}

/**
 * 同ランクの相方がいない札のうち一番重いもの。
 *
 * ワイルドは捨てない。捨てると山が凍り、次に取るのが自分でも他人でも難しくなる
 * (`HandAndFoot.go:599`)。全部が相方持ちなら一番重い札を返す。
 */
function heaviestLoose(hand: Card[]): number {
  const usable = hand.map((_, i) => i).filter((i) => !isWild(hand[i]));
  const pool = usable.length > 0 ? usable : hand.map((_, i) => i);
  const loose = pool.filter(
    (i) =>
      !sameRankElsewhere(
        hand[i],
        hand.filter((_, j) => j !== i),
      ),
  );
  const candidates = loose.length > 0 ? loose : pool;
  let best = candidates[0];
  for (const i of candidates) {
    if (hand[i].value > hand[best].value) best = i;
  }
  return best;
}
