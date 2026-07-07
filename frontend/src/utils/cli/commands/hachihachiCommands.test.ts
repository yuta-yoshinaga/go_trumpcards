import { describe, expect, it } from 'vitest';
import { HACHIHACHI_HELP, parseHachiHachiCommand } from './hachihachiCommands';

describe('parseHachiHachiCommand', () => {
  it('parses play with a hand index only', () => {
    expect(parseHachiHachiCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseHachiHachiCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('parses play with a hand index and field index', () => {
    expect(parseHachiHachiCommand('p 1 3')).toEqual({ args: ['play', { cardIndex: 1, fieldIndex: 3 }] });
  });

  it('errors on play without a hand index', () => {
    expect(parseHachiHachiCommand('p')).toEqual({ error: 'Usage: p <handIdx> [fieldIdx]' });
  });

  it('errors on a non-numeric field index', () => {
    expect(parseHachiHachiCommand('p 1 x')).toEqual({ error: 'Invalid field index: x' });
  });

  it('parses next-round aliases', () => {
    expect(parseHachiHachiCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseHachiHachiCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseHachiHachiCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseHachiHachiCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint and reset', () => {
    expect(parseHachiHachiCommand('h')).toEqual({ args: ['hint'] });
    expect(parseHachiHachiCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseHachiHachiCommand('r')).toEqual({ args: ['reset'] });
    expect(parseHachiHachiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const res = parseHachiHachiCommand('rese');
    expect('error' in res && res.error).toContain('Did you mean');
  });

  it('errors on an unknown command', () => {
    expect(parseHachiHachiCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('exposes help text', () => {
    expect(HACHIHACHI_HELP.length).toBeGreaterThan(0);
  });
});
