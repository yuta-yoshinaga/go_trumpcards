import { describe, expect, it } from 'vitest';
import { HASENPFEFFER_HELP, parseHasenpfefferCommand } from './hasenpfefferCommands';

describe('parseHasenpfefferCommand', () => {
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
    expect(parseHasenpfefferCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 5', ['play', 5]],
  ])('parses %s', (input, args) => {
    expect(parseHasenpfefferCommand(input)).toEqual({ args });
  });

  // **宣言は5番目の引数。** 位置がずれると別の値として届く。
  it.each([
    ['b 3', 3],
    ['bid 6', 6],
  ])('parses %s', (input, bid) => {
    expect(parseHasenpfefferCommand(input)).toEqual({ args: ['bid', undefined, undefined, undefined, bid] });
  });

  // **降りるのは pass 専用。** `b 0` を通すと下限の検査がすり抜ける。
  it('sends a pass as bid 0', () => {
    expect(parseHasenpfefferCommand('pass')).toEqual({ args: ['bid', undefined, undefined, undefined, 0] });
  });

  it.each(['b', 'b x', 'b 0', 'b 2', 'b 7'])('rejects %s', (input) => {
    expect(parseHasenpfefferCommand(input)).toEqual({ error: 'Usage: b <3-6>' });
  });

  // **捨て札は札とスートの2引数。** どちらも既定値で埋めない。
  it('parses a discard with its suit', () => {
    expect(parseHasenpfefferCommand('d 2 3')).toEqual({ args: ['discard', 2, undefined, 3] });
    expect(parseHasenpfefferCommand('discard 0 1')).toEqual({ args: ['discard', 0, undefined, 1] });
  });

  it.each(['d', 'd 2', 'd x 3', 'd 2 x', 'd 2 0', 'd 2 5'])('rejects %s', (input) => {
    expect(parseHasenpfefferCommand(input)).toEqual({ error: 'Usage: d <cardIdx> <suit 1-4>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseHasenpfefferCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near-miss command', () => {
    expect(parseHasenpfefferCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseHasenpfefferCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { b: 'b 3', d: 'd 0 1', p: 'p 0' };
    for (const line of HASENPFEFFER_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseHasenpfefferCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
