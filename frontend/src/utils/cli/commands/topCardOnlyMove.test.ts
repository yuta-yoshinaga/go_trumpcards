import { describe, expect, it } from 'vitest';
import { BAKERSDOZEN_HELP, parseBakersDozenCommand } from './bakersdozenCommands';
import { BELEAGUEREDCASTLE_HELP, parseBeleagueredCastleCommand } from './beleagueredcastleCommands';
import { FORTRESS_HELP, parseFortressCommand } from './fortressCommands';
import { parseSomersetCommand, SOMERSET_HELP } from './somersetCommands';
import { parseStreetsAndAlleysCommand, STREETSANDALLEYS_HELP } from './streetsandalleysCommands';

/**
 * These five solitaires move the top card of a column and nothing else: their
 * domain answers "only the top card can be moved" to any other index, and the
 * server never even asks -- `dispatchTopCardMove` passes -1 in place of
 * whatever the client sent.
 *
 * The CLI used to advertise `m t<c> <i> t<c>` and parse the index anyway, so
 * `m t0 1 t5` moved the top card instead of the card at index 1 and said
 * nothing about it. A per-game parse test cannot catch that: it asserts the
 * parse result, and the parse result was correct -- it was the server that
 * threw the value away. The fix is to stop offering the form, and this pins
 * both halves of it.
 */
const GAMES = [
  { name: 'bakersdozen', help: BAKERSDOZEN_HELP, parse: parseBakersDozenCommand },
  { name: 'beleagueredcastle', help: BELEAGUEREDCASTLE_HELP, parse: parseBeleagueredCastleCommand },
  { name: 'fortress', help: FORTRESS_HELP, parse: parseFortressCommand },
  { name: 'somerset', help: SOMERSET_HELP, parse: parseSomersetCommand },
  { name: 'streetsandalleys', help: STREETSANDALLEYS_HELP, parse: parseStreetsAndAlleysCommand },
] as const;

describe('top-card-only solitaires do not offer a card index', () => {
  it.each(GAMES)('$name help never advertises an index form', ({ help }) => {
    // The negative control: the help must still describe the move it does have,
    // or an empty help array would satisfy the assertion below.
    expect(help.some((line) => line.startsWith('m t<c> t<c>'))).toBe(true);
    expect(help.filter((line) => /^m t<c> <i>/.test(line))).toEqual([]);
  });

  it.each(GAMES)('$name rejects m t0 1 t5 instead of silently moving the top card', ({ parse }) => {
    const result = parse('m t0 1 t5');
    expect(result).toHaveProperty('error');
    expect('args' in result).toBe(false);
  });

  it.each(GAMES)('$name still parses the top-card moves it does support', ({ parse }) => {
    expect(parse('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }],
    });
    expect(parse('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
    expect(parse('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 2 }],
    });
  });
});
