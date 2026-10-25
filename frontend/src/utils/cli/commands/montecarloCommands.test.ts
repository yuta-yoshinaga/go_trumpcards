import { describe, expect, it } from 'vitest';
import { parseMonteCarloCommand } from './montecarloCommands';

describe('parseMonteCarloCommand', () => {
  it('maps each remove syntax to the page button exec coordinates', () => {
    expect(parseMonteCarloCommand('p 0 1')).toEqual({ args: ['remove', 0, 0, 0, 1] });
    expect(parseMonteCarloCommand('play 1 7')).toEqual({ args: ['remove', 0, 1, 1, 2] });
    expect(parseMonteCarloCommand('remove 24 0')).toEqual({ args: ['remove', 4, 4, 0, 0] });
  });

  it('maps every other API command to the same values as its page button', () => {
    expect(parseMonteCarloCommand('d')).toEqual({ args: ['deal'] });
    expect(parseMonteCarloCommand('deal')).toEqual({ args: ['deal'] });
    expect(parseMonteCarloCommand('u')).toEqual({ args: ['undo'] });
    expect(parseMonteCarloCommand('undo')).toEqual({ args: ['undo'] });
    expect(parseMonteCarloCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseMonteCarloCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseMonteCarloCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMonteCarloCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseMonteCarloCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMonteCarloCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing, invalid and out-of-range cell indices', () => {
    for (const input of ['p', 'p 1', 'p x 2', 'p 1 x', 'p -1 2', 'p 25 1', 'p 4 4']) {
      expect(parseMonteCarloCommand(input)).toHaveProperty('error');
    }
  });

  it('rejects unknown commands and does not expose action log', () => {
    expect(parseMonteCarloCommand('log')).toHaveProperty('error');
    expect(parseMonteCarloCommand('xyz')).toHaveProperty('error');
  });
});
