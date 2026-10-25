import { describe, expect, it } from 'vitest';
import { GAIGEL_HELP, parseGaigelCommand } from './gaigelCommands';

describe('parseGaigelCommand', () => {
  it('matches the page actions for play, marriage and progression', () => {
    expect(parseGaigelCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseGaigelCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
    expect(parseGaigelCommand('m 3')).toEqual({ args: ['marriage', undefined, 3] });
    expect(parseGaigelCommand('marriage 1')).toEqual({ args: ['marriage', undefined, 1] });
    expect(parseGaigelCommand('n')).toEqual({ args: ['next'] });
    expect(parseGaigelCommand('next')).toEqual({ args: ['next'] });
    expect(parseGaigelCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGaigelCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseGaigelCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGaigelCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseGaigelCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGaigelCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing and out of range card indices', () => {
    for (const input of ['p', 'play x', 'p -1', 'p 5']) expect(parseGaigelCommand(input)).toHaveProperty('error');
    for (const input of ['m', 'marriage nope', 'm -1', 'marriage 5'])
      expect(parseGaigelCommand(input)).toHaveProperty('error');
  });

  it('rejects unknown commands and does not expose log', () => {
    expect(parseGaigelCommand('xyz')).toEqual({ error: 'Unknown command: xyz' });
    expect(parseGaigelCommand('log')).toHaveProperty('error');
    expect(GAIGEL_HELP.join('\n')).not.toContain('log');
  });
});
