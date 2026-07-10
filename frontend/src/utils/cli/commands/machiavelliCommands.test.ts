import { describe, expect, it } from 'vitest';
import { parseMachiavelliCommand } from './machiavelliCommands';

describe('parseMachiavelliCommand', () => {
  it('parses draw', () => {
    expect(parseMachiavelliCommand('dr')).toEqual({ args: ['draw'] });
    expect(parseMachiavelliCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses newmeld with 3+ indices', () => {
    expect(parseMachiavelliCommand('nm 0 1 2')).toEqual({ args: ['newmeld', { handIndices: [0, 1, 2] }] });
    expect(parseMachiavelliCommand('newmeld 3 4 5 6')).toEqual({ args: ['newmeld', { handIndices: [3, 4, 5, 6] }] });
  });

  it('returns error for newmeld with fewer than 3 indices', () => {
    expect('error' in parseMachiavelliCommand('nm 0 1')).toBe(true);
  });

  it('returns error for newmeld with a non-numeric index', () => {
    expect('error' in parseMachiavelliCommand('nm 0 x 2')).toBe(true);
  });

  it('parses layoff with meld and hand indices', () => {
    expect(parseMachiavelliCommand('lo 1 4')).toEqual({ args: ['layoff', { meldIdx: 1, handIndex: 4 }] });
    expect(parseMachiavelliCommand('layoff 0 2')).toEqual({ args: ['layoff', { meldIdx: 0, handIndex: 2 }] });
  });

  it('returns error for layoff without both indices', () => {
    expect('error' in parseMachiavelliCommand('lo 1')).toBe(true);
    expect('error' in parseMachiavelliCommand('lo')).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseMachiavelliCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMachiavelliCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseMachiavelliCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseMachiavelliCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMachiavelliCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseMachiavelliCommand('drw');
    expect('error' in result && result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    expect('error' in parseMachiavelliCommand('xyz')).toBe(true);
  });
});
