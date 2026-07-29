import { describe, expect, it } from 'vitest';
import { PONTOON_HELP, parsePontoonCommand } from './pontoonCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parsePontoonCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parsePontoonCommand', () => {
  it('parses the commands that carry no amount', () => {
    expect(parsePontoonCommand('deal')).toEqual({ args: ['deal'] });
    expect(parsePontoonCommand('s')).toEqual({ args: ['stick'] });
    expect(parsePontoonCommand('stick')).toEqual({ args: ['stick'] });
    expect(parsePontoonCommand('t')).toEqual({ args: ['twist'] });
    expect(parsePontoonCommand('twist')).toEqual({ args: ['twist'] });
    expect(parsePontoonCommand('sp')).toEqual({ args: ['split'] });
    expect(parsePontoonCommand('split')).toEqual({ args: ['split'] });
    expect(parsePontoonCommand('bt')).toEqual({ args: ['bankertwist'] });
    expect(parsePontoonCommand('bs')).toEqual({ args: ['bankerstay'] });
    expect(parsePontoonCommand('log')).toEqual({ args: ['log'] });
    expect(parsePontoonCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses bet and buy with their amount', () => {
    expect(parsePontoonCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parsePontoonCommand('bet 100')).toEqual({ args: ['bet', 100] });
    expect(parsePontoonCommand('buy 50')).toEqual({ args: ['buy', 50] });
  });

  it('rejects a missing or unusable amount', () => {
    expect(errorOf('b')).toContain('Usage');
    expect(errorOf('b abc')).toContain('Usage');
    expect(errorOf('b 0')).toContain('Usage');
    expect(errorOf('b -5')).toContain('Usage');
    expect(errorOf('buy')).toContain('Usage');
    expect(errorOf('buy abc')).toContain('Usage');
  });

  it('suggests a close command', () => {
    expect(errorOf('stik')).toContain('Did you mean');
  });

  it('reports an unknown command', () => {
    expect(errorOf('zzzzz')).toContain('Unknown command');
  });

  it('documents every action in the help text', () => {
    for (const prefix of ['b <amount>', 'deal', 's/stick', 't/twist', 'buy <amount>', 'sp/split', 'bt', 'bs']) {
      expect(PONTOON_HELP.some((line) => line.startsWith(prefix))).toBe(true);
    }
  });
});
