import { describe, expect, it } from 'vitest';
import { parseNapoleonCommand } from './napoleonCommands';

describe('parseNapoleonCommand', () => {
  it('parses bid with number', () => {
    expect(parseNapoleonCommand('bid 5')).toEqual({ args: ['bid', 5] });
  });

  it('returns error for bid without number', () => {
    const result = parseNapoleonCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses pass', () => {
    expect(parseNapoleonCommand('pass')).toEqual({ args: ['bid', 0] });
  });

  it('parses trump with suit', () => {
    expect(parseNapoleonCommand('trump spade')).toEqual({ args: ['trump', undefined, 1] });
    expect(parseNapoleonCommand('trump heart')).toEqual({ args: ['trump', undefined, 3] });
  });

  it('returns error for trump without suit', () => {
    const result = parseNapoleonCommand('trump');
    expect('error' in result).toBe(true);
  });

  it('returns error for trump with invalid suit', () => {
    const result = parseNapoleonCommand('trump invalid');
    expect('error' in result).toBe(true);
  });

  it('parses adjutant with suit and value', () => {
    expect(parseNapoleonCommand('adj spade 1')).toEqual({ args: ['trump', undefined, undefined, 1, 1] });
    expect(parseNapoleonCommand('adjutant heart 12')).toEqual({ args: ['trump', undefined, undefined, 3, 12] });
  });

  it('returns error for adjutant without enough args', () => {
    const result = parseNapoleonCommand('adj');
    expect('error' in result).toBe(true);
    const result2 = parseNapoleonCommand('adj spade');
    expect('error' in result2).toBe(true);
  });

  it('returns error for adjutant with invalid suit', () => {
    const result = parseNapoleonCommand('adj invalid 1');
    expect('error' in result).toBe(true);
  });

  it('parses exchange with index', () => {
    expect(parseNapoleonCommand('ex 2')).toEqual({ args: ['exchange', undefined, undefined, undefined, undefined, 2] });
    expect(parseNapoleonCommand('exchange 4')).toEqual({
      args: ['exchange', undefined, undefined, undefined, undefined, 4],
    });
  });

  it('returns error for exchange without index', () => {
    const result = parseNapoleonCommand('ex');
    expect('error' in result).toBe(true);
  });

  it('parses log', () => {
    expect(parseNapoleonCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses play from shared trick commands', () => {
    expect(parseNapoleonCommand('p 3')).toEqual({
      args: ['play', undefined, undefined, undefined, undefined, undefined, 3],
    });
  });

  it('parses reset', () => {
    expect(parseNapoleonCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseNapoleonCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
