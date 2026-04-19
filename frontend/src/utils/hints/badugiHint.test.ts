import { describe, expect, it } from 'vitest';
import type { BadugiResponse, Card, CardDesign } from '../../types/card';
import { BadugiPhase } from '../../types/phases';
import { getBadugiHint } from './badugiHint';

function card(design: CardDesign, value: number): Card {
  return { design, value };
}

function makeState(overrides: Partial<BadugiResponse> = {}): BadugiResponse {
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
        handSize: 0,
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
    phase: BadugiPhase.DEAL,
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

describe('getBadugiHint', () => {
  it('returns null when the human folded', () => {
    const s = makeState();
    s.players[0].folded = true;
    expect(getBadugiHint(s)).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    const s = makeState({ currentTurn: 1 });
    expect(getBadugiHint(s)).toBeNull();
  });

  it('returns null outside bet/draw phases', () => {
    const s = makeState({ phase: BadugiPhase.INIT });
    expect(getBadugiHint(s)).toBeNull();
  });

  it('recommends stand on a completed Badugi in draw phase', () => {
    const s = makeState({ phase: BadugiPhase.DRAW });
    s.players[0].cards = [card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3), card('CLOVER', 4)];
    const hint = getBadugiHint(s);
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.reason).toBe('hint.standPat');
  });

  it('recommends exchange in draw phase when subset is under 4', () => {
    const s = makeState({ phase: BadugiPhase.DRAW });
    s.players[0].cards = [card('SPADE', 1), card('SPADE', 5), card('DIAMOND', 3), card('CLOVER', 4)];
    const hint = getBadugiHint(s);
    expect(hint?.targetAction).toBe('exchange');
  });

  it('recommends big bet on a Badugi in betting phase', () => {
    const s = makeState({ phase: BadugiPhase.BET });
    s.players[0].cards = [card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3), card('CLOVER', 4)];
    const hint = getBadugiHint(s);
    expect(hint?.targetAction).toBe('raise');
  });

  it('recommends call with a 3-card low', () => {
    const s = makeState({ phase: BadugiPhase.BET });
    s.players[0].cards = [card('SPADE', 1), card('SPADE', 5), card('DIAMOND', 3), card('CLOVER', 4)];
    const hint = getBadugiHint(s);
    expect(hint?.targetAction).toBe('call');
  });

  it('recommends fold on a weak hand facing a bet', () => {
    const s = makeState({ phase: BadugiPhase.BET, lastBet: 20 });
    s.players[0].cards = [card('SPADE', 1), card('SPADE', 2), card('SPADE', 3), card('SPADE', 4)];
    const hint = getBadugiHint(s);
    expect(hint?.targetAction).toBe('fold');
  });

  it('recommends check on a weak hand with no outstanding bet', () => {
    const s = makeState({ phase: BadugiPhase.BET, lastBet: 0 });
    s.players[0].cards = [card('SPADE', 1), card('SPADE', 2), card('SPADE', 3), card('SPADE', 4)];
    const hint = getBadugiHint(s);
    expect(hint?.targetAction).toBe('check');
  });
});
