import type { somersetApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SomersetArgs = Parameters<typeof somersetApi.exec>;

const VALID_COMMANDS = [
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

/** Parse a Somerset CLI command into API exec arguments. */
export function parseSomersetCommand(input: string): CliParseResult<SomersetArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
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

function parseMoveCommand(args: string[]): CliParseResult<SomersetArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m t0 f, m t0 f<idx>)' };
  }
  const fromTok = args[0].toLowerCase();

  if (!fromTok.startsWith('t')) {
    return { error: 'Invalid source: use t<col> (tableau)' };
  }
  const fromCol = Number(fromTok.slice(1));
  if (Number.isNaN(fromCol)) return { error: 'Usage: m t<col> ...' };

  // Optional cardIndex on source: m t<col> <i> t<col>|f
  let cardIndex: number | undefined;
  let toIdx = 1;
  if (args.length >= 3 && /^\d+$/.test(args[1])) {
    const ci = parseIntArg(args, 1);
    if ('error' in ci) return { error: 'Invalid card index' };
    cardIndex = ci.value;
    toIdx = 2;
  }

  const target = args[toIdx]?.toLowerCase() ?? '';
  if (target.startsWith('t')) {
    const toCol = Number(target.slice(1));
    if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> t<col>' };
    return {
      args: ['move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'tableau', col: toCol }],
    };
  }
  if (target === 'f') {
    return { args: ['move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'foundation' }] };
  }
  if (target.startsWith('f')) {
    const fCol = Number(target.slice(1));
    if (Number.isNaN(fCol)) return { error: 'Usage: m t<col> f<idx>' };
    return {
      args: ['move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'foundation', col: fCol }],
    };
  }
  return { error: 'Invalid target: use t<col> (tableau) or f / f<idx> (foundation)' };
}

/** Help text for Somerset CLI mode. */
export const SOMERSET_HELP: string[] = [
  'm t<c> t<c>     - Move tableau top to column',
  'm t<c> <i> t<c> - Move card at index to column',
  'm t<c> f        - Move to any foundation',
  'm t<c> f<i>     - Move to specific foundation',
  'ac/autocomplete - Auto-complete to foundation',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
