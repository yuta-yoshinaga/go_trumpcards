import { describe, expect, it } from 'vitest';
import { parseTrucoCommand } from './trucoCommands';

describe('parseTrucoCommand', () => {
  it.each([
    ['p 0', ['play', 0]],
    ['play 2', ['play', 2]],
    ['t', ['truco']],
    ['truco', ['truco']],
    ['a', ['accept']],
    ['accept', ['accept']],
    ['d', ['decline']],
    ['decline', ['decline']],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ])('maps %s to the page action arguments', (input, args) => {
    expect(parseTrucoCommand(input)).toEqual({ args });
  });
  it.each(['p', 'play x', 'p -1', 'p 3', 'p 1.5', 'p 01'])('rejects invalid card index: %s', (input) => {
    expect(parseTrucoCommand(input)).toEqual({ error: 'Usage: p <0-2>' });
  });
  it('rejects unknown commands', () => {
    expect(parseTrucoCommand('bid 1')).toMatchObject({ error: expect.stringContaining('Unknown command: bid') });
  });
  it.each(['l', 'log'])('treats %s as an unknown command', (input) => {
    expect(parseTrucoCommand(input)).toMatchObject({ error: expect.stringContaining(`Unknown command: ${input}`) });
  });
});
