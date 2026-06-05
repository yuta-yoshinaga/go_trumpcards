import { describe, expect, it } from 'vitest';
import { parseGongZhuCommand } from './gongzhuCommands';

describe('parseGongZhuCommand', () => {
  it('parses expose with indices', () => {
    expect(parseGongZhuCommand('expose 0 3')).toEqual({ args: ['expose', [0, 3]] });
  });

  it('parses expose with no indices as empty selection', () => {
    expect(parseGongZhuCommand('expose')).toEqual({ args: ['expose', []] });
  });

  it('returns error for expose with invalid index', () => {
    const result = parseGongZhuCommand('expose 0 x');
    expect('error' in result).toBe(true);
  });

  it('parses play', () => {
    expect(parseGongZhuCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseGongZhuCommand('p')).toBe(true);
  });

  it('parses next', () => {
    expect(parseGongZhuCommand('n')).toEqual({ args: ['next'] });
  });

  it('parses nextround', () => {
    expect(parseGongZhuCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses hint', () => {
    expect(parseGongZhuCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset', () => {
    expect(parseGongZhuCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseGongZhuCommand('xyz')).toBe(true);
  });
});
