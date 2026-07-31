import { describe, expect, it } from 'vitest';
import { BOSTON_HELP, parseBostonCommand } from './bostonCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseBostonCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseBostonCommand', () => {
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
      expect(parseBostonCommand(input)).toEqual({ args: [want] });
    }
  });

  // **段の番号で指す。**ミゼールが間に挟まるのでトリック数では一意にならない。
  it('bids by ladder step, with the suit optional', () => {
    expect(parseBostonCommand('b 4 3')).toEqual({ args: ['bid', { level: 4, suit: 3 }] });
    // ミゼールはスートを取らない。
    expect(parseBostonCommand('bid 3')).toEqual({ args: ['bid', { level: 3 }] });
    expect(parseBostonCommand('b 15')).toEqual({ args: ['bid', { level: 15 }] });
  });

  it('rejects a step outside the ladder or a bad suit', () => {
    for (const input of ['b', 'b abc', 'b 0', 'b 16']) {
      expect(parseError(input)).toContain('ladder step');
    }
    expect(parseError('b 1 9')).toContain('suit is 1-4');
  });

  // **-1 は「単独で戦う」という有効な選択。**
  it('calls a partner or goes alone', () => {
    expect(parseBostonCommand('cp 2')).toEqual({ args: ['callpartner', { partner: 2 }] });
    expect(parseBostonCommand('callpartner -1')).toEqual({ args: ['callpartner', { partner: -1 }] });
    for (const input of ['cp', 'cp abc', 'cp -2', 'cp 4']) {
      expect(parseError(input)).toContain('cp -1 to play alone');
    }
  });

  it('parses a play and rejects an index outside the hand', () => {
    expect(parseBostonCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseBostonCommand('play 12')).toEqual({ args: ['play', { cardIndex: 12 }] });
    for (const input of ['p', 'p abc', 'p -1', 'p 13']) {
      expect(parseError(input)).toContain('Usage: p');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('callpartne')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['b <1-15> [suit]', 'ps / pass', 'cp <seat|-1>', 'p <0-12>']) {
      expect(BOSTON_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // 段の番号であってトリック数ではない、と help でも言う。
    expect(BOSTON_HELP.some((line) => line.includes('LADDER STEP'))).toBe(true);
  });
});
