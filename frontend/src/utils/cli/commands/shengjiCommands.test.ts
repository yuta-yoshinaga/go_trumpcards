import { describe, expect, it } from 'vitest';
import { parseShengJiCommand, SHENGJI_HELP } from './shengjiCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseShengJiCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseShengJiCommand', () => {
  it('parses the parameterless actions and their aliases', () => {
    for (const [input, want] of [
      ['n', 'next'],
      ['next', 'next'],
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseShengJiCommand(input)).toEqual({ args: [want] });
    }
  });

  // **0 はパス。**省略と同じ扱いにすると亮牌を降りられなくなる。
  it('declares a suit, and zero passes', () => {
    expect(parseShengJiCommand('d 3')).toEqual({ args: ['declare', { suit: 3 }] });
    expect(parseShengJiCommand('declare 0')).toEqual({ args: ['declare', { suit: 0 }] });
    expect(parseShengJiCommand('d 4')).toEqual({ args: ['declare', { suit: 4 }] });
  });

  it('rejects a bad declaration', () => {
    for (const input of ['d', 'd abc', 'd -1', 'd 5']) {
      expect(parseError(input)).toContain('d <0-4>');
    }
  });

  // **底牌はちょうど 8 枚。**
  it('buries exactly eight cards', () => {
    expect(parseShengJiCommand('b 0 1 2 3 4 5 6 7')).toEqual({
      args: ['bury', { cardIndexes: [0, 1, 2, 3, 4, 5, 6, 7] }],
    });
    expect(parseError('b 0 1')).toContain('exactly 8');
    expect(parseError('b')).toContain('card indexes are required');
  });

  it('plays any number of cards', () => {
    expect(parseShengJiCommand('p 3')).toEqual({ args: ['play', { cardIndexes: [3] }] });
    expect(parseShengJiCommand('play 0 1')).toEqual({ args: ['play', { cardIndexes: [0, 1] }] });
  });

  it('rejects a bad play', () => {
    expect(parseError('p')).toContain('card indexes are required');
    for (const input of ['p abc', 'p -1', 'p 99']) {
      expect(parseError(input)).toContain('every index is 0-32');
    }
  });

  // **同じ札を 2 回数えられない。**通すと 1 枚から対子が作れてしまう。
  it('rejects the same index given twice', () => {
    expect(parseError('p 1 1')).toContain('index 1 was given twice');
    expect(parseError('b 0 0 1 2 3 4 5 6')).toContain('index 0 was given twice');
  });

  it('suggests a near miss and rejects the rest', () => {
    expect(parseError('declar')).toContain('Did you mean');
    expect(parseError('zzz')).toBe('Unknown command: zzz');
  });

  it('documents every action it accepts', () => {
    const help = SHENGJI_HELP.join('\n');
    for (const fragment of ['d <0-4>', 'b <idx x8>', 'p <idx...>', 'n / next', 'r / reset']) {
      expect(help).toContain(fragment);
    }
    // パスできることがヘルプから読めること。
    expect(help).toContain('0 passes');
  });
});
