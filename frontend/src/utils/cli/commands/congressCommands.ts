import type { congressApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CongressArgs = Parameters<typeof congressApi.exec>;

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

/** Parse a Congress CLI command into API exec arguments. */
export function parseCongressCommand(input: string): CliParseResult<CongressArgs> {
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

function parseMoveCommand(args: string[]): CliParseResult<CongressArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 f, m t0 t5, m w f, m w t2, m s t3)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  let from: { zone: string; col?: number };
  if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else if (fromTok === 's') {
    from = { zone: 'stock' };
  } else if (fromTok.startsWith('t')) {
    const pile = parsePileToken(fromTok);
    if (typeof pile !== 'number') return pile;
    from = { zone: 'tableau', col: pile };
  } else {
    return { error: 'Invalid source: use t<pile> (tableau), w (waste) or s (stock)' };
  }

  if (toTok === 'f') {
    // The stock never reaches a foundation directly -- it can only fill a gap.
    if (from.zone === 'stock') return { error: 'Invalid move: the stock can only fill an empty pile' };
    return { args: ['move', from, { zone: 'foundation' }] };
  }
  if (toTok.startsWith('t')) {
    const pile = parsePileToken(toTok);
    if (typeof pile !== 'number') return pile;
    return { args: ['move', from, { zone: 'tableau', col: pile }] };
  }
  return { error: 'Invalid target: use f (foundation) or t<pile> (tableau)' };
}

/** Help text for Congress CLI mode. */
export const CONGRESS_HELP: string[] = [
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
