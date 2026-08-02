import { describe, expect, it } from 'vitest';
import type { BourreResponse, Card } from '../../types/card';
import { getBourreHint } from './bourreHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 5), card('HEART', 9)], ...overrides }: Partial<BourreResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        folded: false,
        decided: true,
        drawn: true,
        bourreed: false,
        chips: 100,
        tricks: 0,
        cardCount: hand.length,
        cards: hand,
      },
      {
        id: 1,
        isHuman: false,
        isFinished: false,
        folded: false,
        decided: true,
        drawn: true,
        bourreed: false,
        chips: 100,
        tricks: 0,
        cardCount: 5,
        cards: [],
      },
    ],
    phase: 'play',
    currentPlayerIdx: 0,
    dealerIdx: 1,
    pot: 20,
    carryPot: 0,
    trumpSuit: 'SPADE',
    trumpCard: card('SPADE', 13),
    trickNumber: 1,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    leadPlayerIdx: 0,
    handNumber: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    validPlays: [0, 1],
    results: [],
    message: '',
    config: { cpuDifficulty: 1, ante: 5 },
    ...overrides,
  } as BourreResponse;
}

describe('getBourreHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getBourreHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getBourreHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet after folding', () => {
    const s = base();
    s.players[0].folded = true;
    expect(getBourreHint(s)).toBeNull();
  });

  // **フォローと勝ち札は強制。**選択肢が 1 枚に潰れている局面が起きる。
  it('names the only legal card', () => {
    expect(getBourreHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.bourreForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getBourreHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  it('points at the legal plays when there is a choice', () => {
    expect(getBourreHint(base({ validPlays: [0, 1] }))).toEqual({
      targetAction: 'card-0',
      reason: 'frontendHint.bourreChoose',
      confidence: 'moderate',
    });
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getBourreHint(base({ validPlays: [] }))).toBeNull();
  });

  // **降りるかどうかは切り札の枚数で決める。**3 枚以上なら勝負に残る。
  it('recommends staying in with a trump-heavy hand', () => {
    const hand = [card('SPADE', 13), card('SPADE', 12), card('SPADE', 9), card('HEART', 2)];
    expect(getBourreHint(base({ phase: 'decide', hand }))).toEqual({
      targetAction: 'stay',
      reason: 'frontendHint.bourreStay',
      confidence: 'moderate',
    });
  });

  it('recommends folding a hand with almost no trumps', () => {
    const hand = [card('HEART', 3), card('CLOVER', 4), card('DIAMOND', 5), card('HEART', 2)];
    expect(getBourreHint(base({ phase: 'decide', hand }))?.targetAction).toBe('fold');
  });

  it('stays quiet during the decision without a trump suit', () => {
    expect(getBourreHint(base({ phase: 'decide', trumpSuit: '' }))).toBeNull();
  });

  it('stays quiet in the draw phase', () => {
    expect(getBourreHint(base({ phase: 'draw' }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getBourreHint(base({ hand: [] }))).toBeNull();
  });
});
