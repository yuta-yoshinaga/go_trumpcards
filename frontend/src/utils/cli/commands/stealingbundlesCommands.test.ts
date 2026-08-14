import { describe, expect, it } from 'vitest';
import { parseStealingBundlesCommand, STEALINGBUNDLES_HELP } from './stealingbundlesCommands';

describe('parseStealingBundlesCommand', () => {
  it.each([
    ['t 2', ['take', 2]],
    ['take 0', ['take', 0]],
    ['d 1', ['trail', 1]],
    ['trail 1', ['trail', 1]],
    // **略奪は札と相手の 2 つを取ります。**
    ['s 1 3', ['steal', 1, 3]],
    ['steal 0 2', ['steal', 0, 2]],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseStealingBundlesCommand(input)).toEqual({ args: expected });
  });

  it.each(['t', 't x'])('rejects take %s', (input) => {
    expect(parseStealingBundlesCommand(input)).toEqual({ error: 'Usage: t <cardIdx>' });
  });

  it.each(['d', 'd x'])('rejects trail %s', (input) => {
    expect(parseStealingBundlesCommand(input)).toEqual({ error: 'Usage: d <cardIdx>' });
  });

  // **相手を書かないと誰の束か決まりません。**
  it.each(['s', 's 1', 's x 1', 's 1 y'])('rejects steal %s', (input) => {
    expect(parseStealingBundlesCommand(input)).toEqual({ error: 'Usage: s <cardIdx> <victimIdx>' });
  });

  it('suggests a near miss', () => {
    expect(parseStealingBundlesCommand('stea')).toEqual({ error: 'Unknown command: stea. Did you mean: steal?' });
  });

  it('reports an unknown command', () => {
    expect(parseStealingBundlesCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = STEALINGBUNDLES_HELP.join('\n');
    for (const fragment of ['t <cardIdx>', 's <cardIdx> <victimIdx>', 'd <cardIdx>', 'h/hint', 'g/giveup', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **置けるのは取れないときだけ、が読めること。**
    expect(help).toMatch(/only when nothing can be captured/);
  });
});
