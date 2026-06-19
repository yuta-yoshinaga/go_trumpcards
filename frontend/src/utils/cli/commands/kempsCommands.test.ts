import { describe, expect, it } from 'vitest';
import { KEMPS_HELP, parseKempsCommand } from './kempsCommands';

describe('parseKempsCommand', () => {
  it('parses swap with two indices (short and long)', () => {
    expect(parseKempsCommand('s 1 2')).toEqual({ args: ['swap', { handIndex: 1, fieldIndex: 2 }] });
    expect(parseKempsCommand('swap 0 3')).toEqual({ args: ['swap', { handIndex: 0, fieldIndex: 3 }] });
  });

  it('returns error for swap without two valid indices', () => {
    expect('error' in parseKempsCommand('s')).toBe(true);
    expect('error' in parseKempsCommand('s 1')).toBe(true);
    expect('error' in parseKempsCommand('s -1 2')).toBe(true);
  });

  it('parses pass (short and long)', () => {
    expect(parseKempsCommand('p')).toEqual({ args: ['pass'] });
    expect(parseKempsCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses signal with a type (short and long)', () => {
    expect(parseKempsCommand('sig 1')).toEqual({ args: ['signal', { signalType: 1 }] });
    expect(parseKempsCommand('signal 0')).toEqual({ args: ['signal', { signalType: 0 }] });
  });

  it('rejects an out-of-range signal type', () => {
    expect('error' in parseKempsCommand('sig 2')).toBe(true);
    expect('error' in parseKempsCommand('sig')).toBe(true);
  });

  it('parses kemps (short and long)', () => {
    expect(parseKempsCommand('k')).toEqual({ args: ['kemps'] });
    expect(parseKempsCommand('kemps')).toEqual({ args: ['kemps'] });
  });

  it('parses counter with a target seat (short and long)', () => {
    expect(parseKempsCommand('c 1')).toEqual({ args: ['counter', { targetSeat: 1 }] });
    expect(parseKempsCommand('counter 3')).toEqual({ args: ['counter', { targetSeat: 3 }] });
  });

  it('returns error for counter without a valid seat', () => {
    expect('error' in parseKempsCommand('c')).toBe(true);
    expect('error' in parseKempsCommand('c -1')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseKempsCommand('n')).toEqual({ args: ['next'] });
    expect(parseKempsCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseKempsCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    expect('error' in parseKempsCommand('sd 9')).toBe(true);
  });

  it('parses log and reset (short and long)', () => {
    expect(parseKempsCommand('l')).toEqual({ args: ['log'] });
    expect(parseKempsCommand('log')).toEqual({ args: ['log'] });
    expect(parseKempsCommand('r')).toEqual({ args: ['reset'] });
    expect(parseKempsCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseKempsCommand('kem');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseKempsCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(KEMPS_HELP.length).toBeGreaterThan(0);
    expect(KEMPS_HELP.some((line) => line.toLowerCase().includes('swap'))).toBe(true);
    expect(KEMPS_HELP.some((line) => line.toLowerCase().includes('kemps'))).toBe(true);
  });
});
