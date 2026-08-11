import { describe, expect, it } from 'vitest';
import { parseSergeantMajorCommand, SERGEANTMAJOR_HELP } from './sergeantmajorCommands';

describe('parseSergeantMajorCommand', () => {
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
    expect(parseSergeantMajorCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 15', ['play', 15]],
  ])('parses %s', (input, args) => {
    expect(parseSergeantMajorCommand(input)).toEqual({ args });
  });

  // **切り札は4番目の引数。** 位置がずれると別の値として届く。
  it.each([
    ['t 1', 1],
    ['trump 4', 4],
  ])('parses %s', (input, suit) => {
    expect(parseSergeantMajorCommand(input)).toEqual({ args: ['trump', undefined, undefined, suit] });
  });

  it.each(['t', 't x', 't 0', 't 5'])('rejects %s', (input) => {
    expect(parseSergeantMajorCommand(input)).toEqual({ error: 'Usage: t <suit 1-4>' });
  });

  // **捨て札は 4 枚ちょうど。** 既定値で埋めない。
  it('parses a four-card discard', () => {
    expect(parseSergeantMajorCommand('d 0 2 5 7')).toEqual({
      args: ['discard', undefined, undefined, undefined, [0, 2, 5, 7]],
    });
  });

  it.each(['d', 'd 0', 'd 0 1', 'd 0 1 2', 'd 0 x 2 3'])('rejects %s', (input) => {
    expect(parseSergeantMajorCommand(input)).toEqual({ error: 'Usage: d <i> <i> <i> <i>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseSergeantMajorCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **ノルマを宣言するコマンドは無い。** 8・5・3 は席順で決まる。
  it.each(['b 8', 'bid 5'])('has no bid command (%s)', (input) => {
    expect(parseSergeantMajorCommand(input)).toHaveProperty('error');
    expect(parseSergeantMajorCommand(input)).not.toHaveProperty('args');
  });

  it('suggests a near-miss command', () => {
    expect(parseSergeantMajorCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseSergeantMajorCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { t: 't 1', d: 'd 0 1 2 3', p: 'p 0' };
    for (const line of SERGEANTMAJOR_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseSergeantMajorCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
