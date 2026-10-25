import type { trucoApi } from '../../../api/gameApi';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by trucoApi.exec. */
export type TrucoCliArgs = Parameters<typeof trucoApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  't',
  'truco',
  'a',
  'accept',
  'd',
  'decline',
  'n',
  'next',
  'h',
  'hint',
  'r',
  'reset',
];

/** Parse a Truco terminal command into the same arguments used by page actions. */
export function parseTrucoCommand(input: string): CliParseResult<TrucoCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (!Number.isInteger(idx) || idx < 0 || idx > 2 || String(idx) !== args[0]) return { error: 'Usage: p <0-2>' };
      return { args: ['play', idx] };
    }
    case 't':
    case 'truco':
      return { args: ['truco'] };
    case 'a':
    case 'accept':
      return { args: ['accept'] };
    case 'd':
    case 'decline':
      return { args: ['decline'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Truco CLI terminal. */
export const TRUCO_HELP: string[] = [
  'p <idx>       - Play a card by its 0-based hand index',
  't / truco     - Declare or raise Truco',
  'a / accept    - Accept a Truco call',
  'd / decline   - Decline a Truco call',
  'n / next      - Advance after a trick or hand',
  'h / hint      - Show a hint',
  'r / reset     - Reset the game',
];
