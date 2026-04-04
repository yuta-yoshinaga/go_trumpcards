import { describe, expect, it } from 'vitest';
import { parseThreecardCommand } from './threecardCommands';

describe('parseThreecardCommand', () => {
  it('parses bet with amount', () => {
    expect(parseThreecardCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseThreecardCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with amount and pair plus bet', () => {
    expect(parseThreecardCommand('b 100 50')).toEqual({ args: ['bet', 100, 50] });
  });

  it('returns error for bet without amount', () => {
    const result = parseThreecardCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses play', () => {
    expect(parseThreecardCommand('p')).toEqual({ args: ['play'] });
    expect(parseThreecardCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseThreecardCommand('f')).toEqual({ args: ['fold'] });
    expect(parseThreecardCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseThreecardCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseThreecardCommand('r')).toEqual({ args: ['reset'] });
    expect(parseThreecardCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseThreecardCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
