import { describe, expect, it } from 'vitest';
import { parseSetteEMezzoCommand, SETTEMEZZO_HELP } from './settemezzoCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseSetteEMezzoCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseSetteEMezzoCommand', () => {
  it('parses the commands that carry no amount', () => {
    expect(parseSetteEMezzoCommand('deal')).toEqual({ args: ['deal'] });
    expect(parseSetteEMezzoCommand('h')).toEqual({ args: ['hit'] });
    expect(parseSetteEMezzoCommand('hit')).toEqual({ args: ['hit'] });
    expect(parseSetteEMezzoCommand('s')).toEqual({ args: ['stand'] });
    expect(parseSetteEMezzoCommand('stand')).toEqual({ args: ['stand'] });
    expect(parseSetteEMezzoCommand('bh')).toEqual({ args: ['bankerhit'] });
    expect(parseSetteEMezzoCommand('bs')).toEqual({ args: ['bankerstand'] });
    expect(parseSetteEMezzoCommand('log')).toEqual({ args: ['log'] });
    expect(parseSetteEMezzoCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses a bet', () => {
    expect(parseSetteEMezzoCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseSetteEMezzoCommand('bet 100')).toEqual({ args: ['bet', 100] });
  });

  // The player types POINTS and the API takes HALVES. A wrong conversion would
  // silently halve or double the matta, which no error would reveal.
  it('converts the matta from points to halves', () => {
    expect(parseSetteEMezzoCommand('matta 0.5')).toEqual({ args: ['matta', 1] });
    expect(parseSetteEMezzoCommand('matta 1')).toEqual({ args: ['matta', 2] });
    expect(parseSetteEMezzoCommand('matta 3')).toEqual({ args: ['matta', 6] });
    expect(parseSetteEMezzoCommand('matta 7')).toEqual({ args: ['matta', 14] });
  });

  // 0.5 is the only fraction the matta may take; 1.5 and friends are not values.
  it('rejects a matta value outside 0.5 and 1-7', () => {
    for (const bad of ['matta', 'matta 0', 'matta 8', 'matta 1.5', 'matta abc', 'matta -1']) {
      expect(errorOf(bad)).toContain('Usage: matta');
    }
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
    for (const prefix of ['b <amount>', 'deal', 'h/hit', 's/stand', 'matta <v>', 'bh', 'bs']) {
      expect(SETTEMEZZO_HELP.some((line) => line.startsWith(prefix))).toBe(true);
    }
  });
});
