import { describe, expect, it } from 'vitest';
import { BALOOT_HELP, parseBalootCommand } from './balootCommands';

describe('parseBalootCommand', () => {
  it.each([
    ['sun', ['sun']],
    ['pass', ['pass']],
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
    expect(parseBalootCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 7', ['play', 7]],
  ])('parses %s', (input, args) => {
    expect(parseBalootCommand(input)).toEqual({ args });
  });

  // **hokom はスートを引数の 4 番目に置く。** cardIndex/config を飛ばすので
  // 位置がずれると切り札が黙って変わる。
  it.each([
    ['hokom 1', 1],
    ['hokom 2', 2],
    ['hokom 3', 3],
    ['hokom 4', 4],
  ])('parses %s', (input, suit) => {
    expect(parseBalootCommand(input)).toEqual({ args: ['hokom', undefined, undefined, suit] });
  });

  // **範囲外のスートを通さない。** 0 や 5 は切り札の無い Hokom になる。
  it.each(['hokom', 'hokom x', 'hokom 0', 'hokom 5'])('rejects %s', (input) => {
    expect(parseBalootCommand(input)).toEqual({ error: 'Usage: hokom <suit 1-4>' });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseBalootCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near-miss command', () => {
    expect(parseBalootCommand('pas')).toEqual({
      error: 'Unknown command: pas. Did you mean: pass?',
    });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseBalootCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが実在するコマンドだけを案内していること。
  it('documents only commands the parser accepts', () => {
    for (const line of BALOOT_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseBalootCommand(cmd === 'hokom' ? 'hokom 1' : cmd === 'p' ? 'p 0' : cmd)).not.toHaveProperty(
        'error',
        `Unknown command: ${cmd}`,
      );
    }
  });
});
