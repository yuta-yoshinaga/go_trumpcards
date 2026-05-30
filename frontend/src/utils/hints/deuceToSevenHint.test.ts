import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, DeuceToSevenResponse } from '../../types/card';
import { DeuceToSevenPhase } from '../../types/phases';
import { getDeuceToSevenHint } from './deuceToSevenHint';

function card(design: CardDesign, value: number): Card {
  return { design, value };
}

function makeState(overrides: Partial<DeuceToSevenResponse> = {}): DeuceToSevenResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cards: [],
        chips: 1000,
        currentBet: 0,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        drawCount: 0,
        totalDraws: 0,
        playStyleName: '',
      },
    ],
    pot: 0,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: DeuceToSevenPhase.DEAL,
    drawIndex: 0,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 0,
    ante: 0,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    cpuExchanges: [],
    message: '',
    ...overrides,
  };
}

const NUT_LOW = [card('SPADE', 7), card('HEART', 5), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)];
const TEN_HIGH = [card('SPADE', 10), card('HEART', 6), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)];
const PAIR_KINGS = [card('SPADE', 13), card('HEART', 13), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)];

describe('getDeuceToSevenHint', () => {
  it('returns null when the human folded', () => {
    const s = makeState();
    s.players[0].folded = true;
    expect(getDeuceToSevenHint(s)).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    const s = makeState({ currentTurn: 1 });
    expect(getDeuceToSevenHint(s)).toBeNull();
  });

  it('returns null outside bet/draw phases', () => {
    const s = makeState({ phase: DeuceToSevenPhase.INIT });
    expect(getDeuceToSevenHint(s)).toBeNull();
  });

  it('recommends stand on a made low in draw phase', () => {
    const s = makeState({ phase: DeuceToSevenPhase.DRAW });
    s.players[0].cards = NUT_LOW;
    const hint = getDeuceToSevenHint(s);
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.reason).toBe('hint.standPat');
  });

  it('recommends exchange in draw phase on a paired hand', () => {
    const s = makeState({ phase: DeuceToSevenPhase.DRAW });
    s.players[0].cards = PAIR_KINGS;
    const hint = getDeuceToSevenHint(s);
    expect(hint?.targetAction).toBe('exchange');
  });

  it('recommends big bet on a made low in betting phase', () => {
    const s = makeState({ phase: DeuceToSevenPhase.BET });
    s.players[0].cards = NUT_LOW;
    const hint = getDeuceToSevenHint(s);
    expect(hint?.targetAction).toBe('raise');
  });

  it('recommends call with a ten-high low', () => {
    const s = makeState({ phase: DeuceToSevenPhase.BET });
    s.players[0].cards = TEN_HIGH;
    const hint = getDeuceToSevenHint(s);
    expect(hint?.targetAction).toBe('call');
  });

  it('recommends fold on a weak hand facing a bet', () => {
    const s = makeState({ phase: DeuceToSevenPhase.BET, lastBet: 20 });
    s.players[0].cards = PAIR_KINGS;
    const hint = getDeuceToSevenHint(s);
    expect(hint?.targetAction).toBe('fold');
  });

  it('recommends check on a weak hand with no outstanding bet', () => {
    const s = makeState({ phase: DeuceToSevenPhase.BET, lastBet: 0 });
    s.players[0].cards = PAIR_KINGS;
    const hint = getDeuceToSevenHint(s);
    expect(hint?.targetAction).toBe('check');
    expect(hint?.reason).toBe('hint.checkWeak');
  });
});
