import { describe, expect, it } from 'vitest';
import { FARO_HELP, parseFaroCommand } from './faroCommands';

describe('parseFaroCommand', () => {
  it('parses a bet (short and long) with an optional copper flag', () => {
    expect(parseFaroCommand('b 7 100')).toEqual({ args: ['bet', { rank: 7, amount: 100, copper: false }] });
    expect(parseFaroCommand('bet 13 50 c')).toEqual({ args: ['bet', { rank: 13, amount: 50, copper: true }] });
  });

  it('rejects an invalid bet', () => {
    expect('error' in parseFaroCommand('b')).toBe(true);
    expect('error' in parseFaroCommand('b 0 100')).toBe(true);
    expect('error' in parseFaroCommand('b 14 100')).toBe(true);
    expect('error' in parseFaroCommand('b 7 0')).toBe(true);
  });

  it('parses clearBet, clearAll, and deal', () => {
    expect(parseFaroCommand('cb 5')).toEqual({ args: ['clearBet', { rank: 5 }] });
    expect('error' in parseFaroCommand('cb 99')).toBe(true);
    expect(parseFaroCommand('ca')).toEqual({ args: ['clearAll'] });
    expect(parseFaroCommand('d')).toEqual({ args: ['deal'] });
    expect(parseFaroCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses a full call and skips when fewer than 3 ranks', () => {
    expect(parseFaroCommand('call 3 9 12')).toEqual({ args: ['call', { order: [3, 9, 12] }] });
    expect(parseFaroCommand('call')).toEqual({ args: ['call', { order: [] }] });
    expect('error' in parseFaroCommand('call 3 9 99')).toBe(true);
  });

  it('parses next, log, and reset (short and long)', () => {
    expect(parseFaroCommand('n')).toEqual({ args: ['next'] });
    expect(parseFaroCommand('next')).toEqual({ args: ['next'] });
    expect(parseFaroCommand('l')).toEqual({ args: ['log'] });
    expect(parseFaroCommand('r')).toEqual({ args: ['reset'] });
    expect(parseFaroCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command and errors on unknown', () => {
    const near = parseFaroCommand('bett');
    expect('error' in near).toBe(true);
    if ('error' in near) expect(near.error).toContain('Did you mean');
    const unknown = parseFaroCommand('xyz');
    expect('error' in unknown).toBe(true);
    if ('error' in unknown) expect(unknown.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(FARO_HELP.length).toBeGreaterThan(0);
    expect(FARO_HELP.some((line) => line.toLowerCase().includes('bet'))).toBe(true);
  });
});
