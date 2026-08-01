import { describe, expect, it } from 'vitest';
import { KARNOFFEL_HELP, parseKarnoffelCommand } from './karnoffelCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseKarnoffelCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseKarnoffelCommand', () => {
  it('parses the parameterless actions and their aliases', () => {
    for (const [input, want] of [
      ['n', 'next'],
      ['next', 'next'],
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseKarnoffelCommand(input)).toEqual({ args: [want] });
    }
  });

  // **手札は 5 枚。**issue の 12 枚なら 11 まで通ってしまう。
  it('parses a play and rejects an index outside the five-card hand', () => {
    expect(parseKarnoffelCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseKarnoffelCommand('play 4')).toEqual({ args: ['play', { cardIndex: 4 }] });
    for (const input of ['p', 'p abc', 'p -1', 'p 5', 'p 11']) {
      expect(parseError(input)).toContain('Usage: p <0-4>');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('nex')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['p <0-4>', 'n / next', 'l / log', 'r / reset']) {
      expect(KARNOFFEL_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // 5 枚配りと追随不要を help でも言う。
    expect(KARNOFFEL_HELP.some((line) => line.includes('five each'))).toBe(true);
    expect(KARNOFFEL_HELP.some((line) => line.includes('NO obligation to follow suit'))).toBe(true);
  });
});
