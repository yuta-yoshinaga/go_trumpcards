import { describe, expect, it } from 'vitest';
import { parseRussianSolitaireCommand } from './russiansolitaireCommands';

describe('parseRussianSolitaireCommand', () => {
  it('parses a two-column move as a top-of-column block (cardIndex -1)', () => {
    expect(parseRussianSolitaireCommand('m 0 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: -1 }, { zone: 'tableau', col: 3 }],
    });
    expect(parseRussianSolitaireCommand('move 1 2')).toEqual({
      args: ['move', { zone: 'tableau', col: 1, cardIndex: -1 }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses a three-argument move with an explicit card index', () => {
    expect(parseRussianSolitaireCommand('m 0 2 5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('errors on a non-numeric column', () => {
    expect(parseRussianSolitaireCommand('m a 3')).toEqual({ error: 'Invalid column' });
    expect(parseRussianSolitaireCommand('m a 2 3')).toEqual({ error: 'Invalid column' });
  });

  it('errors on a negative or non-numeric card index', () => {
    expect(parseRussianSolitaireCommand('m 0 -1 3')).toEqual({ error: 'Invalid card index' });
    expect(parseRussianSolitaireCommand('m 0 z 3')).toEqual({ error: 'Invalid card index' });
  });

  it('errors on the wrong number of move arguments', () => {
    expect(parseRussianSolitaireCommand('m 0')).toEqual({ error: 'Usage: m <fromCol> [cardIdx] <toCol>' });
    expect(parseRussianSolitaireCommand('m 0 1 2 3')).toEqual({ error: 'Usage: m <fromCol> [cardIdx] <toCol>' });
  });

  it('parses the simple commands', () => {
    expect(parseRussianSolitaireCommand('r')).toEqual({ args: ['reset'] });
    expect(parseRussianSolitaireCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseRussianSolitaireCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseRussianSolitaireCommand('h')).toEqual({ args: ['hint'] });
    expect(parseRussianSolitaireCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseRussianSolitaireCommand('u')).toEqual({ args: ['undo'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseRussianSolitaireCommand('rese');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseRussianSolitaireCommand('zzz')).toBe(true);
  });
});
