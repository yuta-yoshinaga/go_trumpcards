import { describe, expect, it } from 'vitest';
import { LITERATURE_HELP, parseLiteratureCommand } from './literatureCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseLiteratureCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseLiteratureCommand', () => {
  it('parses the parameterless actions and their aliases', () => {
    for (const [input, want] of [
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseLiteratureCommand(input)).toEqual({ args: [want] });
    }
  });

  // **要求は 相手・スート・ランク の 3 つ。**札はスートとランクの両方で決まる。
  it('asks for a specific card of a specific seat', () => {
    expect(parseLiteratureCommand('a 1 1 2')).toEqual({ args: ['ask', { target: 1, suit: 1, value: 2 }] });
    expect(parseLiteratureCommand('ask 5 4 13')).toEqual({ args: ['ask', { target: 5, suit: 4, value: 13 }] });
  });

  it('rejects a bad ask', () => {
    for (const input of ['a', 'a 1', 'a 1 1']) {
      expect(parseError(input)).toContain('Usage: a <seat>');
    }
    for (const input of ['a 6 1 2', 'a -1 1 2', 'a abc 1 2']) {
      expect(parseError(input)).toContain('the seat is 0-5');
    }
    for (const input of ['a 1 0 2', 'a 1 5 2', 'a 1 abc 2']) {
      expect(parseError(input)).toContain('the suit is 1-4');
    }
    for (const input of ['a 1 1 0', 'a 1 1 14', 'a 1 1 abc']) {
      expect(parseError(input)).toContain('the rank is 1-13');
    }
  });

  // **宣言は 6 枚すべての所在を申告する。**1 つでも欠けたら宣言にならない。
  it('claims a half-suit with all six placements', () => {
    expect(parseLiteratureCommand('c 0 0 0 2 2 4 4')).toEqual({
      args: ['claim', { halfSuit: 0, holders: [0, 0, 2, 2, 4, 4] }],
    });
    expect(parseLiteratureCommand('claim 7 1 1 1 3 3 5')).toEqual({
      args: ['claim', { halfSuit: 7, holders: [1, 1, 1, 3, 3, 5] }],
    });
  });

  it('rejects a bad claim', () => {
    for (const input of ['c', 'c 0', 'c 0 0 0 2']) {
      expect(parseError(input)).toContain('all 6 holders');
    }
    for (const input of ['c 8 0 0 2 2 4 4', 'c -1 0 0 2 2 4 4', 'c abc 0 0 2 2 4 4']) {
      expect(parseError(input)).toContain('the half-suit is 0-7');
    }
    for (const input of ['c 0 0 0 2 2 4 6', 'c 0 0 0 2 2 4 abc']) {
      expect(parseError(input)).toContain('every holder is a seat 0-5');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('as')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['a <seat> <1-4> <1-13>', 'c <half> <seat x6>', 'l / log', 'r / reset']) {
      expect(LITERATURE_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
    // **相手にのみ・持っている組・持っていない札**を help でも言う。
    expect(LITERATURE_HELP.some((line) => line.includes('OPPONENT'))).toBe(true);
    expect(LITERATURE_HELP.some((line) => line.includes('not that card'))).toBe(true);
    expect(LITERATURE_HELP.some((line) => line.includes('all six cards'))).toBe(true);
  });
});
