import type { missMilliganApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MissMilliganArgs = Parameters<typeof missMilliganApi.exec>;

const VALID_COMMANDS = [
  'd',
  'deal',
  'm',
  'move',
  'wv',
  'waive',
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

/** Parse a Miss Milligan CLI command into API exec arguments. */
export function parseMissMilliganCommand(input: string): CliParseResult<MissMilliganArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'wv':
    case 'waive':
      return parseWaiveCommand(args);
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

/** `wv t<col>[.<idx>]` — lift a run aside; the index defaults to the top card. */
function parseWaiveCommand(args: string[]): CliParseResult<MissMilliganArgs> {
  if (args.length === 0) return { error: 'Usage: wv t<col>[.<idx>]' };
  const parsed = parseTableauToken(args[0]);
  if ('error' in parsed) return parsed;
  return { args: ['waive', { zone: 'tableau', col: parsed.col, cardIndex: parsed.cardIndex }] };
}

function parseTableauToken(tok: string): { col: number; cardIndex?: number } | { error: string } {
  const lower = tok.toLowerCase();
  if (!lower.startsWith('t')) return { error: 'Invalid source: use t<col> (tableau)' };
  const [colTok, idxTok] = lower.slice(1).split('.');
  const col = Number(colTok);
  if (Number.isNaN(col) || colTok === '') return { error: 'Usage: t<col> ...' };
  if (idxTok === undefined) return { col };
  const cardIndex = Number(idxTok);
  if (Number.isNaN(cardIndex)) return { error: 'Usage: t<col>.<idx> ...' };
  return { col, cardIndex };
}

function parseMoveCommand(args: string[]): CliParseResult<MissMilliganArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m t0.2 t5, m t0 f, m w t3, m w f)' };
  }
  const fromTok = args[0].toLowerCase();
  const target = args[1].toLowerCase();

  if (fromTok === 'w') {
    if (target === 'f') return { args: ['move', { zone: 'waived' }, { zone: 'foundation' }] };
    if (target.startsWith('t')) {
      const col = Number(target.slice(1));
      if (Number.isNaN(col)) return { error: 'Usage: m w t<col>' };
      return { args: ['move', { zone: 'waived' }, { zone: 'tableau', col }] };
    }
    return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
  }

  const parsed = parseTableauToken(fromTok);
  if ('error' in parsed) return parsed;

  if (target === 'f') {
    return { args: ['move', { zone: 'tableau', col: parsed.col }, { zone: 'foundation' }] };
  }
  if (target.startsWith('t')) {
    const toCol = Number(target.slice(1));
    if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> t<col>' };
    return {
      args: [
        'move',
        { zone: 'tableau', col: parsed.col, cardIndex: parsed.cardIndex },
        { zone: 'tableau', col: toCol },
      ],
    };
  }
  return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
}

/** Help text for Miss Milligan CLI mode. */
export const MISSMILLIGAN_HELP: string[] = [
  'd/deal          - Deal one card onto every column',
  'm t<c> f        - Tableau top to a foundation',
  'm t<c> t<c>     - Move the top card between columns',
  'm t<c>.<i> t<c> - Move the run starting at index i',
  'wv t<c>[.<i>]   - Waive a run aside (stock must be gone)',
  'm w t<c>        - Put the waived cards back on a column',
  'm w f           - Send a single waived card to a foundation',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
