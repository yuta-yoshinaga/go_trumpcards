import { describe, expect, it } from 'vitest';
import { parseScartoCommand, SCARTO_HELP } from './scartoCommands';

describe('parseScartoCommand', () => {
  it('parses a 3-card scarto (full and short forms)', () => {
    expect(parseScartoCommand('scarto 0 1 2')).toEqual({ args: ['scarto', { cardIndices: [0, 1, 2] }] });
    expect(parseScartoCommand('s 3 4 5')).toEqual({ args: ['scarto', { cardIndices: [3, 4, 5] }] });
    expect(parseScartoCommand('discard 6 7 8')).toEqual({ args: ['scarto', { cardIndices: [6, 7, 8] }] });
    expect(parseScartoCommand('d 10 11 12')).toEqual({ args: ['scarto', { cardIndices: [10, 11, 12] }] });
  });

  it('rejects a scarto with fewer than 3 indices', () => {
    expect(parseScartoCommand('scarto 0 1')).toHaveProperty('error');
    expect(parseScartoCommand('s')).toHaveProperty('error');
  });

  it('parses play (full and short forms)', () => {
    expect(parseScartoCommand('play 4')).toEqual({ args: ['play', { cardIndex: 4 }] });
    expect(parseScartoCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects a play with a missing index', () => {
    expect(parseScartoCommand('play')).toHaveProperty('error');
  });

  it('parses next / nextround (full and short forms)', () => {
    expect(parseScartoCommand('next')).toEqual({ args: ['next'] });
    expect(parseScartoCommand('n')).toEqual({ args: ['next'] });
    expect(parseScartoCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseScartoCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses hint and reset', () => {
    expect(parseScartoCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseScartoCommand('h')).toEqual({ args: ['hint'] });
    expect(parseScartoCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseScartoCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const res = parseScartoCommand('paly 0');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('Did you mean');
  });

  it('reports a fully unknown command', () => {
    const res = parseScartoCommand('zzz');
    expect(res).toHaveProperty('error');
  });

  it('exposes non-empty help text', () => {
    expect(SCARTO_HELP.length).toBeGreaterThan(0);
  });
});
