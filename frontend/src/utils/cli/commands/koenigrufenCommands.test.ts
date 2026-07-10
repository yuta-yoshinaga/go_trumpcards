import { describe, expect, it } from 'vitest';
import { KOENIGRUFEN_HELP, parseKoenigrufenCommand } from './koenigrufenCommands';

describe('parseKoenigrufenCommand', () => {
  it('parses the rufer bid (full and shorthand)', () => {
    expect(parseKoenigrufenCommand('bid rufer')).toEqual({ args: ['bid', { bid: 'rufer' }] });
    expect(parseKoenigrufenCommand('b r')).toEqual({ args: ['bid', { bid: 'rufer' }] });
  });

  it('rejects an unknown or missing bid', () => {
    expect(parseKoenigrufenCommand('bid')).toHaveProperty('error');
    expect(parseKoenigrufenCommand('bid petite')).toHaveProperty('error');
  });

  it('parses pass', () => {
    expect(parseKoenigrufenCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses callking by number', () => {
    expect(parseKoenigrufenCommand('callking 1')).toEqual({ args: ['callking', { callSuit: 1 }] });
    expect(parseKoenigrufenCommand('ck 4')).toEqual({ args: ['callking', { callSuit: 4 }] });
  });

  it('parses callking by suit name', () => {
    expect(parseKoenigrufenCommand('callking spade')).toEqual({ args: ['callking', { callSuit: 1 }] });
    expect(parseKoenigrufenCommand('ck club')).toEqual({ args: ['callking', { callSuit: 2 }] });
    expect(parseKoenigrufenCommand('ck heart')).toEqual({ args: ['callking', { callSuit: 3 }] });
    expect(parseKoenigrufenCommand('ck diamond')).toEqual({ args: ['callking', { callSuit: 4 }] });
  });

  it('rejects callking with an invalid suit', () => {
    expect(parseKoenigrufenCommand('ck')).toHaveProperty('error');
    expect(parseKoenigrufenCommand('ck 9')).toHaveProperty('error');
    expect(parseKoenigrufenCommand('ck zzz')).toHaveProperty('error');
  });

  it('parses a 6-card discard (talon)', () => {
    expect(parseKoenigrufenCommand('discard 0 1 2 3 4 5')).toEqual({
      args: ['discard', { cardIndices: [0, 1, 2, 3, 4, 5] }],
    });
    expect(parseKoenigrufenCommand('d 6 7 8 9 10 11')).toEqual({
      args: ['discard', { cardIndices: [6, 7, 8, 9, 10, 11] }],
    });
  });

  it('rejects a discard with fewer than 6 indices', () => {
    expect(parseKoenigrufenCommand('discard 0 1 2')).toHaveProperty('error');
    expect(parseKoenigrufenCommand('discard 0 1 2 3 4 x')).toHaveProperty('error');
  });

  it('parses play with alias', () => {
    expect(parseKoenigrufenCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parseKoenigrufenCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a valid index', () => {
    expect(parseKoenigrufenCommand('p')).toHaveProperty('error');
    expect(parseKoenigrufenCommand('play x')).toHaveProperty('error');
  });

  it('parses next / nextround / hint / reset with aliases', () => {
    expect(parseKoenigrufenCommand('n')).toEqual({ args: ['next'] });
    expect(parseKoenigrufenCommand('next')).toEqual({ args: ['next'] });
    expect(parseKoenigrufenCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseKoenigrufenCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseKoenigrufenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseKoenigrufenCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseKoenigrufenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseKoenigrufenCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const res = parseKoenigrufenCommand('discrd 0 1 2 3 4 5');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('discard');
  });

  it('reports an unknown command with no close match', () => {
    const res = parseKoenigrufenCommand('zzz');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(KOENIGRUFEN_HELP.length).toBeGreaterThan(0);
    expect(KOENIGRUFEN_HELP.some((l) => l.includes('callking'))).toBe(true);
  });
});
