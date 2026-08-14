import { describe, expect, it } from 'vitest';
import { HONEYMOONBRIDGE_HELP, parseHoneymoonBridgeCommand } from './honeymoonbridgeCommands';

describe('parseHoneymoonBridgeCommand', () => {
  it.each([
    ['b 3 0', ['bid', undefined, undefined, 3, 0]],
    ['bid 1 4', ['bid', undefined, undefined, 1, 4]],
    ['b 7 0', ['bid', undefined, undefined, 7, 0]],
    ['pass', ['pass']],
    ['p 2', ['play', 2]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseHoneymoonBridgeCommand(input)).toEqual({ args: expected });
  });

  // **スートを既定値で埋めない。** 0 はノートランプという明示の選択肢なので、
  // 省略を 0 に丸めると宣言していない契約を落札してしまう。
  it.each(['b', 'b 3', 'b x 0', 'b 3 x', 'b 0 0', 'b 8 0', 'b 3 -1', 'b 3 5'])('rejects %s', (input) => {
    expect(parseHoneymoonBridgeCommand(input)).toEqual({ error: 'Usage: b <level 1-7> <suit 0-4>' });
  });

  it('rejects play without an index', () => {
    expect(parseHoneymoonBridgeCommand('p')).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near miss', () => {
    expect(parseHoneymoonBridgeCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command', () => {
    expect(parseHoneymoonBridgeCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = HONEYMOONBRIDGE_HELP.join('\n');
    for (const fragment of ['b <lvl> <s>', 'pass', 'p <cardIdx>', 'n/next', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **ノートランプが 0 だと分かること。** 綴りを覚える手がかりが要る。
    expect(help).toMatch(/0=NT/);
  });
});
