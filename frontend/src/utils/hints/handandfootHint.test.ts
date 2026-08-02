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

  it('takes a discard that connects while the pile is open', () => {
    const hand = [card('SPADE', 9), card('HEART', 12)];
    const s = base({ phase: HandAndFootPhase.DRAW, hand, discardTop: card('CLOVER', 9) });
    expect(getHandAndFootHint(s)?.targetAction).toBe('takeDiscard');
  });

  // **凍結中は自然のペアが要る。**繋がっているだけでは取れない。
  it('does not offer a frozen pile to a single matching card', () => {
    const hand = [card('SPADE', 9), card('HEART', 12)];
    const s = base({ phase: HandAndFootPhase.DRAW, hand, isFrozen: true, discardTop: card('CLOVER', 9) });
    expect(getHandAndFootHint(s)).toEqual({
      targetAction: 'drawStock',
      reason: 'frontendHint.handandfootFrozenBlocked',
      confidence: 'moderate',
    });
  });

  it('offers a frozen pile to a natural pair', () => {
    const hand = [card('SPADE', 9), card('HEART', 9)];
    const s = base({ phase: HandAndFootPhase.DRAW, hand, isFrozen: true, discardTop: card('CLOVER', 9) });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootTakeFrozen');
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

  it('discards the heaviest loose card once in the foot', () => {
    const hand = [card('SPADE', 6), card('SPADE', 7), card('HEART', 13), card('DIAMOND', 2)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 13), card('SPADE', 6), card('SPADE', 7)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  it('falls back to the heaviest card when everything connects', () => {
    const hand = [card('SPADE', 6), card('SPADE', 7), card('SPADE', 8)];
    expect(getHandAndFootHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  // 凍結中でも捨て札が無いことはある。
  it('draws from the stock when the pile is frozen and empty', () => {
    const s = base({ phase: HandAndFootPhase.DRAW, isFrozen: true, discardTop: null });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootDrawStock');
  });

  // 開いた場でも繋がらなければ山。
  it('draws from the stock when an open pile connects with nothing', () => {
    const hand = [card('SPADE', 2), card('HEART', 5)];
    const s = base({ phase: HandAndFootPhase.DRAW, hand, discardTop: card('CLOVER', 12) });
    expect(getHandAndFootHint(s)?.reason).toBe('frontendHint.handandfootDrawStock');
  });

  // MELD フェーズはこのヒントの対象外。
  it('stays quiet during the meld phase', () => {
    expect(getHandAndFootHint(base({ phase: HandAndFootPhase.MELD }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getHandAndFootHint(base({ hand: [] }))).toBeNull();
  });
});
