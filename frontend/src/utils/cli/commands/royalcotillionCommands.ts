import type { royalcotillionApi } from '../../../api/gameApi';
import type { RoyalCotillionMoveZone } from '../../../types/card';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RoyalCotillionArgs = Parameters<typeof royalcotillionApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'm',
  'move',
  'g',
  'giveup',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'h',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a RoyalCotillion CLI command into API exec arguments. */
export function parseRoyalCotillionCommand(input: string): CliParseResult<RoyalCotillionArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'm':
    case 'move':
      return parseMoveCommand(args);
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

function parsePileToken(tok: string): number | { error: string } {
  const body = tok.slice(1);
  const n = Number(body);
  if (body === '' || Number.isNaN(n)) return { error: 'Usage: t<pile>' };
  return n;
}

function parseMoveCommand(args: string[]): CliParseResult<RoyalCotillionArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 f, m r1 f, m w f, m w t3, m s t3)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  let from: RoyalCotillionMoveZone;
  if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else if (fromTok === 's') {
    from = { zone: 'stock' };
  } else if (fromTok.startsWith('r')) {
    const pile = parsePileToken(fromTok);
    if (typeof pile !== 'number') return pile;
    from = { zone: 'reserve', col: pile };
  } else if (fromTok.startsWith('t')) {
    const slot = parsePileToken(fromTok);
    if (typeof slot !== 'number') return slot;
    from = { zone: 'tableau', col: slot };
  } else {
    return { error: 'Invalid source: use t<slot> (tableau), r<pile> (reserve), w (waste) or s (stock)' };
  }

  if (toTok === 'f') {
    // Only a card already on the board or in the waste can be sent up; the
    // stock has to be turned first.
    if (from.zone === 'stock') return { error: 'Invalid move: the stock can only fill an empty slot' };
    return { args: ['move', from, { zone: 'foundation' }] };
  }
  if (toTok.startsWith('t')) {
    // A slot holds one card and is only ever refilled from the stock or the
    // waste -- a tableau or reserve card has nowhere to go but a foundation.
    if (from.zone === 'tableau' || from.zone === 'reserve') {
      return { error: 'Invalid move: a tableau or reserve card can only go to a foundation' };
    }
    const slot = parsePileToken(toTok);
    if (typeof slot !== 'number') return slot;
    return { args: ['move', from, { zone: 'tableau', col: slot }] };
  }
  return { error: 'Invalid target: use f (foundation) or t<slot> (tableau)' };
}

/** Help text for RoyalCotillion CLI mode. */
export const ROYALCOTILLION_HELP: string[] = [
  'd/draw          - Turn one card from the stock (no redeal)',
  'm t<p> f        - Tableau top to a foundation',
  'm t<p> t<p>     - Move one card between piles',
  'm w f           - Waste to a foundation',
  'm w t<p>        - Waste to a tableau pile',
  'm s t<p>        - Stock straight into an empty pile',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
