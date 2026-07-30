import { describe, expect, it } from 'vitest';
import { parseToepenCommand, TOEPEN_HELP } from './toepenCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseToepenCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseToepenCommand', () => {
  it('parses play with a hand index', () => {
    expect(parseToepenCommand('p 2')).toEqual({ args: ['play', 2] });
  });

  it('maps stay and fold onto the same command with opposite answers', () => {
    // Separate words rather than `answer true|false`: a mistyped boolean would
    // silently invert the decision, which costs lives.
    expect(parseToepenCommand('s')).toEqual({ args: ['answer', undefined, true] });
    expect(parseToepenCommand('stay')).toEqual({ args: ['answer', undefined, true] });
    expect(parseToepenCommand('f')).toEqual({ args: ['answer', undefined, false] });
    expect(parseToepenCommand('fold')).toEqual({ args: ['answer', undefined, false] });
  });

  it('parses toep, next, log and reset', () => {
    expect(parseToepenCommand('t')).toEqual({ args: ['toep'] });
    expect(parseToepenCommand('toep')).toEqual({ args: ['toep'] });
    expect(parseToepenCommand('n')).toEqual({ args: ['next'] });
    expect(parseToepenCommand('log')).toEqual({ args: ['log'] });
    expect(parseToepenCommand('r')).toEqual({ args: ['reset'] });
  });

  it('rejects a missing or malformed index', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('p x')).toBe('Invalid card index: x');
    expect(errorOf('p -1')).toBe('Invalid card index: -1');
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('tope')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = TOEPEN_HELP.join('\n');
    for (const cmd of ['p <i>', 't/toep', 's/stay', 'f/fold', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});

describe('parseToepenCommand redeal', () => {
  it('parses redeal with its alias', () => {
    expect(parseToepenCommand('d')).toEqual({ args: ['redeal'] });
    expect(parseToepenCommand('redeal')).toEqual({ args: ['redeal'] });
  });

  it('documents it', () => {
    expect(TOEPEN_HELP.join('\n')).toContain('d/redeal');
  });
});
