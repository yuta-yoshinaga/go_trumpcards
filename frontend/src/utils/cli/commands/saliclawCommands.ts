import type { salicLawApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SalicLawArgs = Parameters<typeof salicLawApi.exec>;

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

/** Parse a Salic Law CLI command into API exec arguments. */
export function parseSalicLawCommand(input: string): CliParseResult<SalicLawArgs> {
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
  if (body === '' || Number.isNaN(n)) return { error: 'Usage: t<column>' };
  return n;
}

function parseMoveCommand(args: string[]): CliParseResult<SalicLawArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 f, m t0 t5)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  // **移動元はタブローだけ。**捨て札は無く、山札の札は配られた時点で列に乗る。
  if (!fromTok.startsWith('t')) {
    return { error: 'Invalid source: use t<column> (there is no waste, and the stock is not a move source)' };
  }
  const fromPile = parsePileToken(fromTok);
  if (typeof fromPile !== 'number') return fromPile;
  const from: { zone: string; col?: number } = { zone: 'tableau', col: fromPile };

  if (toTok === 'f') {
    return { args: ['move', from, { zone: 'foundation' }] };
  }
  if (toTok.startsWith('t')) {
    const pile = parsePileToken(toTok);
    if (typeof pile !== 'number') return pile;
    return { args: ['move', from, { zone: 'tableau', col: pile }] };
  }
  return { error: 'Invalid target: use f (foundation) or t<column> (a column holding just its king)' };
}

/** Help text for Salic Law CLI mode. */
export const SALICLAW_HELP: string[] = [
  'd/draw          - Deal one card onto the tableau (a king opens the next column)',
  'm t<c> f        - Column top to a foundation (ace up to jack, any suit)',
  'm t<c> t<c>     - Move one card onto a column holding just its king',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
