import { describe, expect, it } from 'vitest';
import { ESTIMATION_HELP, parseEstimationCommand } from './estimationCommands';

describe('parseEstimationCommand', () => {
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
    expect(parseEstimationCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 12', ['play', 12]],
  ])('parses %s', (input, args) => {
    expect(parseEstimationCommand(input)).toEqual({ args });
  });

  // **切り札は4番目、宣言は5番目の引数。** 位置がずれると別の値として届く。
  it.each([
    ['t 1', 1],
    ['trump 4', 4],
  ])('parses %s as a trump suit', (input, suit) => {
    expect(parseEstimationCommand(input)).toEqual({ args: ['trump', undefined, undefined, suit] });
  });

  it.each([
    ['b 0', 0],
    ['bid 13', 13],
    ['b 5', 5],
  ])('parses %s as a call', (input, bid) => {
    expect(parseEstimationCommand(input)).toEqual({ args: ['bid', undefined, undefined, undefined, bid] });
  });

  it.each(['t', 't x', 't 0', 't 5'])('rejects %s', (input) => {
    expect(parseEstimationCommand(input)).toEqual({ error: 'Usage: t <suit 1-4>' });
  });

  // **0 は合法な宣言。** 下限で弾いてはいけない（上のケースが通ることで担保）。
  it.each(['b', 'b x', 'b -1', 'b 14'])('rejects %s', (input) => {
    expect(parseEstimationCommand(input)).toEqual({ error: 'Usage: b <0-13>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseEstimationCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near-miss command', () => {
    expect(parseEstimationCommand('nex')).toEqual({
      error: 'Unknown command: nex. Did you mean: next?',
    });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseEstimationCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが実在するコマンドだけを案内していること。
  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { t: 't 1', b: 'b 0', p: 'p 0' };
    for (const line of ESTIMATION_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseEstimationCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
