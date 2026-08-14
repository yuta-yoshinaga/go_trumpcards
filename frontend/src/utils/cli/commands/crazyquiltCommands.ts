import type { crazyquiltApi } from '../../../api/gameApi';
import type { CrazyQuiltMoveZone } from '../../../types/card';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CrazyQuiltArgs = Parameters<typeof crazyquiltApi.exec>;

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

/** Parse a CrazyQuilt CLI command into API exec arguments. */
export function parseCrazyQuiltCommand(input: string): CliParseResult<CrazyQuiltArgs> {
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

function parseMoveCommand(args: string[]): CliParseResult<CrazyQuiltArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m q0 f, m q7 w, m w f)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  let from: CrazyQuiltMoveZone;
  if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else if (fromTok.startsWith('q')) {
    const cell = parsePileToken(fromTok);
    if (typeof cell !== 'number') return cell;
    from = { zone: 'quilt', col: cell };
  } else {
    // The stock is not a source: it is only ever turned.
    return { error: 'Invalid source: use q<cell> (quilt) or w (waste)' };
  }

  if (toTok === 'f') {
    return { args: ['move', from, { zone: 'foundation' }] };
  }
  if (toTok === 'w') {
    // Only a quilt card goes onto the waste, and only when it is one rank away.
    if (from.zone !== 'quilt') {
      return { error: 'Invalid move: only a quilt card can go onto the waste' };
    }
    return { args: ['move', from, { zone: 'waste' }] };
  }
  return { error: 'Invalid target: use f (foundation) or w (waste)' };
}

/** Help text for CrazyQuilt CLI mode. */
export const CRAZYQUILT_HELP: string[] = [
  'd/draw          - Turn one card from the stock (no redeal)',
  'm q<c> f        - A quilt card to a foundation',
  'm q<c> w        - A quilt card onto the waste (one rank away)',
  'm w f           - Waste to a foundation',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
