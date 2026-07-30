import { describe, expect, it } from 'vitest';
import { parseSkitgubbeCommand, SKITGUBBE_HELP } from './skitgubbeCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseSkitgubbeCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseSkitgubbeCommand', () => {
  it('parses play with a hand index', () => {
    expect(parseSkitgubbeCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseSkitgubbeCommand('play 0')).toEqual({ args: ['play', 0] });
  });

  it('parses pickup with no index at all', () => {
    // The server refuses the pick-up whenever anything still beats the pile,
    // so an index here would be a choice the player does not have.
    expect(parseSkitgubbeCommand('u')).toEqual({ args: ['pickup'] });
    expect(parseSkitgubbeCommand('pickup')).toEqual({ args: ['pickup'] });
    expect(parseSkitgubbeCommand('u 3')).toEqual({ args: ['pickup'] });
  });

  it('rejects a missing or malformed index', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('p x')).toBe('Invalid card index: x');
    expect(errorOf('play -1')).toBe('Invalid card index: -1');
  });

  it('parses log and reset', () => {
    expect(parseSkitgubbeCommand('log')).toEqual({ args: ['log'] });
    expect(parseSkitgubbeCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSkitgubbeCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('pickp')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = SKITGUBBE_HELP.join('\n');
    for (const cmd of ['p <i>', 'u/pickup', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
