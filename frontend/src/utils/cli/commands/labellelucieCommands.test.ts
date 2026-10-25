import { describe, expect, it } from 'vitest';
import { parseLaBelleLucieCommand } from './labellelucieCommands';

describe('parseLaBelleLucieCommand', () => {
  it('maps every supported command and alias to the page API operations', () => {
    const cases: [string, unknown[]][] = [
      ['p 2 4', ['mf', 2, 4]],
      ['play 2 4', ['mf', 2, 4]],
      ['f 3', ['ff', 3]],
      ['foundation 3', ['ff', 3]],
      ['rd', ['rd']],
      ['redeal', ['rd']],
      ['ac', ['ac']],
      ['autocomplete', ['ac']],
      ['u', ['u']],
      ['undo', ['u']],
      ['undo_n 2', ['undo_n', undefined, undefined, 2]],
      ['g', ['giveup']],
      ['giveup', ['giveup']],
      ['h', ['hint']],
      ['hint', ['hint']],
      ['r', ['reset']],
      ['reset', ['reset']],
    ];
    for (const [input, args] of cases) expect(parseLaBelleLucieCommand(input)).toEqual({ args });
  });

  it('rejects missing arguments, invalid fan indices, and invalid undo counts', () => {
    for (const input of ['p', 'p 2', 'p -1 2', 'p 2 9', 'p x 2'])
      expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
    for (const input of ['f', 'f -1', 'f 9', 'f x']) expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
    for (const input of ['undo_n', 'undo_n 0', 'undo_n x'])
      expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
  });

  it('rejects unknown commands without accepting log', () => {
    expect(parseLaBelleLucieCommand('bogus')).toEqual({ error: 'Unknown command: bogus' });
    expect(parseLaBelleLucieCommand('log')).toHaveProperty('error');
  });
});
