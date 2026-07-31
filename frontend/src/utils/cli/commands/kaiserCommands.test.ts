import { describe, expect, it } from 'vitest';
import { KAISER_HELP, parseKaiserCommand } from './kaiserCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseKaiserCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseKaiserCommand', () => {
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
      expect(parseKaiserCommand(input)).toEqual({ args: [want] });
    }
  });

  // **契約は省略できる。**省略時は切札あり。
  it('parses a bid with and without a contract', () => {
    expect(parseKaiserCommand('b 8')).toEqual({ args: ['bid', { bid: 8, contract: 0 }] });
    expect(parseKaiserCommand('bid 12 2')).toEqual({ args: ['bid', { bid: 12, contract: 2 }] });
  });

  // **最低は 7。**キティを見られる利があるので 6 では始まらない。
  it('rejects a bid outside 7-12', () => {
    for (const input of ['b', 'b abc', 'b 6', 'b 13']) {
      expect(parseError(input)).toContain('Usage: b');
    }
    expect(parseError('b 8 9')).toContain('contract is 0');
  });

  it('parses a trump call and rejects a bad suit', () => {
    expect(parseKaiserCommand('t 3')).toEqual({ args: ['trump', { suit: 3 }] });
    for (const input of ['t', 't abc', 't 0', 't 5']) {
      expect(parseError(input)).toContain('Usage: t');
    }
  });

  // **捨て札は必ず 2 枚。**キティと同数でないと手札が合わなくなる。
  it('takes exactly two discard indices', () => {
    expect(parseKaiserCommand('d 0 3')).toEqual({ args: ['discard', { indices: [0, 3] }] });
    for (const input of ['d', 'd 0', 'd a b', 'd -1 2']) {
      expect(parseError(input)).toContain('exactly 2 cards');
    }
  });

  it('parses a play and rejects a bad index', () => {
    expect(parseKaiserCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    for (const input of ['p', 'p abc', 'p -1']) {
      expect(parseError(input)).toContain('Usage: p');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('discar')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['b <7-12> [0-2]', 'ps / pass', 't <1-4>', 'd <i> <j>', 'p <i>']) {
      expect(KAISER_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // 点数であってトリック数ではない、と help でも言う。
    expect(KAISER_HELP.some((line) => line.includes('POINTS'))).toBe(true);
  });
});
