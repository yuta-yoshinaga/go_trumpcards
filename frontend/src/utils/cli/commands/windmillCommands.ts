import type { windmillApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type WindmillArgs = Parameters<typeof windmillApi.exec>;

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

/** Parse a Windmill CLI command into API exec arguments. */
export function parseWindmillCommand(input: string): CliParseResult<WindmillArgs> {
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

/** `s<n>` sail, `w` waste, `k<n>` corner, `c` centre. */
function parseIndexedToken(tok: string, prefix: string): number | { error: string } {
  const body = tok.slice(prefix.length);
  const n = Number(body);
  if (body === '' || Number.isNaN(n)) return { error: `Usage: ${prefix}<index>` };
  return n;
}

function parseMoveCommand(args: string[]): CliParseResult<WindmillArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m s0 c, m s0 k1, m w c, m w k2, m k0 c)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  let from: WindmillMoveZoneLike;
  if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else if (fromTok.startsWith('s')) {
    const idx = parseIndexedToken(fromTok, 's');
    if (typeof idx !== 'number') return idx;
    from = { zone: 'sail', col: idx };
  } else if (fromTok.startsWith('k')) {
    const idx = parseIndexedToken(fromTok, 'k');
    if (typeof idx !== 'number') return idx;
    from = { zone: 'corner', col: idx };
  } else {
    return { error: 'Invalid source: use s<n> (sail), w (waste) or k<n> (corner)' };
  }

  if (toTok === 'c') {
    return { args: ['move', from, { zone: 'center' }] };
  }
  if (toTok.startsWith('k')) {
    // The rescue runs one way only, so a corner is never a source AND a target.
    if (from.zone === 'corner') return { error: 'Invalid move: a corner card can only go to the centre' };
    const idx = parseIndexedToken(toTok, 'k');
    if (typeof idx !== 'number') return idx;
    return { args: ['move', from, { zone: 'corner', col: idx }] };
  }
  return { error: 'Invalid target: use c (centre) or k<n> (corner)' };
}

/** Local shape for a parsed zone; mirrors WindmillMoveZone without the import cycle. */
type WindmillMoveZoneLike = { zone: string; col?: number };

/** Help text for Windmill CLI mode. */
export const WINDMILL_HELP: string[] = [
  'd/draw          - Turn one card from the stock',
  'm s<n> c        - Sail card to the centre foundation',
  'm s<n> k<n>     - Sail card to a corner foundation',
  'm w c           - Waste to the centre foundation',
  'm w k<n>        - Waste to a corner foundation',
  'm k<n> c        - Pull a corner top back onto the centre (rescue)',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
