import { describe, expect, it } from 'vitest';
import { CRIBBAGESQUARES_HELP, parseCribbageSquaresCommand } from './cribbagesquaresCommands';

describe('parseCribbageSquaresCommand', () => {
  it.each([
    ['r', 'reset'],
    ['reset', 'reset'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['u', 'undo'],
    ['undo', 'undo'],
    ['g', 'giveup'],
    ['giveup', 'giveup'],
    ['log', 'log'],
  ])('maps %s to %s', (input, expected) => {
    expect(parseCribbageSquaresCommand(input)).toEqual({ args: [expected] });
  });

  it('parses a placement', () => {
    expect(parseCribbageSquaresCommand('p 1 2')).toEqual({ args: ['place', 1, 2] });
    expect(parseCribbageSquaresCommand('place 0 0')).toEqual({ args: ['place', 0, 0] });
  });

  // The grid is 4x4, so 4 is off the board -- Poker Squares' 0-4 range would
  // silently accept a fifth row here.
  it.each([['p 4 0'], ['p 0 4'], ['p -1 0'], ['p 0 -1']])('rejects out-of-range %s', (input) => {
    expect(parseCribbageSquaresCommand(input)).toEqual({ error: 'Row and col must be between 0 and 3' });
  });

  it('accepts the last cell on the board', () => {
    expect(parseCribbageSquaresCommand('p 3 3')).toEqual({ args: ['place', 3, 3] });
  });

  it.each(['p', 'p 1', 'p a b'])('reports usage for %s', (input) => {
    expect(parseCribbageSquaresCommand(input)).toEqual({ error: 'Usage: p <row 0-3> <col 0-3>' });
  });

  it('suggests a near-miss command', () => {
    const result = parseCribbageSquaresCommand('plce');
    expect('error' in result && result.error).toContain('place');
  });

  it('reports an unknown command', () => {
    expect(parseCribbageSquaresCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    expect(CRIBBAGESQUARES_HELP.join('\n')).toContain('p <r> <c>');
    expect(CRIBBAGESQUARES_HELP.join('\n')).toContain('hint');
  });
});
