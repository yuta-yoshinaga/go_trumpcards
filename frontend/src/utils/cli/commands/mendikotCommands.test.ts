import { describe, expect, it } from 'vitest';
import { MENDIKOT_HELP, parseMendikotCommand } from './mendikotCommands';

describe('parseMendikotCommand', () => {
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
    expect(parseMendikotCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 12', ['play', 12]],
  ])('parses %s', (input, args) => {
    expect(parseMendikotCommand(input)).toEqual({ args });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseMendikotCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **切り札を選ぶコマンドは無い。** 受け付けると、フォローできない札で決まる
  // という規則と二重になる。
  it.each(['t 1', 'trump 3'])('has no trump command (%s)', (input) => {
    expect(parseMendikotCommand(input)).toHaveProperty('error');
    expect(parseMendikotCommand(input)).not.toHaveProperty('args');
  });

  it('suggests a near-miss command', () => {
    expect(parseMendikotCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseMendikotCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが実在するコマンドだけを案内していること。
  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { p: 'p 0' };
    for (const line of MENDIKOT_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseMendikotCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
