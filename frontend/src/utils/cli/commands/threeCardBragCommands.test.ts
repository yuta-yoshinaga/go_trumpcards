import { describe, expect, it } from 'vitest';
import { parseThreeCardBragCommand, THREE_CARD_BRAG_HELP } from './threeCardBragCommands';

describe('parseThreeCardBragCommand', () => {
  it('parses see (short and long)', () => {
    expect(parseThreeCardBragCommand('s')).toEqual({ args: ['see'] });
    expect(parseThreeCardBragCommand('see')).toEqual({ args: ['see'] });
  });

  it('parses bet (short and long)', () => {
    expect(parseThreeCardBragCommand('b')).toEqual({ args: ['bet'] });
    expect(parseThreeCardBragCommand('bet')).toEqual({ args: ['bet'] });
  });

  it('parses raise with a stake', () => {
    expect(parseThreeCardBragCommand('rs 4')).toEqual({ args: ['raise', { raiseStake: 4 }] });
    expect(parseThreeCardBragCommand('raise 10')).toEqual({ args: ['raise', { raiseStake: 10 }] });
  });

  it('returns error for raise without a valid stake', () => {
    expect('error' in parseThreeCardBragCommand('rs')).toBe(true);
    expect('error' in parseThreeCardBragCommand('rs 0')).toBe(true);
  });

  it('parses fold (short and long)', () => {
    expect(parseThreeCardBragCommand('f')).toEqual({ args: ['fold'] });
    expect(parseThreeCardBragCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses show (short and long)', () => {
    expect(parseThreeCardBragCommand('sh')).toEqual({ args: ['show'] });
    expect(parseThreeCardBragCommand('show')).toEqual({ args: ['show'] });
  });

  it('parses next (short and long)', () => {
    expect(parseThreeCardBragCommand('n')).toEqual({ args: ['next'] });
    expect(parseThreeCardBragCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseThreeCardBragCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseThreeCardBragCommand('sd 9')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parseThreeCardBragCommand('sa 5')).toEqual({ args: ['reset', { config: { ante: 5 } }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parseThreeCardBragCommand('sa 0')).toBe(true);
  });

  it('parses sc into a reset with starting-chips config', () => {
    expect(parseThreeCardBragCommand('sc 200')).toEqual({ args: ['reset', { config: { startingChips: 200 } }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parseThreeCardBragCommand('sc 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseThreeCardBragCommand('h')).toEqual({ args: ['hint'] });
    expect(parseThreeCardBragCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseThreeCardBragCommand('l')).toEqual({ args: ['log'] });
    expect(parseThreeCardBragCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseThreeCardBragCommand('r')).toEqual({ args: ['reset'] });
    expect(parseThreeCardBragCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseThreeCardBragCommand('bett');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseThreeCardBragCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(THREE_CARD_BRAG_HELP.length).toBeGreaterThan(0);
    expect(THREE_CARD_BRAG_HELP.some((line) => line.toLowerCase().includes('see'))).toBe(true);
  });
});
