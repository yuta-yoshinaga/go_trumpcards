import type { bisleyApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BisleyArgs = Parameters<typeof bisleyApi.exec>;

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

/** Parse a Bisley CLI command into API exec arguments. */
export function parseBisleyCommand(input: string): CliParseResult<BisleyArgs> {
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

function parseMoveCommand(args: string[]): CliParseResult<BisleyArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m t0 a, m t0 k)' };
  }
  const fromTok = args[0].toLowerCase();

  if (!fromTok.startsWith('t')) {
    return { error: 'Invalid source: use t<col> (tableau)' };
  }
  const fromCol = Number(fromTok.slice(1));
  if (Number.isNaN(fromCol)) return { error: 'Usage: m t<col> ...' };

  const target = args[1].toLowerCase();
  if (target === 'a') {
    return { args: ['move', { zone: 'tableau', col: fromCol }, { zone: 'ace' }] };
  }
  if (target === 'k') {
    return { args: ['move', { zone: 'tableau', col: fromCol }, { zone: 'king' }] };
  }
  if (target.startsWith('t')) {
    const toCol = Number(target.slice(1));
    if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> t<col>' };
    return { args: ['move', { zone: 'tableau', col: fromCol }, { zone: 'tableau', col: toCol }] };
  }
  return { error: 'Invalid target: use t<col> (tableau), a (ascending) or k (descending)' };
}

/** Help text for Bisley CLI mode. */
export const BISLEY_HELP: string[] = [
  'm t<c> t<c>     - Move tableau top to column (same suit, +/-1)',
  'm t<c> a        - Move to the ascending (A->K) foundation',
  'm t<c> k        - Move to the descending (K->A) foundation',
  'ac/autocomplete - Auto-complete to foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
