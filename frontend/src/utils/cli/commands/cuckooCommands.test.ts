import { describe, expect, it } from 'vitest';
import { CUCKOO_HELP, parseCuckooCommand } from './cuckooCommands';

describe('parseCuckooCommand', () => {
  it('parses keep (short and long)', () => {
    expect(parseCuckooCommand('k')).toEqual({ args: ['keep'] });
    expect(parseCuckooCommand('keep')).toEqual({ args: ['keep'] });
  });

  it('parses swap (short and long)', () => {
    expect(parseCuckooCommand('s')).toEqual({ args: ['swap'] });
    expect(parseCuckooCommand('swap')).toEqual({ args: ['swap'] });
  });

  it('parses refuse (short and long)', () => {
    expect(parseCuckooCommand('rf')).toEqual({ args: ['refuse'] });
    expect(parseCuckooCommand('refuse')).toEqual({ args: ['refuse'] });
  });

  it('parses accept (short and long)', () => {
    expect(parseCuckooCommand('ac')).toEqual({ args: ['accept'] });
    expect(parseCuckooCommand('accept')).toEqual({ args: ['accept'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseCuckooCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCuckooCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseCuckooCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    expect('error' in parseCuckooCommand('sd 9')).toBe(true);
    expect('error' in parseCuckooCommand('sd')).toBe(true);
  });

  it('parses log and reset (short and long)', () => {
    expect(parseCuckooCommand('l')).toEqual({ args: ['log'] });
    expect(parseCuckooCommand('log')).toEqual({ args: ['log'] });
    expect(parseCuckooCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCuckooCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseCuckooCommand('kee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseCuckooCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(CUCKOO_HELP.length).toBeGreaterThan(0);
    expect(CUCKOO_HELP.some((line) => line.toLowerCase().includes('keep'))).toBe(true);
    expect(CUCKOO_HELP.some((line) => line.toLowerCase().includes('swap'))).toBe(true);
  });
});
