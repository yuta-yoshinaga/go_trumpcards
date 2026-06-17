import { describe, expect, it } from 'vitest';
import { parseTeenPattiCommand, TEEN_PATTI_HELP } from './teenPattiCommands';

describe('parseTeenPattiCommand', () => {
  it('parses see (short and long)', () => {
    expect(parseTeenPattiCommand('s')).toEqual({ args: ['see'] });
    expect(parseTeenPattiCommand('see')).toEqual({ args: ['see'] });
  });

  it('parses bet (short and long)', () => {
    expect(parseTeenPattiCommand('b')).toEqual({ args: ['bet'] });
    expect(parseTeenPattiCommand('bet')).toEqual({ args: ['bet'] });
  });

  it('parses raise with a stake', () => {
    expect(parseTeenPattiCommand('rs 4')).toEqual({ args: ['raise', { raiseStake: 4 }] });
    expect(parseTeenPattiCommand('raise 10')).toEqual({ args: ['raise', { raiseStake: 10 }] });
  });

  it('returns error for raise without a valid stake', () => {
    expect('error' in parseTeenPattiCommand('rs')).toBe(true);
    expect('error' in parseTeenPattiCommand('rs 0')).toBe(true);
  });

  it('parses fold (short and long)', () => {
    expect(parseTeenPattiCommand('f')).toEqual({ args: ['fold'] });
    expect(parseTeenPattiCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses show (short and long)', () => {
    expect(parseTeenPattiCommand('sh')).toEqual({ args: ['show'] });
    expect(parseTeenPattiCommand('show')).toEqual({ args: ['show'] });
  });

  it('parses sideshow (short and long)', () => {
    expect(parseTeenPattiCommand('ss')).toEqual({ args: ['sideshow'] });
    expect(parseTeenPattiCommand('sideshow')).toEqual({ args: ['sideshow'] });
  });

  it('parses accept into a respond accept=true', () => {
    expect(parseTeenPattiCommand('ac')).toEqual({ args: ['respond', { accept: true }] });
    expect(parseTeenPattiCommand('accept')).toEqual({ args: ['respond', { accept: true }] });
  });

  it('parses decline into a respond accept=false', () => {
    expect(parseTeenPattiCommand('dc')).toEqual({ args: ['respond', { accept: false }] });
    expect(parseTeenPattiCommand('decline')).toEqual({ args: ['respond', { accept: false }] });
  });

  it('parses next (short and long)', () => {
    expect(parseTeenPattiCommand('n')).toEqual({ args: ['next'] });
    expect(parseTeenPattiCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseTeenPattiCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseTeenPattiCommand('sd 9')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parseTeenPattiCommand('sa 5')).toEqual({ args: ['reset', { config: { ante: 5 } }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parseTeenPattiCommand('sa 0')).toBe(true);
  });

  it('parses sc into a reset with starting-chips config', () => {
    expect(parseTeenPattiCommand('sc 200')).toEqual({ args: ['reset', { config: { startingChips: 200 } }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parseTeenPattiCommand('sc 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseTeenPattiCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTeenPattiCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseTeenPattiCommand('l')).toEqual({ args: ['log'] });
    expect(parseTeenPattiCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseTeenPattiCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTeenPattiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseTeenPattiCommand('bett');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseTeenPattiCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(TEEN_PATTI_HELP.length).toBeGreaterThan(0);
    expect(TEEN_PATTI_HELP.some((line) => line.toLowerCase().includes('see'))).toBe(true);
    expect(TEEN_PATTI_HELP.some((line) => line.toLowerCase().includes('side show'))).toBe(true);
  });
});
