import { describe, expect, it } from 'vitest';
import { parseTeenDoPaanchCommand, TEENDOPAANCH_HELP } from './teendopaanchCommands';

describe('parseTeenDoPaanchCommand', () => {
  it.each([
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['g', ['giveup']],
    ['giveup', ['giveup']],
    ['r', ['reset']],
    ['reset', ['reset']],
    ['log', ['log']],
    ['l', ['log']],
  ])('parses %s', (input, args) => {
    expect(parseTeenDoPaanchCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 9', ['play', 9]],
  ])('parses %s', (input, args) => {
    expect(parseTeenDoPaanchCommand(input)).toEqual({ args });
  });

  // **切り札は4番目の引数。** 位置がずれると別の値として届く。
  it.each([
    ['t 1', 1],
    ['trump 4', 4],
  ])('parses %s', (input, suit) => {
    expect(parseTeenDoPaanchCommand(input)).toEqual({ args: ['trump', undefined, undefined, suit] });
  });

  // **スート無し／範囲外は通さない。** 通すと選んでいないスートが切り札になる。
  it.each(['t', 't x', 't 0', 't 5'])('rejects %s', (input) => {
    expect(parseTeenDoPaanchCommand(input)).toEqual({ error: 'Usage: t <suit 1-4>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseTeenDoPaanchCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **ノルマを宣言するコマンドは無い。** 3・2・5 は割り当てで選ぶ余地が無い。
  it.each(['b 3', 'bid 5'])('has no bid command (%s)', (input) => {
    expect(parseTeenDoPaanchCommand(input)).toHaveProperty('error');
    expect(parseTeenDoPaanchCommand(input)).not.toHaveProperty('args');
  });

  it('suggests a near-miss command', () => {
    expect(parseTeenDoPaanchCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseTeenDoPaanchCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { t: 't 1', p: 'p 0' };
    for (const line of TEENDOPAANCH_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseTeenDoPaanchCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
