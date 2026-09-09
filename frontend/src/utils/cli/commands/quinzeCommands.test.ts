import { describe, expect, it } from 'vitest';
import { parseQuinzeCommand, QUINZE_HELP } from './quinzeCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseQuinzeCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseQuinzeCommand', () => {
  it('parses the commands that carry no amount', () => {
    expect(parseQuinzeCommand('deal')).toEqual({ args: ['deal'] });
    expect(parseQuinzeCommand('h')).toEqual({ args: ['hit'] });
    expect(parseQuinzeCommand('hit')).toEqual({ args: ['hit'] });
    expect(parseQuinzeCommand('s')).toEqual({ args: ['stand'] });
    expect(parseQuinzeCommand('stand')).toEqual({ args: ['stand'] });
    expect(parseQuinzeCommand('bh')).toEqual({ args: ['bankerhit'] });
    expect(parseQuinzeCommand('bs')).toEqual({ args: ['bankerstand'] });
    expect(parseQuinzeCommand('log')).toEqual({ args: ['log'] });
    expect(parseQuinzeCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses a bet', () => {
    expect(parseQuinzeCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseQuinzeCommand('bet 100')).toEqual({ args: ['bet', 100] });
  });

  it('rejects a missing or unusable stake', () => {
    for (const bad of ['b', 'b abc', 'b 0', 'b -5']) {
      expect(errorOf(bad)).toContain('Usage: b');
    }
  });

  it('suggests a close command', () => {
    expect(errorOf('stnd')).toContain('Did you mean');
  });

  it('reports an unknown command', () => {
    expect(errorOf('zzzzz')).toContain('Unknown command');
  });

  it('documents every action in the help text', () => {
    for (const prefix of ['b <amount>', 'deal', 'h/hit', 's/stand', 'bh', 'bs']) {
      expect(QUINZE_HELP.some((line) => line.startsWith(prefix))).toBe(true);
    }
  });
});
