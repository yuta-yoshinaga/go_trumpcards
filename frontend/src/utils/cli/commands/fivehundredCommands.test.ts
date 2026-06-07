import { describe, expect, it } from 'vitest';
import type { FiveHundredResponse } from '../../../types/card';
import { FIVE_HUNDRED_HELP, formatFiveHundredState, parseFiveHundredCommand } from './fivehundredCommands';

describe('parseFiveHundredCommand', () => {
  it('parses a suit bid', () => {
    expect(parseFiveHundredCommand('b 7 1')).toEqual({ args: ['bid', { bidKind: 1, bidTricks: 7, bidSuit: 1 }] });
    expect(parseFiveHundredCommand('bid 8 3')).toEqual({ args: ['bid', { bidKind: 1, bidTricks: 8, bidSuit: 3 }] });
  });

  it('errors on an incomplete suit bid', () => {
    expect(parseFiveHundredCommand('b 7')).toHaveProperty('error');
    expect(parseFiveHundredCommand('b x y')).toHaveProperty('error');
  });

  it('parses no-trump, misere and open misere', () => {
    expect(parseFiveHundredCommand('bnt 9')).toEqual({ args: ['bid', { bidKind: 2, bidTricks: 9 }] });
    expect(parseFiveHundredCommand('bnt')).toHaveProperty('error');
    expect(parseFiveHundredCommand('m')).toEqual({ args: ['bid', { bidKind: 3 }] });
    expect(parseFiveHundredCommand('misere')).toEqual({ args: ['bid', { bidKind: 3 }] });
    expect(parseFiveHundredCommand('om')).toEqual({ args: ['bid', { bidKind: 4 }] });
    expect(parseFiveHundredCommand('openmisere')).toEqual({ args: ['bid', { bidKind: 4 }] });
  });

  it('parses pass', () => {
    expect(parseFiveHundredCommand('pa')).toEqual({ args: ['pass'] });
    expect(parseFiveHundredCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses exchange with three indices', () => {
    expect(parseFiveHundredCommand('e 0 1 2')).toEqual({ args: ['exchange', { discardIndices: [0, 1, 2] }] });
    expect(parseFiveHundredCommand('e 0 1')).toHaveProperty('error');
    expect(parseFiveHundredCommand('exchange 0 1 x')).toHaveProperty('error');
  });

  it('parses play with and without a joker suit', () => {
    expect(parseFiveHundredCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3, jokerSuit: undefined }] });
    expect(parseFiveHundredCommand('play 3 2')).toEqual({ args: ['play', { cardIndex: 3, jokerSuit: 2 }] });
    expect(parseFiveHundredCommand('p')).toHaveProperty('error');
  });

  it('parses navigation and reset', () => {
    expect(parseFiveHundredCommand('n')).toEqual({ args: ['next'] });
    expect(parseFiveHundredCommand('next')).toEqual({ args: ['next'] });
    expect(parseFiveHundredCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseFiveHundredCommand('r')).toEqual({ args: ['reset'] });
    expect(parseFiveHundredCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for typos and errors on unknown', () => {
    expect(parseFiveHundredCommand('exchnge 0 1 2')).toHaveProperty('error');
    expect(parseFiveHundredCommand('zzz')).toHaveProperty('error');
  });
});

describe('formatFiveHundredState', () => {
  const base: FiveHundredResponse = {
    players: [
      { id: 0, isHuman: true, cardCount: 10, cards: [], team: 0, trickCount: 2, passed: false, isDeclarer: true },
      { id: 1, isHuman: false, cardCount: 10, cards: [], team: 1, trickCount: 0, passed: true, isDeclarer: false },
    ] as FiveHundredResponse['players'],
    phase: 2,
    roundNumber: 1,
    trickNumber: 3,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    leadPlayerIdx: 0,
    trumpSuit: 1,
    contractKind: 1,
    contractTricks: 7,
    contractValue: 140,
    declarerIdx: 0,
    highestBid: null,
    highestBidder: 0,
    jokerLeadSuit: -1,
    kittyCount: 0,
    currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 5 } as never }],
    teamScores: [120, 30],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, targetScore: 500 },
    message: 'go',
  };

  it('formats players, scores, trick and message', () => {
    const out = formatFiveHundredState(base);
    expect(out).toContain('Round 1');
    expect(out).toContain('Team0 120');
    expect(out).toContain('You');
    expect(out).toContain('[declarer]');
    expect(out).toContain('[pass]');
    expect(out).toContain('Trick:');
    expect(out).toContain('go');
  });

  it('omits the trick line when empty', () => {
    const out = formatFiveHundredState({ ...base, currentTrick: [], message: '' });
    expect(out).not.toContain('Trick:');
  });

  it('exposes help text', () => {
    expect(FIVE_HUNDRED_HELP.length).toBeGreaterThan(0);
  });
});
