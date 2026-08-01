import { describe, expect, it } from 'vitest';
import { GUANDAN_HELP, parseGuandanCommand } from './guandanCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseGuandanCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseGuandanCommand', () => {
  it('parses the parameterless actions and their aliases', () => {
    for (const [input, want] of [
      ['ps', 'pass'],
      ['pass', 'pass'],
      ['n', 'next'],
      ['next', 'next'],
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseGuandanCommand(input)).toEqual({ args: [want] });
    }
  });

  // **1 手は札 1 枚とは限らない。**役はまとめて出す。
  it('plays several cards as one combination', () => {
    expect(parseGuandanCommand('p 3')).toEqual({ args: ['play', { cardIndexes: [3] }] });
    expect(parseGuandanCommand('play 0 1 2')).toEqual({ args: ['play', { cardIndexes: [0, 1, 2] }] });
  });

  it('rejects a bad play', () => {
    expect(parseError('p')).toContain('Usage: p <index>');
    for (const input of ['p abc', 'p -1', 'p 27']) {
      expect(parseError(input)).toContain('every index is 0-26');
    }
  });

  // **同じ札を 2 回数えられない。**通すと 1 枚からペアが作れてしまう。
  it('rejects the same index given twice', () => {
    expect(parseError('p 1 1')).toContain('index 1 was given twice');
    expect(parseError('p 0 2 0')).toContain('index 0 was given twice');
  });

  it('returns a tribute card', () => {
    expect(parseGuandanCommand('t 4')).toEqual({ args: ['tribute', { cardIndex: 4 }] });
    expect(parseGuandanCommand('tribute 0')).toEqual({ args: ['tribute', { cardIndex: 0 }] });
  });

  it('rejects a bad tribute', () => {
    for (const input of ['t', 't abc', 't -1', 't 27']) {
      expect(parseError(input)).toContain('t <index 0-26>');
    }
  });

  it('suggests a near miss and rejects the rest', () => {
    expect(parseError('pas')).toContain('Did you mean');
    expect(parseError('zzz')).toBe('Unknown command: zzz');
  });

  it('documents every action it accepts', () => {
    const help = GUANDAN_HELP.join('\n');
    for (const fragment of ['p <index...>', 'ps / pass', 't <index>', 'n / next', 'r / reset']) {
      expect(help).toContain(fragment);
    }
  });
});
