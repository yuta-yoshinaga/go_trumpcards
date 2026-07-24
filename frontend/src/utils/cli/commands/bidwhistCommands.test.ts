import { describe, expect, it } from 'vitest';
import type { BidWhistResponse } from '../../../types/card';
import { BID_WHIST_HELP, formatBidWhistState, parseBidWhistCommand } from './bidwhistCommands';

describe('parseBidWhistCommand', () => {
  it('parses a bid with tricks and direction', () => {
    expect(parseBidWhistCommand('b 4 0')).toEqual({ args: ['bid', { bidTricks: 4, bidDirection: 0 }] });
    expect(parseBidWhistCommand('bid 7 2')).toEqual({ args: ['bid', { bidTricks: 7, bidDirection: 2 }] });
  });

  it('errors on an incomplete or invalid bid', () => {
    expect(parseBidWhistCommand('b 4')).toHaveProperty('error');
    expect(parseBidWhistCommand('b x y')).toHaveProperty('error');
  });

  it('parses pass', () => {
    expect(parseBidWhistCommand('pa')).toEqual({ args: ['pass'] });
    expect(parseBidWhistCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses a trump declaration', () => {
    expect(parseBidWhistCommand('t 1')).toEqual({ args: ['trump', { trumpSuit: 1 }] });
    expect(parseBidWhistCommand('trump 3')).toEqual({ args: ['trump', { trumpSuit: 3 }] });
    expect(parseBidWhistCommand('t')).toHaveProperty('error');
  });

  it('parses exchange with six indices', () => {
    expect(parseBidWhistCommand('e 0 1 2 3 4 5')).toEqual({
      args: ['exchange', { discardIndices: [0, 1, 2, 3, 4, 5] }],
    });
    expect(parseBidWhistCommand('e 0 1 2')).toHaveProperty('error');
    expect(parseBidWhistCommand('exchange 0 1 2 3 4 x')).toHaveProperty('error');
  });

  it('parses play', () => {
    expect(parseBidWhistCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parseBidWhistCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseBidWhistCommand('p')).toHaveProperty('error');
  });

  it('parses navigation and reset', () => {
    expect(parseBidWhistCommand('n')).toEqual({ args: ['next'] });
    expect(parseBidWhistCommand('next')).toEqual({ args: ['next'] });
    expect(parseBidWhistCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseBidWhistCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBidWhistCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for typos and errors on unknown', () => {
    expect(parseBidWhistCommand('exchnge 0 1 2 3 4 5')).toHaveProperty('error');
    expect(parseBidWhistCommand('zzz')).toHaveProperty('error');
  });
});

describe('formatBidWhistState', () => {
  const base: BidWhistResponse = {
    players: [
      { id: 0, isHuman: true, cardCount: 12, cards: [], team: 0, trickCount: 2, passed: false, isDeclarer: true },
      { id: 1, isHuman: false, cardCount: 12, cards: [], team: 1, trickCount: 0, passed: true, isDeclarer: false },
      { id: 2, isHuman: false, cardCount: 12, cards: [], team: 0, trickCount: 1, passed: false, isDeclarer: false },
      { id: 3, isHuman: false, cardCount: 12, cards: [], team: 1, trickCount: 0, passed: false, isDeclarer: false },
    ],
    phase: 3,
    roundNumber: 1,
    trickNumber: 4,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    leadPlayerIdx: 0,
    trumpSuit: 1,
    contractTricks: 3,
    contractDirection: 0,
    declarerIdx: 0,
    highestBid: { tricks: 3, direction: 0 },
    highestBidder: 0,
    kittyCount: 0,
    kittyIndices: [],
    currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 13 } as never }],
    teamScores: [2, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, targetScore: 7 },
    message: 'your turn',
  };

  it('renders scores, contract, players and the current trick', () => {
    const out = formatBidWhistState(base);
    expect(out).toContain('Round 1');
    expect(out).toContain('Team0 2');
    expect(out).toContain('Uptown');
    expect(out).toContain('[declarer]');
    expect(out).toContain('your turn');
  });

  it('exposes help text', () => {
    expect(BID_WHIST_HELP.length).toBeGreaterThan(0);
  });
});
