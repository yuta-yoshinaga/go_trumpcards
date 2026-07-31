import { describe, expect, it } from 'vitest';
import { BIDEUCHRE_HELP, parseBidEuchreCommand } from './bideuchreCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseBidEuchreCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseBidEuchreCommand', () => {
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
      expect(parseBidEuchreCommand(input)).toEqual({ args: [want] });
    }
  });

  // **最低ビッドは 3。**2 以下は入口で弾く。
  it('bids a trick count from three up', () => {
    expect(parseBidEuchreCommand('b 3')).toEqual({ args: ['bid', { value: 3 }] });
    expect(parseBidEuchreCommand('bid 6')).toEqual({ args: ['bid', { value: 6 }] });
    for (const input of ['b', 'b abc', 'b 0', 'b 2', 'b 7']) {
      expect(parseError(input)).toContain('Usage: b <3-6>');
    }
  });

  // **ノートランプが 2 種類ある。**ローは序列が逆転する。
  it('names any of the six declarations', () => {
    for (const n of [0, 1, 2, 3, 4, 5]) {
      expect(parseBidEuchreCommand(`t ${n}`)).toEqual({ args: ['trump', { trump: n }] });
    }
    expect(parseBidEuchreCommand('trump 5')).toEqual({ args: ['trump', { trump: 5 }] });
    for (const input of ['t', 't abc', 't -1', 't 6']) {
      expect(parseError(input)).toContain('Usage: t <0-5>');
    }
  });

  // **手札は 6 枚。**
  it('parses a play and rejects an index outside the hand', () => {
    expect(parseBidEuchreCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseBidEuchreCommand('play 5')).toEqual({ args: ['play', { cardIndex: 5 }] });
    for (const input of ['p', 'p abc', 'p -1', 'p 6']) {
      expect(parseError(input)).toContain('Usage: p');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('pas')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['b <3-6>', 'ps / pass', 't <0-5>', 'p <0-5>']) {
      expect(BIDEUCHRE_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // 親の同額奪取と NT ローの逆転を help でも言う。
    expect(BIDEUCHRE_HELP.some((line) => line.includes('DEALER may EQUAL'))).toBe(true);
    expect(BIDEUCHRE_HELP.some((line) => line.includes('NT LOW, nine highest'))).toBe(true);
  });
});
