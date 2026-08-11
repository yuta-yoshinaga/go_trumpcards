import { describe, expect, it } from 'vitest';
import { BHABHI_HELP, parseBhabhiCommand } from './bhabhiCommands';

describe('parseBhabhiCommand', () => {
  it.each([
    ['h', ['hint']],
    ['hint', ['hint']],
    ['g', ['giveup']],
    ['giveup', ['giveup']],
    ['r', ['reset']],
    ['reset', ['reset']],
    ['log', ['log']],
    ['l', ['log']],
  ])('parses %s', (input, args) => {
    expect(parseBhabhiCommand(input)).toEqual({ args });
  });

  it.each([
    ['p 0', ['play', 0]],
    ['play 12', ['play', 12]],
  ])('parses %s', (input, args) => {
    expect(parseBhabhiCommand(input)).toEqual({ args });
  });

  it.each(['p', 'play abc'])('rejects %s', (input) => {
    expect(parseBhabhiCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **次のハンドへ進むコマンドは無い。** 配り切りの 1 ゲームで終わるので、
  // 受け付けるとありもしない区切りを案内することになる。
  it.each(['n', 'next'])('has no next command (%s)', (input) => {
    expect(parseBhabhiCommand(input)).toHaveProperty('error');
    expect(parseBhabhiCommand(input)).not.toHaveProperty('args');
  });

  it('suggests a near-miss command', () => {
    expect(parseBhabhiCommand('pla')).toEqual({ error: 'Unknown command: pla. Did you mean: play?' });
  });

  it('reports an unknown command with no near match', () => {
    expect(parseBhabhiCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが実在するコマンドだけを案内していること。
  it('documents only commands the parser accepts', () => {
    const sample: Record<string, string> = { p: 'p 0' };
    for (const line of BHABHI_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseBhabhiCommand(sample[cmd] ?? cmd)).not.toHaveProperty('error', `Unknown command: ${cmd}`);
    }
  });
});
