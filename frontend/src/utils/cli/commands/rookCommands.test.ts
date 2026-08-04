import { describe, expect, it } from 'vitest';
import type { RookResponse } from '../../../types/card';
import { formatRookState, parseRookCommand, ROOK_HELP } from './rookCommands';

describe('parseRookCommand', () => {
  it('parses a numeric bid', () => {
    expect(parseRookCommand('b 75')).toEqual({ args: ['bid', { bid: 75 }] });
    expect(parseRookCommand('bid 120')).toEqual({ args: ['bid', { bid: 120 }] });
  });

  it('errors on a non-numeric bid', () => {
    expect(parseRookCommand('b')).toHaveProperty('error');
    expect(parseRookCommand('b x')).toHaveProperty('error');
  });

  it('parses pass', () => {
    expect(parseRookCommand('pa')).toEqual({ args: ['pass'] });
    expect(parseRookCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses exchange with five indices and a trump color', () => {
    expect(parseRookCommand('e 0 1 2 3 4 1')).toEqual({
      args: ['exchange', { discardIndices: [0, 1, 2, 3, 4], trumpColor: 1 }],
    });
    expect(parseRookCommand('exchange 5 6 7 8 9 3')).toEqual({
      args: ['exchange', { discardIndices: [5, 6, 7, 8, 9], trumpColor: 3 }],
    });
  });

  it('errors on a malformed exchange', () => {
    expect(parseRookCommand('e 0 1 2 3 4')).toHaveProperty('error');
    expect(parseRookCommand('e 0 1 2 3 x 1')).toHaveProperty('error');
  });

  it('parses play', () => {
    expect(parseRookCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parseRookCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseRookCommand('p')).toHaveProperty('error');
  });

  it('parses navigation and reset', () => {
    expect(parseRookCommand('n')).toEqual({ args: ['next'] });
    expect(parseRookCommand('next')).toEqual({ args: ['next'] });
    expect(parseRookCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseRookCommand('r')).toEqual({ args: ['reset'] });
    expect(parseRookCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for typos and errors on unknown', () => {
    expect(parseRookCommand('exchnge 0 1 2 3 4 1')).toHaveProperty('error');
    expect(parseRookCommand('zzz')).toHaveProperty('error');
  });
});

describe('formatRookState', () => {
  const base: RookResponse = {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [],
        team: 0,
        trickCount: 2,
        points: 40,
        bid: 75,
        passed: false,
        isDeclarer: true,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        team: 1,
        trickCount: 0,
        points: 0,
        bid: 0,
        passed: true,
        isDeclarer: false,
      },
    ] as RookResponse['players'],
    phase: 2,
    roundNumber: 1,
    trickNumber: 3,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    leadPlayerIdx: 0,
    trumpColor: 1,
    contractBid: 75,
    declarerIdx: 0,
    highestBid: 75,
    highestBidder: 0,
    nestCount: 0,
    nest: [],
    playableIndices: [],
    currentTrick: [
      { playerIdx: 0, card: { design: 'SPADE', value: 7, label: '7', deck: 'rook', color: 'red' } as never },
    ],
    teamScores: [120, 30],
    teamPoints: [40, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, targetScore: 500 },
    message: 'go',
  };

  it('formats players, scores, contract, trick and message', () => {
    const out = formatRookState(base);
    expect(out).toContain('Round 1');
    expect(out).toContain('Team0 120');
    expect(out).toContain('Contract: 75 (trump Red)');
    expect(out).toContain('You');
    expect(out).toContain('[declarer]');
    expect(out).toContain('[pass]');
    expect(out).toContain('Trick:');
    expect(out).toContain('go');
  });

  it('shows the highest bid when no contract yet', () => {
    const out = formatRookState({ ...base, contractBid: 0, highestBid: 70, currentTrick: [], message: '' });
    expect(out).toContain('Highest bid: 70');
    expect(out).not.toContain('Trick:');
  });

  it('exposes help text', () => {
    expect(ROOK_HELP.length).toBeGreaterThan(0);
  });
});
