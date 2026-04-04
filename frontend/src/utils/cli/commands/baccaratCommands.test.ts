import { describe, expect, it } from 'vitest';
import { parseBaccaratCommand } from './baccaratCommands';

describe('parseBaccaratCommand', () => {
  it('parses bet with amount and type', () => {
    expect(parseBaccaratCommand('b 100 0')).toEqual({ args: ['bet', 100, 0] });
    expect(parseBaccaratCommand('bet 50 1')).toEqual({ args: ['bet', 50, 1] });
    expect(parseBaccaratCommand('b 200 2')).toEqual({ args: ['bet', 200, 2] });
  });

  it('returns error for bet without enough args', () => {
    const result = parseBaccaratCommand('b');
    expect('error' in result).toBe(true);
    const result2 = parseBaccaratCommand('b 100');
    expect('error' in result2).toBe(true);
  });

  it('parses player pair side bet', () => {
    expect(parseBaccaratCommand('pp 50')).toEqual({ args: ['bet', undefined, undefined, 50] });
    expect(parseBaccaratCommand('playerpair 25')).toEqual({ args: ['bet', undefined, undefined, 25] });
  });

  it('returns error for player pair without amount', () => {
    const result = parseBaccaratCommand('pp');
    expect('error' in result).toBe(true);
  });

  it('parses banker pair side bet', () => {
    expect(parseBaccaratCommand('bp 50')).toEqual({ args: ['bet', undefined, undefined, undefined, 50] });
    expect(parseBaccaratCommand('bankerpair 25')).toEqual({ args: ['bet', undefined, undefined, undefined, 25] });
  });

  it('returns error for banker pair without amount', () => {
    const result = parseBaccaratCommand('bp');
    expect('error' in result).toBe(true);
  });

  it('parses log', () => {
    expect(parseBaccaratCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses clearhistory', () => {
    expect(parseBaccaratCommand('cl')).toEqual({ args: ['clearhistory'] });
    expect(parseBaccaratCommand('clearhistory')).toEqual({ args: ['clearhistory'] });
  });

  it('parses reset', () => {
    expect(parseBaccaratCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBaccaratCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseBaccaratCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
