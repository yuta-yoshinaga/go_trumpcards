import { describe, expect, it } from 'vitest';
import { CARIBBEANDRAW_HELP, parseCaribbeandrawCommand } from './caribbeandrawCommands';

describe('parseCaribbeandrawCommand', () => {
  it('parses bet with amount', () => {
    expect(parseCaribbeandrawCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseCaribbeandrawCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with amount and jackpot side bet', () => {
    expect(parseCaribbeandrawCommand('b 100 10')).toEqual({ args: ['bet', 100, 10] });
  });

  it('returns error for bet without amount', () => {
    const result = parseCaribbeandrawCommand('b');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid jackpot value', () => {
    const result = parseCaribbeandrawCommand('b 100 abc');
    expect('error' in result).toBe(true);
  });

  describe('draw', () => {
    it('converts the 1-based card numbers the player types into 0-based indices', () => {
      // The screen numbers the hand from 1; the API takes positions from 0.
      // Passing the typed numbers straight through would discard a different
      // card than the one named, and nothing would report an error.
      expect(parseCaribbeandrawCommand('d 1 3')).toEqual({ args: ['draw', undefined, undefined, [0, 2]] });
      expect(parseCaribbeandrawCommand('draw 5')).toEqual({ args: ['draw', undefined, undefined, [4]] });
    });

    it('stands pat on a bare d', () => {
      expect(parseCaribbeandrawCommand('d')).toEqual({ args: ['draw', undefined, undefined, []] });
      expect(parseCaribbeandrawCommand('draw')).toEqual({ args: ['draw', undefined, undefined, []] });
    });

    it('rejects more than 2 cards', () => {
      const result = parseCaribbeandrawCommand('d 1 2 3');
      expect(result).toEqual({ error: expect.stringContaining('at most 2') });
    });

    it('rejects a card number outside the hand', () => {
      expect(parseCaribbeandrawCommand('d 0')).toEqual({ error: expect.stringContaining('1-5') });
      expect(parseCaribbeandrawCommand('d 6')).toEqual({ error: expect.stringContaining('1-5') });
    });

    it('rejects the same card named twice', () => {
      expect(parseCaribbeandrawCommand('d 2 2')).toEqual({ error: expect.stringContaining('twice') });
    });

    it('rejects a non-numeric card number', () => {
      expect(parseCaribbeandrawCommand('d x')).toEqual({ error: expect.stringContaining('1-5') });
    });
  });

  it('parses play', () => {
    expect(parseCaribbeandrawCommand('p')).toEqual({ args: ['play'] });
    expect(parseCaribbeandrawCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseCaribbeandrawCommand('f')).toEqual({ args: ['fold'] });
    expect(parseCaribbeandrawCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseCaribbeandrawCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseCaribbeandrawCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCaribbeandrawCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseCaribbeandrawCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('suggests command for typo', () => {
    const result = parseCaribbeandrawCommand('bett 100');
    expect('error' in result).toBe(true);
  });

  it('accepts hint under both spellings and lists it in help', () => {
    expect(parseCaribbeandrawCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCaribbeandrawCommand('hint')).toEqual({ args: ['hint'] });
    expect(CARIBBEANDRAW_HELP.some((l) => l.includes('hint'))).toBe(true);
  });

  it('documents the draw command in the help text', () => {
    const line = CARIBBEANDRAW_HELP.find((l) => l.startsWith('d '));
    expect(line).toBeDefined();
    expect(line).toContain('2 cards');
    expect(line).toContain('stands pat');
  });
});
