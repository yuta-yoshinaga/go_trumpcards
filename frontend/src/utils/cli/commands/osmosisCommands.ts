import type { osmosisApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type OsmosisArgs = Parameters<typeof osmosisApi.exec>;

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

/** Parse an Osmosis CLI command into API exec arguments. */
export function parseOsmosisCommand(input: string): CliParseResult<OsmosisArgs> {
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

/**
 * Resolve a foundation row index from the remaining tokens, accepting both
 * compact (`f2`) and spaced (`f 2`) forms as well as a bare index (`2`).
 */
function resolveFoundationIndex(tokens: string[]): number | undefined {
  if (tokens.length === 0) return undefined;
  const first = tokens[0].toLowerCase();
  if (first === 'f' || first === 'foundation') {
    if (tokens.length < 2) return undefined;
    const n = Number(tokens[1]);
    return Number.isNaN(n) ? undefined : n;
  }
  const raw = first.startsWith('f') ? first.slice(1) : first;
  if (raw === '') return undefined;
  const n = Number(raw);
  return Number.isNaN(n) ? undefined : n;
}

function parseMoveCommand(args: string[]): CliParseResult<OsmosisArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m w f<seg> | m r<col> f<seg> (e.g., m w f0, m r1 f2)' };
  }
  const src = args[0].toLowerCase();

  // Waste source: m w f<seg>
  if (src === 'w' || src === 'waste') {
    const fIdx = resolveFoundationIndex(args.slice(1));
    if (fIdx === undefined) return { error: 'Usage: m w f<seg> (e.g., m w f0)' };
    return { args: ['move', { zone: 'waste' }, { zone: 'foundation', col: fIdx }] };
  }

  // Reserve source: m r<col> f<seg> (compact) or m r <col> f <seg> (spaced).
  if (src.startsWith('r')) {
    const compact = src !== 'r' && src !== 'reserve';
    const colRaw = compact ? src.slice(1) : args[1];
    const rest = compact ? args.slice(1) : args.slice(2);
    const rCol = Number(colRaw);
    if (Number.isNaN(rCol)) return { error: 'Usage: m r<col> f<seg> (e.g., m r1 f2)' };
    const fIdx = resolveFoundationIndex(rest);
    if (fIdx === undefined) return { error: 'Usage: m r<col> f<seg> (e.g., m r1 f2)' };
    return { args: ['move', { zone: 'reserve', col: rCol }, { zone: 'foundation', col: fIdx }] };
  }

  return { error: 'Invalid source: use w (waste) or r<col> (reserve)' };
}

/** Help text for Osmosis CLI mode. */
export const OSMOSIS_HELP: string[] = [
  'd/draw          - Draw 1 from stock',
  'm w f<seg>      - Move waste to foundation row <seg>',
  'm r<col> f<seg> - Move reserve column to foundation row <seg>',
  'ac/autocomplete - Auto-complete to foundations',
  'u/undo          - Undo last move',
  'h/hint          - Get a hint',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
