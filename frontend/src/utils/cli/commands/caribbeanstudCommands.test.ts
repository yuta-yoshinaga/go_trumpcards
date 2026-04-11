import { describe, expect, it } from 'vitest';
import { parseCaribbeanstudCommand } from './caribbeanstudCommands';

describe('parseCaribbeanstudCommand', () => {
  it('parses bet with amount', () => {
    expect(parseCaribbeanstudCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseCaribbeanstudCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with amount and jackpot side bet', () => {
    expect(parseCaribbeanstudCommand('b 100 10')).toEqual({ args: ['bet', 100, 10] });
  });

  it('returns error for bet without amount', () => {
    const result = parseCaribbeanstudCommand('b');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid jackpot value', () => {
    const result = parseCaribbeanstudCommand('b 100 abc');
    expect('error' in result).toBe(true);
  });

  it('parses play', () => {
    expect(parseCaribbeanstudCommand('p')).toEqual({ args: ['play'] });
    expect(parseCaribbeanstudCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseCaribbeanstudCommand('f')).toEqual({ args: ['fold'] });
    expect(parseCaribbeanstudCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseCaribbeanstudCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseCaribbeanstudCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCaribbeanstudCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseCaribbeanstudCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('suggests command for typo', () => {
    const result = parseCaribbeanstudCommand('bett 100');
    expect('error' in result).toBe(true);
  });
});
