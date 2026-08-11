import { describe, expect, it } from 'vitest';
import { MINIBRIDGE_HELP, parseMinibridgeCommand } from './minibridgeCommands';

describe('parseMinibridgeCommand', () => {
  it.each([
    ['c 3 0', ['contract', undefined, undefined, 3, 0]],
    ['contract 1 4', ['contract', undefined, undefined, 1, 4]],
    ['c 7 0', ['contract', undefined, undefined, 7, 0]],
    ['p 2', ['play', 2]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseMinibridgeCommand(input)).toEqual({ args: expected });
  });

  // **スートを既定値で埋めない。** 0 はノートランプという明示の選択肢。
  it.each(['c', 'c 3', 'c x 0', 'c 3 x', 'c 0 0', 'c 8 0', 'c 3 -1', 'c 3 5'])('rejects %s', (input) => {
    expect(parseMinibridgeCommand(input)).toEqual({ error: 'Usage: c <level 1-7> <suit 0-4>' });
  });

  it('rejects play without an index', () => {
    expect(parseMinibridgeCommand('p')).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **競りは無いので bid コマンドも無い。**
  it('has no bid command', () => {
    expect(parseMinibridgeCommand('bid 3')).toEqual({ error: 'Unknown command: bid' });
  });

  it('suggests a near miss', () => {
    expect(parseMinibridgeCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command', () => {
    expect(parseMinibridgeCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = MINIBRIDGE_HELP.join('\n');
    for (const fragment of ['c <lvl> <s>', 'p <cardIdx>', 'n/next', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    expect(help).toMatch(/0=NT/);
    // **ダミーも自分で出すことが読めること。**
    expect(help).toMatch(/dummy/i);
  });
});
