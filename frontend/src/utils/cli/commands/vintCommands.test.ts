import { describe, expect, it } from 'vitest';
import { parseVintCommand, VINT_HELP } from './vintCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseVintCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseVintCommand', () => {
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
      expect(parseVintCommand(input)).toEqual({ args: [want] });
    }
  });

  // **denom は 0=♠ 1=♣ 2=♦ 3=♥ 4=NT。**ブリッジと序列が逆なので番号で指す。
  it('bids a level and a denomination', () => {
    expect(parseVintCommand('b 3 4')).toEqual({ args: ['bid', { level: 3, denom: 4 }] });
    expect(parseVintCommand('bid 1 0')).toEqual({ args: ['bid', { level: 1, denom: 0 }] });
    expect(parseVintCommand('b 7 3')).toEqual({ args: ['bid', { level: 7, denom: 3 }] });
  });

  it('rejects a bad level or denomination', () => {
    for (const input of ['b', 'b abc 1', 'b 0 1', 'b 8 1']) {
      expect(parseError(input)).toContain('Usage: b');
    }
    for (const input of ['b 3', 'b 3 abc', 'b 3 -1', 'b 3 5']) {
      expect(parseError(input)).toContain('denomination is 0-4');
    }
  });

  it('parses a play and rejects an index outside the hand', () => {
    expect(parseVintCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseVintCommand('play 12')).toEqual({ args: ['play', { cardIndex: 12 }] });
    for (const input of ['p', 'p abc', 'p -1', 'p 13']) {
      expect(parseError(input)).toContain('Usage: p');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('pas')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['b <1-7> <0-4>', 'ps / pass', 'p <0-12>']) {
      expect(VINT_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // ♠ が最弱であることを help でも言う。
    expect(VINT_HELP.some((line) => line.includes('spades LOWEST'))).toBe(true);
  });
});
