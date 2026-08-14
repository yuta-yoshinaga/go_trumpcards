import { describe, expect, it } from 'vitest';
import { parseShelemCommand, SHELEM_HELP } from './shelemCommands';

describe('parseShelemCommand', () => {
  it.each([
    ['shelem', ['shelem']],
    ['pass', ['pass']],
    ['n', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['r', ['reset']],
    ['log', ['log']],
    ['l', ['log']],
  ])('parses %s', (input, args) => {
    expect(parseShelemCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 11', ['play', 11]],
  ])('parses %s', (input, args) => {
    expect(parseShelemCommand(input)).toEqual({ args });
  });

  // **入札額は5番目の引数。** 位置がずれると別の値として届く。
  it.each([
    ['b 55', 55],
    ['bid 100', 100],
    ['b 80', 80],
  ])('parses %s', (input, bid) => {
    expect(parseShelemCommand(input)).toEqual({
      args: ['bid', undefined, undefined, undefined, bid],
    });
  });

  it.each(['b', 'b x', 'b 50', 'b 105', 'b 200'])('rejects %s', (input) => {
    expect(parseShelemCommand(input)).toEqual({ error: 'Usage: b <55-100>' });
  });

  // **捨て札は4つのインデックス + スートで1つ。** どれが欠けても成立しない。
  it('parses a discard with four indices and a suit', () => {
    expect(parseShelemCommand('d 0 3 7 11 3')).toEqual({
      args: ['discard', undefined, undefined, 3, undefined, [0, 3, 7, 11]],
    });
  });

  it.each(['d', 'd 0 1 2 3', 'd 0 x 2 3 1', 'd 0 1 2 3 0', 'd 0 1 2 3 5', 'd 0 1 2 3 x'])('rejects %s', (input) => {
    expect(parseShelemCommand(input)).toEqual({ error: 'Usage: d <i> <i> <i> <i> <suit 1-4>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseShelemCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near-miss command', () => {
    expect(parseShelemCommand('pas')).toEqual({ error: 'Unknown command: pas. Did you mean: pass?' });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseShelemCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが実在するコマンドだけを案内していること。
  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { b: 'b 55', d: 'd 0 1 2 3 1', p: 'p 0' };
    for (const line of SHELEM_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseShelemCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
