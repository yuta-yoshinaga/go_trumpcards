import type { rankAndFileApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RankAndFileArgs = Parameters<typeof rankAndFileApi.exec>;

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

/** Parse a Rank and File CLI command into API exec arguments. */
export function parseRankandfileCommand(input: string): CliParseResult<RankAndFileArgs> {
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

/** Strict zone tokens: only non-negative integer indices (rejects bare 't', 't-1', 't1.5'). */
const TABLEAU_TOKEN = /^t(\d+)$/;
const FOUNDATION_TOKEN = /^f(\d+)$/;

function parseMoveCommand(args: string[]): CliParseResult<RankAndFileArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m w t2, m t0 f, m t0 f<idx>)' };
  }
  const fromTok = args[0].toLowerCase();

  let from: { zone: string; col?: number };
  let toIdx = 1;
  if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else {
    const fromMatch = TABLEAU_TOKEN.exec(fromTok);
    if (!fromMatch) return { error: 'Invalid source: use t<col> (tableau) or w (waste)' };
    from = { zone: 'tableau', col: Number(fromMatch[1]) };
  }

  // Optional cardIndex on a tableau source: m t<col> <i> <target>
  let cardIndex: number | undefined;
  if (from.zone === 'tableau' && args.length >= 3 && /^\d+$/.test(args[1])) {
    const ci = parseIntArg(args, 1);
    if ('error' in ci) return { error: 'Invalid card index' };
    cardIndex = ci.value;
    toIdx = 2;
  }

  const target = args[toIdx]?.toLowerCase() ?? '';
  const tableauTarget = TABLEAU_TOKEN.exec(target);
  if (tableauTarget) {
    return { args: ['move', { ...from, cardIndex }, { zone: 'tableau', col: Number(tableauTarget[1]) }] };
  }
  if (target === 'f') {
    return { args: ['move', { ...from, cardIndex }, { zone: 'foundation' }] };
  }
  const foundationTarget = FOUNDATION_TOKEN.exec(target);
  if (foundationTarget) {
    return { args: ['move', { ...from, cardIndex }, { zone: 'foundation', col: Number(foundationTarget[1]) }] };
  }
  return { error: 'Invalid target: use t<col> (tableau) or f / f<idx> (foundation)' };
}

/** Help text for Rank and File CLI mode. */
export const RANKANDFILE_HELP: string[] = [
  'd/draw          - Draw a card from the stock',
  'm t<c> t<c>     - Move tableau top to column',
  'm t<c> <i> t<c> - Move card at index to column',
  'm w t<c>        - Move waste top to column',
  'm t<c>|w f      - Move to any foundation',
  'm t<c>|w f<i>   - Move to specific foundation',
  'ac/autocomplete - Auto-complete to foundation',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
