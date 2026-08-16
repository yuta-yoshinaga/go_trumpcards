import type { zwickerApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ZwickerArgs = Parameters<typeof zwickerApi.exec>;

const VALID_COMMANDS = [
  't',
  'take',
  'b',
  'build',
  'tr',
  'trail',
  'n',
  'next',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a non-negative integer argument, or null when it is not one. */
function parseIdx(raw: string | undefined): number | null {
  if (raw === undefined || raw.trim() === '') return null;
  const n = Number(raw);
  if (!Number.isInteger(n) || n < 0) return null;
  return n;
}

/** Parse `t:0,2` / `b:1` / `0,2` into indices plus which list they belong to. */
function parseList(raw: string): { list: number[]; kind: 't' | 'b' } | null {
  let kind: 't' | 'b' = 't';
  let body = raw.trim();
  const colon = body.indexOf(':');
  if (colon >= 0) {
    const prefix = body.slice(0, colon).trim().toLowerCase();
    if (prefix !== 't' && prefix !== 'b') return null;
    kind = prefix;
    body = body.slice(colon + 1);
  }
  if (body === '') return null;
  const list: number[] = [];
  for (const part of body.split(',')) {
    const n = parseIdx(part);
    if (n === null) return null;
    list.push(n);
  }
  return { list, kind };
}

/** Parse a Zwicker CLI command into API exec arguments. */
export function parseZwickerCommand(input: string): CliParseResult<ZwickerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'take': {
      if (args.length < 2) return { error: `Usage: ${cmd} <card> <value> t:<a,b>` };
      const cardIndex = parseIdx(args[0]);
      if (cardIndex === null) return { error: `Invalid card index: ${args[0]}` };
      // **値は札とは別の引数。**A と絵札は 2 つの値を持つので、札だけでは
      // どちらのつもりか決まらない。
      const playedValue = Number(args[1]);
      if (!Number.isInteger(playedValue) || playedValue <= 0) {
        return { error: `Invalid value: ${args[1]}` };
      }
      const tableIndices: number[] = [];
      const buildIndices: number[] = [];
      for (const raw of args.slice(2)) {
        const parsed = parseList(raw);
        if (parsed === null) return { error: `Invalid selection: ${raw}` };
        (parsed.kind === 'b' ? buildIndices : tableIndices).push(...parsed.list);
      }
      if (tableIndices.length === 0 && buildIndices.length === 0) {
        return { error: `Usage: ${cmd} <card> <value> t:<a,b>` };
      }
      return { args: ['take', { cardIndex, playedValue, tableIndices, buildIndices }] };
    }
    case 'b':
    case 'build': {
      if (args.length < 3) return { error: `Usage: ${cmd} <card> <a,b> <value>` };
      const cardIndex = parseIdx(args[0]);
      if (cardIndex === null) return { error: `Invalid card index: ${args[0]}` };
      const parsed = parseList(args[1]);
      if (parsed === null || parsed.list.length === 0) {
        return { error: `Invalid table selection: ${args[1]}` };
      }
      const declaredValue = Number(args[2]);
      if (!Number.isInteger(declaredValue) || declaredValue <= 0) {
        return { error: `Invalid value: ${args[2]}` };
      }
      return { args: ['build', { cardIndex, tableIndices: parsed.list, declaredValue }] };
    }
    case 'tr':
    case 'trail': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const cardIndex = parseIdx(args[0]);
      if (cardIndex === null) return { error: `Invalid card index: ${args[0]}` };
      return { args: ['trail', { cardIndex }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Zwicker CLI mode. */
export const ZWICKER_HELP: string[] = [
  't <i> <v> t:<a,b> - Play hand card i as value v and take table cards a,b',
  't <i> <v> b:<c>   - ...and/or take build c',
  'b <i> <a,b> <v>   - Build value v from hand card i and table cards a,b',
  'tr <i>            - Put hand card i on the table',
  'n/next            - Deal the next hand',
  'log               - Show action log',
  'r/reset           - New game',
  'h/hint      - Get a hint',
];
