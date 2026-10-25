import type { russianbankApi } from '../../../api/games/russianbank';
import type { RussianBankResponse } from '../../../types/card';
import { RussianBankPhase } from '../../../types/phases';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by russianbankApi.exec. */
export type RussianBankCliArgs = Parameters<typeof russianbankApi.exec>;

const VALID_COMMANDS = ['pf', 'mt', 'd', 's', 'u', 'h', 'r', 'help', '?'];

/** Parse Russian Bank terminal commands into the web API's exact exec arguments. */
export function parseRussianBankCommand(
  input: string,
  state?: RussianBankResponse | null,
): CliParseResult<RussianBankCliArgs> {
  const { cmd, args } = splitCommand(input);
  const active = state && state.phase === RussianBankPhase.PLAYING && !state.gameEndFlag && state.isHumanTurn;
  if (cmd === 'pf' || cmd === 'mt') {
    const usage =
      cmd === 'pf' ? 'Usage: pf <zone> <fromOpp:0|1> <col>' : 'Usage: mt <zone> <fromOpp:0|1> <col> <toCol>';
    const needed = cmd === 'pf' ? 3 : 4;
    if (args.length < needed) return { error: usage };
    const [zone, opponent, col, toCol] = args.map(Number);
    if (!Number.isInteger(zone) || zone < 0 || zone > 2 || ![0, 1].includes(opponent) || !Number.isInteger(col))
      return { error: usage };
    if (zone !== 2 && col !== 0) return { error: usage };
    if (zone === 2 && state && (col < 0 || col >= state.tableau.length)) return { error: usage };
    if (cmd === 'mt' && (!Number.isInteger(toCol) || toCol < 0 || (state && toCol >= state.tableau.length)))
      return { error: usage };
    if (active && opponent === 1 && !state.players.some((p) => !p.isHuman)) return { error: 'No opponent pile.' };
    if (active && zone === 2 && !state.tableau[col]?.length) return { error: `No card in tableau column ${col}.` };
    if (
      active &&
      zone === 0 &&
      !(opponent ? state.players.find((p) => !p.isHuman)?.reserveTop : state.players.find((p) => p.isHuman)?.reserveTop)
    )
      return { error: 'No reserve card available.' };
    if (
      active &&
      zone === 1 &&
      !(opponent ? state.players.find((p) => !p.isHuman)?.wasteTop : state.players.find((p) => p.isHuman)?.wasteTop)
    )
      return { error: 'No waste card available.' };
    if (cmd === 'pf') return { args: ['pf', { zone, fromOpp: opponent === 1, col }] };
    return { args: ['mt', { zone, fromOpp: opponent === 1, col, toCol }] };
  }
  switch (cmd) {
    case 'd':
      return { args: ['d'] };
    case 's':
      return { args: ['s'] };
    case 'u':
      return { args: ['u'] };
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

/** Help text for Russian Bank terminal mode. */
export const RUSSIANBANK_HELP = [
  'pf <zone> <fromOpp:0|1> <col> - Move a pile top to its foundation',
  'mt <zone> <fromOpp:0|1> <col> <toCol> - Move a pile top to a tableau column',
  'Zones: 0=reserve, 1=waste, 2=tableau; use col 0 except for tableau sources',
  'd - Discard/draw',
  's - Call stop',
  'u - Undo',
  'h / hint - Request a hint',
  'r / reset - Reset game',
];
