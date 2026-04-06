import { describe, expect, it } from 'vitest';
import { parseGinrummyCommand } from './ginrummyCommands';

describe('parseGinrummyCommand', () => {
  it('parses drawstock', () => {
    expect(parseGinrummyCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseGinrummyCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseGinrummyCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseGinrummyCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses discard with index', () => {
    expect(parseGinrummyCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseGinrummyCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseGinrummyCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses knock with index', () => {
    expect(parseGinrummyCommand('kn 2')).toEqual({ args: ['knock', 2] });
    expect(parseGinrummyCommand('knock 4')).toEqual({ args: ['knock', 4] });
  });

  it('returns error for knock without index', () => {
    const result = parseGinrummyCommand('kn');
    expect('error' in result).toBe(true);
  });

  it('parses layoff with indices', () => {
    expect(parseGinrummyCommand('lo 0 1 2')).toEqual({ args: ['layoff', undefined, undefined, [0, 1, 2]] });
    expect(parseGinrummyCommand('layoff 3 4')).toEqual({ args: ['layoff', undefined, undefined, [3, 4]] });
  });

  it('parses layoff without indices as empty', () => {
    expect(parseGinrummyCommand('lo')).toEqual({ args: ['layoff', undefined, undefined, []] });
  });

  it('parses nextround', () => {
    expect(parseGinrummyCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGinrummyCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseGinrummyCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseGinrummyCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGinrummyCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseGinrummyCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
