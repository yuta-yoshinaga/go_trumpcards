import { describe, expect, it } from 'vitest';
import { parseChinchonCommand } from './chinchonCommands';

describe('parseChinchonCommand', () => {
  it('parses drawstock', () => {
    expect(parseChinchonCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseChinchonCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseChinchonCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseChinchonCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses discard with index', () => {
    expect(parseChinchonCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseChinchonCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseChinchonCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses knock with index', () => {
    expect(parseChinchonCommand('kn 2')).toEqual({ args: ['knock', 2] });
    expect(parseChinchonCommand('knock 4')).toEqual({ args: ['knock', 4] });
  });

  it('returns error for knock without index', () => {
    const result = parseChinchonCommand('kn');
    expect('error' in result).toBe(true);
  });

  it('parses layoff with indices', () => {
    expect(parseChinchonCommand('lo 0 1 2')).toEqual({ args: ['layoff', undefined, undefined, [0, 1, 2]] });
    expect(parseChinchonCommand('layoff 3 4')).toEqual({ args: ['layoff', undefined, undefined, [3, 4]] });
  });

  it('parses layoff without indices as empty', () => {
    expect(parseChinchonCommand('lo')).toEqual({ args: ['layoff', undefined, undefined, []] });
  });

  it('parses nextround', () => {
    expect(parseChinchonCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseChinchonCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseChinchonCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseChinchonCommand('r')).toEqual({ args: ['reset'] });
    expect(parseChinchonCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseChinchonCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
