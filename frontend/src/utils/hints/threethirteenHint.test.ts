import { describe, expect, it } from 'vitest';
import type { Card, ThreeThirteenResponse } from '../../types/card';
import { ThreeThirteenPhase } from '../../types/phases';
import { getThreeThirteenHint } from './threethirteenHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({
  hand = [card('SPADE', 4), card('HEART', 12)],
  ...overrides
}: Partial<ThreeThirteenResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand, deadwood: 0, roundScore: 0, cumulativeScore: 0 },
      { id: 1, isHuman: false, cardCount: 5, cards: [], deadwood: 0, roundScore: 0, cumulativeScore: 0 },
    ],
    phase: ThreeThirteenPhase.DISCARD,
    round: 1,
    wildRank: 3,
    dealCount: 5,
    currentPlayerIdx: 0,
    knockerIdx: -1,
    discardTop: card('CLOVER', 9),
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  } as ThreeThirteenResponse;
}

describe('getThreeThirteenHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getThreeThirteenHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getThreeThirteenHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getThreeThirteenHint(base({ phase: ThreeThirteenPhase.ROUND_END }))).toBeNull();
  });

  it('takes a discard that connects with the hand', () => {
    const hand = [card('SPADE', 9), card('HEART', 12)];
    const s = base({ phase: ThreeThirteenPhase.DRAW, hand, discardTop: card('CLOVER', 9) });
    expect(getThreeThirteenHint(s)).toEqual({
      targetAction: 'takeDiscard',
      reason: 'frontendHint.threethirteenTakeDiscard',
      confidence: 'moderate',
    });
  });

  // **ワイルドは何にでもなる。**繋がっていなくても拾う価値がある。
  it('takes a wild discard even when nothing in hand touches it', () => {
    const hand = [card('SPADE', 9), card('HEART', 12)];
    const s = base({ phase: ThreeThirteenPhase.DRAW, hand, wildRank: 5, discardTop: card('CLOVER', 5) });
    expect(getThreeThirteenHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('draws from the stock when the discard is no use', () => {
    const hand = [card('SPADE', 9), card('HEART', 12)];
    const s = base({ phase: ThreeThirteenPhase.DRAW, hand, wildRank: 5, discardTop: card('CLOVER', 2) });
    expect(getThreeThirteenHint(s)?.targetAction).toBe('drawStock');
  });

  it('discards the heaviest loose card', () => {
    const hand = [card('SPADE', 6), card('SPADE', 7), card('HEART', 13), card('DIAMOND', 2)];
    expect(getThreeThirteenHint(base({ hand }))).toEqual({
      targetAction: 'card-2',
      reason: 'frontendHint.threethirteenDiscardHeavy',
      confidence: 'moderate',
    });
  });

  // **ワイルドは捨てない。**一番重くても候補から外す。
  it('never suggests discarding a wild card', () => {
    const hand = [card('SPADE', 13), card('HEART', 4)];
    expect(getThreeThirteenHint(base({ hand, wildRank: 13 }))?.targetAction).toBe('card-1');
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 13), card('SPADE', 6), card('SPADE', 7)];
    expect(getThreeThirteenHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  it('falls back to the heaviest non-wild card when everything connects', () => {
    const hand = [card('SPADE', 6), card('SPADE', 7), card('SPADE', 8)];
    expect(getThreeThirteenHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  // 手札が全部ワイルドなら捨てる先がない。
  it('stays quiet when every card is wild', () => {
    const hand = [card('SPADE', 3), card('HEART', 3)];
    expect(getThreeThirteenHint(base({ hand, wildRank: 3 }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getThreeThirteenHint(base({ hand: [] }))).toBeNull();
  });

  it('treats the ace as adjacent to the king, not only to the two', () => {
    // ドメインは Ace-high のランを認める。生の値で引くと A(1) と K(13) が
    // 12 離れて見えるので、K を拾う理由を見落とす。
    const s = base({
      phase: ThreeThirteenPhase.DRAW,
      hand: [card('SPADE', 1), card('HEART', 5)],
      discardTop: card('SPADE', 13),
    });
    expect(getThreeThirteenHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('still requires the same suit for an ace-king neighbour', () => {
    const s = base({
      phase: ThreeThirteenPhase.DRAW,
      hand: [card('SPADE', 1), card('HEART', 5)],
      discardTop: card('CLOVER', 13),
    });
    expect(getThreeThirteenHint(s)?.targetAction).toBe('drawStock');
  });
});
