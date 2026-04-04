import { describe, expect, it } from 'vitest';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from './commandParserBase';

describe('splitCommand', () => {
  it('splits a simple command', () => {
    expect(splitCommand('hit')).toEqual({ cmd: 'hit', args: [] });
  });

  it('splits command with arguments', () => {
    expect(splitCommand('b 100')).toEqual({ cmd: 'b', args: ['100'] });
  });

  it('splits command with multiple arguments', () => {
    expect(splitCommand('e 0 2 4')).toEqual({ cmd: 'e', args: ['0', '2', '4'] });
  });

  it('trims whitespace', () => {
    expect(splitCommand('  hit  ')).toEqual({ cmd: 'hit', args: [] });
  });

  it('handles multiple spaces between arguments', () => {
    expect(splitCommand('b   100')).toEqual({ cmd: 'b', args: ['100'] });
  });

  it('returns empty cmd for empty input', () => {
    expect(splitCommand('')).toEqual({ cmd: '', args: [] });
  });

  it('returns empty cmd for whitespace-only input', () => {
    expect(splitCommand('   ')).toEqual({ cmd: '', args: [] });
  });

  it('lowercases the command', () => {
    expect(splitCommand('HIT')).toEqual({ cmd: 'hit', args: [] });
  });
});

describe('parseIntArg', () => {
  it('parses a valid integer', () => {
    expect(parseIntArg(['100'], 0)).toEqual({ value: 100 });
  });

  it('returns error for missing argument', () => {
    expect(parseIntArg([], 0)).toEqual({ error: 'Missing argument at position 0' });
  });

  it('returns error for non-numeric argument', () => {
    expect(parseIntArg(['abc'], 0)).toEqual({ error: 'Invalid number: abc' });
  });

  it('returns error for float', () => {
    expect(parseIntArg(['1.5'], 0)).toEqual({ error: 'Invalid number: 1.5' });
  });

  it('parses zero', () => {
    expect(parseIntArg(['0'], 0)).toEqual({ value: 0 });
  });

  it('parses from specified index', () => {
    expect(parseIntArg(['a', '42'], 1)).toEqual({ value: 42 });
  });
});

describe('parseIntSlice', () => {
  it('parses multiple integers', () => {
    expect(parseIntSlice(['0', '2', '4'])).toEqual({ values: [0, 2, 4] });
  });

  it('returns empty array for empty input', () => {
    expect(parseIntSlice([])).toEqual({ values: [] });
  });

  it('returns error for non-numeric element', () => {
    expect(parseIntSlice(['0', 'x', '2'])).toEqual({ error: 'Invalid number: x' });
  });
});

describe('suggestCommand', () => {
  const commands = ['hit', 'stand', 'bet', 'reset', 'help', 'doubledown', 'split', 'surrender'];

  it('suggests closest command for typo', () => {
    expect(suggestCommand('hti', commands)).toBe('hit');
  });

  it('suggests closest command for partial match', () => {
    expect(suggestCommand('stan', commands)).toBe('stand');
  });

  it('returns null for completely unrelated input', () => {
    expect(suggestCommand('zzzzzzzzz', commands)).toBeNull();
  });

  it('returns null for empty input', () => {
    expect(suggestCommand('', commands)).toBeNull();
  });

  it('returns exact match', () => {
    expect(suggestCommand('bet', commands)).toBe('bet');
  });
});
