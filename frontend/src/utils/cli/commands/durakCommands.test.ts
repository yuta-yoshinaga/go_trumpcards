import { describe, expect, it } from 'vitest';
import { parseDurakCommand } from './durakCommands';

describe('parseDurakCommand', () => {
  it('parses attack with a card index', () => {
    expect(parseDurakCommand('attack 2')).toEqual({ args: ['attack', 2] });
    expect(parseDurakCommand('a 0')).toEqual({ args: ['attack', 0] });
  });

  it('errors on attack without a valid index', () => {
    expect('error' in parseDurakCommand('attack')).toBe(true);
    expect('error' in parseDurakCommand('a abc')).toBe(true);
  });

  it('parses defend with card and pair indices', () => {
    expect(parseDurakCommand('defend 1 0')).toEqual({ args: ['defend', 1, 0] });
    expect(parseDurakCommand('d 3 2')).toEqual({ args: ['defend', 3, 2] });
  });

  it('errors on defend missing an index', () => {
    expect('error' in parseDurakCommand('defend 1')).toBe(true);
    expect('error' in parseDurakCommand('d')).toBe(true);
  });

  it('parses pass and take', () => {
    expect(parseDurakCommand('pass')).toEqual({ args: ['pass'] });
    expect(parseDurakCommand('p')).toEqual({ args: ['pass'] });
    expect(parseDurakCommand('take')).toEqual({ args: ['take'] });
    expect(parseDurakCommand('t')).toEqual({ args: ['take'] });
  });

  it('parses sort by suit and value', () => {
    expect(parseDurakCommand('sort suit')).toEqual({ args: ['sort', undefined, undefined, undefined, 0] });
    expect(parseDurakCommand('sort s')).toEqual({ args: ['sort', undefined, undefined, undefined, 0] });
    expect(parseDurakCommand('sort value')).toEqual({ args: ['sort', undefined, undefined, undefined, 1] });
    expect(parseDurakCommand('sort v')).toEqual({ args: ['sort', undefined, undefined, undefined, 1] });
    expect(parseDurakCommand('sort 1')).toEqual({ args: ['sort', undefined, undefined, undefined, 1] });
  });

  it('errors on an invalid sort mode', () => {
    expect('error' in parseDurakCommand('sort')).toBe(true);
    expect('error' in parseDurakCommand('sort rank')).toBe(true);
  });

  it('parses reset', () => {
    expect(parseDurakCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseDurakCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseDurakCommand('atack 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('attack');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseDurakCommand('zzz')).toBe(true);
  });
});
