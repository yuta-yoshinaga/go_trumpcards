import { describe, expect, it } from 'vitest';
import { parseTongitsCommand } from './tongitsCommands';

describe('parseTongitsCommand', () => {
  it.each([
    ['ds', 'drawstock'],
    ['drawstock', 'drawstock'],
    ['dd', 'drawdiscard'],
    ['drawdiscard', 'drawdiscard'],
    ['nr', 'nextround'],
    ['nextround', 'nextround'],
    ['log', 'log'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('parses %s as %s', (input, command) => {
    expect(parseTongitsCommand(input)).toEqual({ args: [command] });
  });

  it.each(['d', 'dis', 'discard'])('parses %s with a card index', (command) => {
    expect(parseTongitsCommand(`${command} 4`)).toEqual({ args: ['discard', 4] });
  });

  it.each(['d', 'dis', 'discard'])('returns an error when %s has no index', (command) => {
    expect(parseTongitsCommand(command)).toEqual({ error: 'Usage: d <idx>' });
  });

  it('returns an error when discard index is not numeric', () => {
    expect(parseTongitsCommand('discard nope')).toEqual({ error: 'Usage: d <idx>' });
  });

  it.each(['m', 'meld'])('parses %s with three or more indices', (command) => {
    expect(parseTongitsCommand(`${command} 0 2 5 7`)).toEqual({
      args: ['meld', undefined, undefined, [0, 2, 5, 7]],
    });
  });

  it('returns an error when meld has fewer than three indices', () => {
    expect(parseTongitsCommand('m 0 1')).toEqual({ error: 'Usage: m <idx> [idx ...]' });
  });

  it('returns an error when a meld index is not numeric', () => {
    expect(parseTongitsCommand('meld 0 nope 2')).toEqual({ error: 'Usage: m <idx> [idx ...]' });
  });

  it.each(['sp', 'sapaw'])('parses %s and reorders its arguments for the API', (command) => {
    expect(parseTongitsCommand(`${command} 1 2 7`)).toEqual({
      args: ['sapaw', 7, undefined, undefined, 1, 2],
    });
  });

  it('returns an error when sapaw does not have exactly three arguments', () => {
    expect(parseTongitsCommand('sp 1 2')).toEqual({ error: 'Usage: sp <player> <meld> <card>' });
    expect(parseTongitsCommand('sapaw 1 2 7 8')).toEqual({ error: 'Usage: sp <player> <meld> <card>' });
  });

  it('returns an error when a sapaw argument is not numeric', () => {
    expect(parseTongitsCommand('sapaw 1 nope 7')).toEqual({ error: 'Usage: sp <player> <meld> <card>' });
  });

  it.each(['c', 'challenge'])('adds the challenge flags for %s', (command) => {
    expect(parseTongitsCommand(command)).toEqual({
      args: ['challenge', undefined, undefined, undefined, undefined, undefined, [true, true]],
    });
  });

  it('suggests a close command typo', () => {
    expect(parseTongitsCommand('drawstoc')).toEqual({
      error: 'Unknown command: drawstoc. Did you mean: drawstock?',
    });
  });

  it('reports an unknown command without a suggestion when it is not close', () => {
    expect(parseTongitsCommand('xyz')).toEqual({ error: 'Unknown command: xyz' });
  });
});
