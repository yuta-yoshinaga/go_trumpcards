import { describe, expect, it } from 'vitest';
import { parseSnapCommand, SNAP_HELP } from './snapCommands';

describe('parseSnapCommand', () => {
  it.each([
    ['s', ['step']],
    ['step', ['step']],
    ['n', ['snap']],
    ['snap', ['snap']],
    ['t', ['tick']],
    ['tick', ['tick']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseSnapCommand(input)).toEqual({ args: expected });
  });

  // **席は取らない。** 席を選べると CPU に誤宣言させられる。
  it('ignores any seat given to snap', () => {
    expect(parseSnapCommand('n 1')).toEqual({ args: ['snap'] });
    expect(parseSnapCommand('snap 2')).toEqual({ args: ['snap'] });
  });

  it('suggests a near miss', () => {
    expect(parseSnapCommand('ste')).toEqual({ error: 'Unknown command: ste. Did you mean: step?' });
  });

  it('reports an unknown command', () => {
    expect(parseSnapCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = SNAP_HELP.join('\n');
    for (const fragment of ['s/step', 'n/snap', 't/tick', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **席を取らないことが読めること。**
    expect(help).toMatch(/no seat argument/);
  });
});
