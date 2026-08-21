import { describe, expect, it } from 'vitest';
import { parseSlyFoxCommand } from './slyfoxCommands';

describe('parseSlyFoxCommand', () => {
  it.each([
    ['r', 'reset'],
    ['reset', 'reset'],
    ['g', 'giveup'],
    ['giveup', 'giveup'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['u', 'undo'],
    ['undo', 'undo'],
    ['ac', 'autocomplete'],
    ['autocomplete', 'autocomplete'],
    ['l', 'log'],
    ['log', 'log'],
  ])('maps %s to %s', (input, expected) => {
    expect(parseSlyFoxCommand(input)).toEqual({ args: [expected] });
  });

  it('parses dealing onto a reserve slot', () => {
    expect(parseSlyFoxCommand('d 7')).toEqual({
      args: ['deal', undefined, { zone: 'tableau', idx: 7 }],
    });
  });

  it('parses dealing straight to a foundation', () => {
    expect(parseSlyFoxCommand('d f 3')).toEqual({
      args: ['deal', undefined, { zone: 'foundation', idx: 3 }],
    });
  });

  // **配り先は必須。**捨て札が無いので、行き先を決めずには配れない。
  it('rejects a bare deal with no destination', () => {
    expect(parseSlyFoxCommand('d')).toEqual({ error: 'Usage: d <slot> | d f <foundation>' });
  });

  // A tableau card has one legal destination, so the zone letter is optional.
  it('parses a tableau move with and without the trailing f', () => {
    const expected = { args: ['move', { zone: 'tableau', idx: 2 }, { zone: 'foundation' }] };
    expect(parseSlyFoxCommand('m t 2')).toEqual(expected);
    expect(parseSlyFoxCommand('m t 2 f')).toEqual(expected);
  });

  it('rejects a reserve-to-reserve move, which the game does not have', () => {
    expect(parseSlyFoxCommand('m t 0 t 5')).toEqual({
      error: 'Usage: m t <slot> (a reserve card can only go to a foundation)',
    });
  });

  // **捨て札も山札も移動元ではない。**クローン元のコロラドから引き継いだ構文を
  // リネームだけして残すと、サーバが 400 で弾くまで分からない。
  it.each(['m w f', 'm w t 7', 'm s t 3', 'm s f'])('rejects the removed syntax %s', (input) => {
    const result = parseSlyFoxCommand(input);
    expect('error' in result && result.error).toContain('there is no waste');
  });

  it.each([
    ['m', 'Usage: m t <slot> (there is no waste, and the stock is not a move source)'],
    ['m x', 'Usage: m t <slot> (there is no waste, and the stock is not a move source)'],
  ])('reports a usage error for %s', (input, error) => {
    expect(parseSlyFoxCommand(input)).toEqual({ error });
  });

  it.each(['d abc', 'd f abc', 'm t abc'])('reports a bad index in %s', (input) => {
    const result = parseSlyFoxCommand(input);
    expect('error' in result).toBe(true);
  });

  it('suggests a near-miss command', () => {
    const result = parseSlyFoxCommand('del');
    expect('error' in result && result.error).toContain('deal');
  });

  it('reports an unknown command', () => {
    const result = parseSlyFoxCommand('zzz');
    expect('error' in result && result.error).toContain('zzz');
  });
});
