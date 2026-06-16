import { describe, expect, it } from 'vitest';
import { ECARTE_HELP, parseEcarteCommand } from './ecarteCommands';

describe('parseEcarteCommand', () => {
  it('parses play (short and long)', () => {
    expect(parseEcarteCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseEcarteCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseEcarteCommand('p')).toBe(true);
  });

  it('parses propose (short and long)', () => {
    expect(parseEcarteCommand('pr')).toEqual({ args: ['propose'] });
    expect(parseEcarteCommand('propose')).toEqual({ args: ['propose'] });
  });

  it('parses stand (short and long)', () => {
    expect(parseEcarteCommand('st')).toEqual({ args: ['stand'] });
    expect(parseEcarteCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses accept into a respond(true)', () => {
    expect(parseEcarteCommand('a')).toEqual({ args: ['respond', { accept: true }] });
    expect(parseEcarteCommand('accept')).toEqual({ args: ['respond', { accept: true }] });
  });

  it('parses refuse into a respond(false)', () => {
    expect(parseEcarteCommand('rf')).toEqual({ args: ['respond', { accept: false }] });
    expect(parseEcarteCommand('refuse')).toEqual({ args: ['respond', { accept: false }] });
  });

  it('parses discard with multiple indices', () => {
    expect(parseEcarteCommand('d 0 2 4')).toEqual({ args: ['discard', { discardIndices: [0, 2, 4] }] });
    expect(parseEcarteCommand('discard 1')).toEqual({ args: ['discard', { discardIndices: [1] }] });
  });

  it('parses discard with no indices as an empty discard', () => {
    expect(parseEcarteCommand('d')).toEqual({ args: ['discard', { discardIndices: [] }] });
  });

  it('parses next (short and long)', () => {
    expect(parseEcarteCommand('n')).toEqual({ args: ['next'] });
    expect(parseEcarteCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseEcarteCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseEcarteCommand('sd 9')).toBe(true);
  });

  it('parses tg into a reset with target-score config', () => {
    expect(parseEcarteCommand('tg 5')).toEqual({ args: ['reset', { config: { targetScore: 5 } }] });
  });

  it('rejects an out-of-range tg', () => {
    expect('error' in parseEcarteCommand('tg 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseEcarteCommand('h')).toEqual({ args: ['hint'] });
    expect(parseEcarteCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseEcarteCommand('l')).toEqual({ args: ['log'] });
    expect(parseEcarteCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseEcarteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseEcarteCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseEcarteCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseEcarteCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(ECARTE_HELP.length).toBeGreaterThan(0);
    expect(ECARTE_HELP.some((line) => line.toLowerCase().includes('propose'))).toBe(true);
  });
});
