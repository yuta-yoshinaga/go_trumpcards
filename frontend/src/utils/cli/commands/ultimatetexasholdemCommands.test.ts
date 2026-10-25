import { describe, expect, it } from 'vitest';
import { UltimateTexasHoldemPhase } from '../../../types/phases';
import { utHoldemBetBounds } from '../../utHoldemBet';
import { parseUltimateTexasHoldemCommand as parse } from './ultimatetexasholdemCommands';

describe('parseUltimateTexasHoldemCommand', () => {
  it('maps GUI bet, play, check, fold and reset actions to exec arguments', () => {
    expect(parse('bet 100 20')).toEqual({ args: ['bet', 100, 20] });
    expect(parse('b 50')).toEqual({ args: ['bet', 50, 0] });
    for (const mult of [1, 2, 3, 4])
      expect(parse(`play ${mult}`)).toEqual({ args: ['play', undefined, undefined, mult] });
    for (const mult of [1, 2, 3, 4])
      expect(parse(String(mult))).toEqual({ args: ['play', undefined, undefined, mult] });
    expect(parse('c')).toEqual({ args: ['check'] });
    expect(parse('check')).toEqual({ args: ['check'] });
    expect(parse('f')).toEqual({ args: ['fold'] });
    expect(parse('fold')).toEqual({ args: ['fold'] });
    expect(parse('r')).toEqual({ args: ['reset'] });
    expect(parse('reset')).toEqual({ args: ['reset'] });
  });
  it('rejects missing, malformed and out of range arguments', () => {
    for (const cmd of ['bet', 'bet 0', 'bet 4 -1', 'bet x', 'play', 'play 0', 'play 5', 'play x'])
      expect(parse(cmd)).toHaveProperty('error');
  });
  it('enforces the page bet range including min, step, cap, and inclusive upper bound', () => {
    const boundsForAnte = (ante: number, trips: number) => utHoldemBetBounds(ante, trips, 200);
    const constraints = {
      chips: 200,
      boundsForAnte: (ante: number, trips: number) => utHoldemBetBounds(ante, trips, 200),
      phase: UltimateTexasHoldemPhase.BET,
    };
    expect(parse('bet 0', constraints)).toHaveProperty('error');
    expect(parse('bet 15', constraints)).toHaveProperty('error');
    expect(parse('bet 100 0', constraints)).toEqual({ args: ['bet', 100, 0] });
    expect(parse('bet 110', constraints)).toHaveProperty('error');
    expect(boundsForAnte(90, 20).valid).toBe(true);
    expect(parse('bet 90 20', constraints)).toEqual({ args: ['bet', 90, 20] });
    expect(parse('bet 90 30', constraints)).toHaveProperty('error');
    const outOfStep = parse('bet 15', constraints);
    expect('error' in outOfStep && outOfStep.error).toContain('range');
  });
  it('rejects phase and chip invalid actions', () => {
    expect(parse('4', { chips: 500, anteBet: 100, phase: UltimateTexasHoldemPhase.FLOP })).toHaveProperty('error');
    expect(parse('bet 10', { chips: 200, phase: UltimateTexasHoldemPhase.PRE_FLOP })).toHaveProperty('error');
    expect(parse('play 3', { chips: 500, anteBet: 100, phase: UltimateTexasHoldemPhase.FLOP })).toHaveProperty('error');
    expect(parse('play 2', { chips: 500, anteBet: 100, phase: UltimateTexasHoldemPhase.PRE_FLOP })).toHaveProperty(
      'error',
    );
    expect(parse('play 1', { chips: 500, anteBet: 100, phase: UltimateTexasHoldemPhase.FLOP })).toHaveProperty('error');
    expect(parse('play 3', { chips: 200, anteBet: 100, phase: UltimateTexasHoldemPhase.PRE_FLOP })).toHaveProperty(
      'error',
    );
    expect(parse('play 2', { chips: 100, anteBet: 100, phase: UltimateTexasHoldemPhase.FLOP })).toHaveProperty('error');
    expect(parse('play 1', { chips: 50, anteBet: 100, phase: UltimateTexasHoldemPhase.RIVER })).toHaveProperty('error');
    expect(parse('check', { chips: 0, phase: UltimateTexasHoldemPhase.RIVER })).toHaveProperty('error');
    expect(parse('fold', { chips: 0, phase: UltimateTexasHoldemPhase.FLOP })).toHaveProperty('error');
    expect(parse('reset', { chips: 0, phase: UltimateTexasHoldemPhase.PRE_FLOP })).toHaveProperty('error');
    expect(parse('check', { chips: 0, phase: UltimateTexasHoldemPhase.PRE_FLOP })).toEqual({ args: ['check'] });
    expect(parse('fold', { chips: 0, phase: UltimateTexasHoldemPhase.RIVER })).toEqual({ args: ['fold'] });
    expect(parse('reset', { chips: 0, phase: UltimateTexasHoldemPhase.END })).toEqual({ args: ['reset'] });
  });
  it('rejects unsupported API commands and unknown commands', () => {
    expect(parse('hint')).toHaveProperty('error');
    expect(parse('log')).toHaveProperty('error');
    expect(parse('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });
});
