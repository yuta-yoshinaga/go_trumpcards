import { describe, expect, it } from 'vitest';
import { ISRAELIWHIST_HELP, parseIsraeliWhistCommand } from './israeliwhistCommands';

describe('parseIsraeliWhistCommand', () => {
  it.each([
    ['pass', ['pass']],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['r', ['reset']],
    ['log', ['log']],
    ['l', ['log']],
  ])('parses %s', (input, args) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 12', ['play', 12]],
  ])('parses %s', (input, args) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({ args });
  });

  // **入札は数が5番目、スートが4番目の引数。** 位置がずれると別の入札になる。
  it.each([
    ['a 5 1', 5, 1],
    ['auction 9 2', 9, 2],
    ['a 13 4', 13, 4],
  ])('parses %s', (input, bid, suit) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({
      args: ['auction', undefined, undefined, suit, bid],
    });
  });

  // **数とスートのどちらが欠けても成立しない。** 最低入札も範囲も見る。
  it.each(['a', 'a 7', 'a x 1', 'a 7 x', 'a 4 1', 'a 14 1', 'a 7 0', 'a 7 5'])('rejects %s', (input) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({ error: 'Usage: a <n 5-13> <suit 1-4>' });
  });

  // **宣言は 0 から。** オークションと違って下限は 0。
  it.each([
    ['b 0', 0],
    ['bid 13', 13],
  ])('parses %s', (input, bid) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({
      args: ['bid', undefined, undefined, undefined, bid],
    });
  });

  it.each(['b', 'b x', 'b -1', 'b 14'])('rejects %s', (input) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({ error: 'Usage: b <0-13>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseIsraeliWhistCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near-miss command', () => {
    expect(parseIsraeliWhistCommand('pas')).toEqual({
      error: 'Unknown command: pas. Did you mean: pass?',
    });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseIsraeliWhistCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが実在するコマンドだけを案内していること。
  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { a: 'a 5 1', b: 'b 0', p: 'p 0' };
    for (const line of ISRAELIWHIST_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseIsraeliWhistCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
