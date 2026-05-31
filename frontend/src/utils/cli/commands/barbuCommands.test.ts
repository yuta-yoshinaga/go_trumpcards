import { describe, expect, it } from 'vitest';
import type { BarbuResponse, Card } from '../../../types/card';
import { BARBU_HELP, formatBarbuState, parseBarbuCommand } from './barbuCommands';

describe('parseBarbuCommand', () => {
  it('parses reset aliases', () => {
    expect(parseBarbuCommand('r')).toEqual({ args: ['r'] });
    expect(parseBarbuCommand('reset')).toEqual({ args: ['r'] });
  });

  it('parses next aliases', () => {
    expect(parseBarbuCommand('n')).toEqual({ args: ['n'] });
    expect(parseBarbuCommand('next')).toEqual({ args: ['n'] });
  });

  it('parses log aliases', () => {
    expect(parseBarbuCommand('l')).toEqual({ args: ['log'] });
    expect(parseBarbuCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses a contract without trump', () => {
    expect(parseBarbuCommand('c 0')).toEqual({ args: ['c', { contract: 0, trumpSuit: -1 }] });
  });

  it('parses a contract with trump', () => {
    expect(parseBarbuCommand('contract 5 1')).toEqual({ args: ['c', { contract: 5, trumpSuit: 1 }] });
  });

  it('errors on contract without an argument', () => {
    expect(parseBarbuCommand('c')).toHaveProperty('error');
  });

  it('errors on a non-numeric contract', () => {
    expect(parseBarbuCommand('c x')).toHaveProperty('error');
  });

  it('errors on a non-numeric trump', () => {
    expect(parseBarbuCommand('c 5 x')).toHaveProperty('error');
  });

  it('parses a play command', () => {
    expect(parseBarbuCommand('p 3')).toEqual({ args: ['p', { handIndex: 3 }] });
  });

  it('parses a pass (-1) command', () => {
    expect(parseBarbuCommand('p -1')).toEqual({ args: ['p', { handIndex: -1 }] });
  });

  it('errors on play without an argument', () => {
    expect(parseBarbuCommand('p')).toHaveProperty('error');
  });

  it('errors on a non-numeric hand index', () => {
    expect(parseBarbuCommand('play foo')).toHaveProperty('error');
  });

  it('errors on an unknown command', () => {
    expect(parseBarbuCommand('zzz')).toHaveProperty('error');
  });
});

describe('formatBarbuState', () => {
  const base: BarbuResponse = {
    message: 'hi',
    players: [
      { id: 0, isHuman: true, cardCount: 5, cards: [], trickCount: 1, dominoRank: 0, totalScore: 3 },
      { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, dominoRank: 0, totalScore: -2 },
    ],
    phase: 'play',
    dealNumber: 0,
    totalDeals: 28,
    dealerIdx: 0,
    currentTurn: 0,
    currentContract: 5,
    trumpSuit: 1,
    trickNumber: 2,
    currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 9 } as Card }],
    lastTrick: [],
    lastTrickWinner: -1,
    tablePlaced: [0, 0, 0, 0, 0],
    dominoPlayable: [],
    usedContracts: [false, false, false, false, false, true, false],
    gameEndFlag: false,
    config: { cpuDifficulty: 1 },
    roundWinners: [],
    lastDealDetail: null,
  } as unknown as BarbuResponse;

  it('renders deal, contract, turn, players, and trick lines', () => {
    const out = formatBarbuState(base);
    expect(out).toContain('Deal 1/28');
    expect(out).toContain('Contract: Trumps');
    expect(out).toContain('You');
    expect(out).toContain('Trick:');
    expect(out).toContain('hi');
  });

  it('shows End when the game is over and no contract line when unselected', () => {
    const out = formatBarbuState({ ...base, gameEndFlag: true, currentContract: -1, currentTrick: [] });
    expect(out).toContain('End');
    expect(out).not.toContain('Contract:');
  });
});

describe('BARBU_HELP', () => {
  it('lists the core commands', () => {
    expect(BARBU_HELP.length).toBeGreaterThan(0);
    expect(BARBU_HELP.join('\n')).toContain('c <0-6>');
  });
});
