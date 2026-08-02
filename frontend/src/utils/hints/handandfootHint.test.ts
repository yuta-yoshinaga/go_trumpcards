import { describe, expect, it } from 'vitest';
import type { Card, HandAndFootResponse } from '../../types/card';
import { HandAndFootPhase } from '../../types/phases';
import { getHandAndFootHint } from './handandfootHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; inFoot?: boolean; footCount?: number };

function base({
  hand = [card('SPADE', 4), card('HEART', 12)],
  inFoot = true,
  footCount = 0,
  ...overrides
}: Partial<HandAndFootResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        team: 0,
        isHuman: true,
        cardCount: hand.length,
        cards: hand,
        footCount,
        inFoot,
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    teams: [],
    phase: HandAndFootPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 9),
    drawPileCount: 40,
    discardPileCount: 5,
    isFrozen: false,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  } as unknown as HandAndFootResponse;
}

describe('getHandAndFootHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getHandAndFootHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getHandAndFootHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('takes the pile only with two naturals of the top rank', () => {
    // 1 枚合っているだけでは取れない。`naturalPairIndices` は常に 2 枚を要求する。
    const one = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 9), card('HEART', 12)],
      discardTop: card('CLOVER', 9),
    });
    expect(getHandAndFootHint(one)?.reason).toBe('frontendHint.handandfootNoPair');

    const pair = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 9), card('HEART', 9)],
      discardTop: card('CLOVER', 9),
    });
    expect(getHandAndFootHint(pair)?.targetAction).toBe('takeDiscard');
  });

  it('requires the pair on an open pile too, not only a frozen one', () => {
    // 凍結は取得条件を変えない。ワイルドが捨てられた記録でしかない。
    const s = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 9), card('HEART', 9)],
      isFrozen: true,
      discardTop: card('CLOVER', 9),
    });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootTakeFrozen');
  });

  it('does not count a wild as one of the two naturals', () => {
    // 2 はワイルド。ランクが 2 の山を 2-2 で取ろうとしても弾かれる。
    const s = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 2), card('HEART', 2)],
      discardTop: card('CLOVER', 2),
    });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootPileBlocked');
  });

  it('never offers a pile topped by a wild', () => {
    const joker = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 9), card('HEART', 9)],
      discardTop: card('JOKER', 0),
    });
    expect(getHandAndFootHint(joker)?.reason).toBe('frontendHint.handandfootPileBlocked');
  });

  it('never offers a pile topped by a black three', () => {
    const s = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 3), card('HEART', 3)],
      discardTop: card('CLOVER', 3),
    });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootPileBlocked');
  });

  it('offers a pile topped by a red three, which is not blocked', () => {
    // 赤 3 は黒 3 と違って取得を止めない。色で分けている枝の負のコントロール。
    const s = base({
      phase: HandAndFootPhase.DRAW,
      hand: [card('SPADE', 3), card('HEART', 3)],
      discardTop: card('DIAMOND', 3),
    });
    expect(getHandAndFootHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('draws from the stock when there is nothing to take', () => {
    const s = base({ phase: HandAndFootPhase.DRAW, discardTop: null });
    expect(getHandAndFootHint(s)?.targetAction).toBe('drawStock');
  });

  // **フットを取る前に上がらない。**手を空けるのが目的。
  it('aims at reaching the foot while one is still waiting', () => {
    expect(getHandAndFootHint(base({ inFoot: false, footCount: 11 }))).toEqual({
      targetAction: 'discard',
      reason: 'frontendHint.handandfootReachFoot',
      confidence: 'moderate',
    });
  });

  it('discards the heaviest card with no rank partner, and never a wild', () => {
    // 6♠ と 7♠ は**繋がっていない** —— このゲームに階段は無い。K が一番重く、
    // 2 はワイルドなので捨てない。
    const hand = [card('SPADE', 6), card('SPADE', 7), card('HEART', 13), card('DIAMOND', 2)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  it('keeps a wild rather than discarding it, even when it is the heaviest', () => {
    // ジョーカーだけが「相方なし」でも捨てない。捨てると山が凍る。
    const hand = [card('JOKER', 0), card('HEART', 5), card('DIAMOND', 5)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).not.toBe('card-0');
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 13), card('SPADE', 6), card('SPADE', 7)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  it('falls back to the heaviest card when every rank is paired', () => {
    const hand = [card('SPADE', 6), card('HEART', 6), card('SPADE', 8), card('HEART', 8)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  // 凍結中でも捨て札が無いことはある。
  it('draws from the stock when the pile is frozen and empty', () => {
    const s = base({ phase: HandAndFootPhase.DRAW, isFrozen: true, discardTop: null });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootDrawStock');
  });

  // 同じスートで隣のランクを持っていても取れない。このゲームに階段は無い。
  it('does not treat a same-suit neighbour as a reason to take the pile', () => {
    const hand = [card('CLOVER', 11), card('CLOVER', 13)];
    const s = base({ phase: HandAndFootPhase.DRAW, hand, discardTop: card('CLOVER', 12) });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootNoPair');
  });

  // MELD フェーズはこのヒントの対象外。
  it('stays quiet during the meld phase', () => {
    expect(getHandAndFootHint(base({ phase: HandAndFootPhase.MELD }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getHandAndFootHint(base({ hand: [] }))).toBeNull();
  });
});
