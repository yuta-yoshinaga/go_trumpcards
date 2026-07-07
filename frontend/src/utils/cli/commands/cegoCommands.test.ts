import { describe, expect, it } from 'vitest';
import { CEGO_HELP, parseCegoCommand } from './cegoCommands';

describe('parseCegoCommand', () => {
  it('parses the play bid (full and shorthand)', () => {
    expect(parseCegoCommand('bid play')).toEqual({ args: ['bid', { bid: 'play' }] });
    expect(parseCegoCommand('b p')).toEqual({ args: ['bid', { bid: 'play' }] });
    expect(parseCegoCommand('bid')).toEqual({ args: ['bid', { bid: 'play' }] });
  });

  it('rejects an unknown bid argument', () => {
    expect(parseCegoCommand('bid solo')).toHaveProperty('error');
  });

  it('parses pass', () => {
    expect(parseCegoCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses the contract choice both ways', () => {
    expect(parseCegoCommand('contract cego')).toEqual({ args: ['contract', { contract: 'cego' }] });
    expect(parseCegoCommand('ct handspiel')).toEqual({ args: ['contract', { contract: 'handspiel' }] });
    expect(parseCegoCommand('cego')).toEqual({ args: ['contract', { contract: 'cego' }] });
    expect(parseCegoCommand('handspiel')).toEqual({ args: ['contract', { contract: 'handspiel' }] });
    expect(parseCegoCommand('solo')).toEqual({ args: ['contract', { contract: 'handspiel' }] });
  });

  it('rejects an invalid contract argument', () => {
    expect(parseCegoCommand('contract')).toHaveProperty('error');
    expect(parseCegoCommand('ct zzz')).toHaveProperty('error');
  });

  it('parses the Cego exchange keep command (keep 1 card)', () => {
    expect(parseCegoCommand('keep 3')).toEqual({ args: ['discard', { cardIndices: [3] }] });
    expect(parseCegoCommand('discard 0')).toEqual({ args: ['discard', { cardIndices: [0] }] });
    expect(parseCegoCommand('d 5')).toEqual({ args: ['discard', { cardIndices: [5] }] });
  });

  it('rejects a keep without a valid index', () => {
    expect(parseCegoCommand('keep')).toHaveProperty('error');
    expect(parseCegoCommand('keep x')).toHaveProperty('error');
  });

  it('parses play with alias', () => {
    expect(parseCegoCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parseCegoCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a valid index', () => {
    expect(parseCegoCommand('p')).toHaveProperty('error');
    expect(parseCegoCommand('play x')).toHaveProperty('error');
  });

  it('parses next / nextround / hint / reset with aliases', () => {
    expect(parseCegoCommand('n')).toEqual({ args: ['next'] });
    expect(parseCegoCommand('next')).toEqual({ args: ['next'] });
    expect(parseCegoCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCegoCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseCegoCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCegoCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseCegoCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCegoCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const res = parseCegoCommand('discrd 0');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('discard');
  });

  it('reports an unknown command with no close match', () => {
    const res = parseCegoCommand('zzz');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(CEGO_HELP.length).toBeGreaterThan(0);
    expect(CEGO_HELP.some((l) => l.includes('contract'))).toBe(true);
  });
});
