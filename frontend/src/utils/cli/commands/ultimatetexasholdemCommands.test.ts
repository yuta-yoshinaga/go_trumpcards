import { describe, expect, it } from 'vitest';
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
    const constraints = {
      maxAnte: 100,
      maxTrips: 0,
      boundsForAnte: (ante: number, trips: number) => ({ maxTrips: 200 - 2 * ante, valid: 2 * ante + trips <= 200 }),
      phase: 1,
    };
    expect(parse('bet 0', constraints)).toHaveProperty('error');
    expect(parse('bet 15', constraints)).toHaveProperty('error');
    expect(parse('bet 100 0', constraints)).toEqual({ args: ['bet', 100, 0] });
    expect(parse('bet 110', constraints)).toHaveProperty('error');
    expect(parse('bet 90 20', constraints)).toEqual({ args: ['bet', 90, 20] });
    expect(parse('bet 90 30', constraints)).toHaveProperty('error');
    const outOfStep = parse('bet 15', constraints);
    expect('error' in outOfStep && outOfStep.error).toContain('range');
  });
  it('rejects phase and chip invalid actions', () => {
    expect(parse('4', { maxAnte: 0, maxTrips: 0, phase: 3, canPlay4x: true })).toHaveProperty('error');
    expect(parse('4', { maxAnte: 0, maxTrips: 0, phase: 2, canPlay4x: false })).toHaveProperty('error');
    expect(parse('check', { maxAnte: 0, maxTrips: 0, phase: 4 })).toHaveProperty('error');
    expect(parse('fold', { maxAnte: 0, maxTrips: 0, phase: 3 })).toHaveProperty('error');
    expect(parse('reset', { maxAnte: 0, maxTrips: 0, phase: 2 })).toHaveProperty('error');
    expect(parse('check', { maxAnte: 0, maxTrips: 0, phase: 2 })).toEqual({ args: ['check'] });
    expect(parse('fold', { maxAnte: 0, maxTrips: 0, phase: 4 })).toEqual({ args: ['fold'] });
    expect(parse('reset', { maxAnte: 0, maxTrips: 0, phase: 5 })).toEqual({ args: ['reset'] });
  });
  it('rejects unsupported API commands and unknown commands', () => {
    expect(parse('hint')).toHaveProperty('error');
    expect(parse('log')).toHaveProperty('error');
    expect(parse('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });
});
