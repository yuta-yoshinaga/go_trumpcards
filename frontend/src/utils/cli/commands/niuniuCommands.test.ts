import { describe, expect, it } from 'vitest';
import { NIUNIU_HELP, parseNiuNiuCommand } from './niuniuCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseNiuNiuCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseNiuNiuCommand', () => {
  it('parses the bet and its aliases', () => {
    expect(parseNiuNiuCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseNiuNiuCommand('bet 100')).toEqual({ args: ['bet', 100] });
  });

  it('parses reset and log', () => {
    expect(parseNiuNiuCommand('r')).toEqual({ args: ['reset'] });
    expect(parseNiuNiuCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseNiuNiuCommand('log')).toEqual({ args: ['log'] });
  });

  it('rejects a missing or unusable stake', () => {
    for (const bad of ['b', 'b abc', 'b 0', 'b -5']) {
      expect(errorOf(bad)).toContain('Usage: b');
    }
  });

  it('suggests a close command', () => {
    expect(errorOf('bte')).toContain('Did you mean');
  });

  it('reports an unknown command', () => {
    expect(errorOf('zzzzz')).toContain('Unknown command');
  });

  it('documents the bet in the help text', () => {
    expect(NIUNIU_HELP.some((line) => line.startsWith('b <amount>'))).toBe(true);
  });
});
