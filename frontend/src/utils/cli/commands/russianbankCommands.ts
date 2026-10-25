import type { russianbankApi } from '../../../api/games/russianbank';
import type { RussianBankResponse } from '../../../types/card';
import { RussianBankPhase } from '../../../types/phases';
import {
  russianBankFoundationTarget,
  russianBankTableauTargets,
  selectedRussianBankCard,
} from '../../russianBankRules';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by russianbankApi.exec. */
export type RussianBankCliArgs = Parameters<typeof russianbankApi.exec>;

const VALID_COMMANDS = ['pf', 'mt', 'd', 's', 'u', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse Russian Bank terminal commands into the web API's exact exec arguments. */
export function parseRussianBankCommand(
  input: string,
  state?: RussianBankResponse | null,
): CliParseResult<RussianBankCliArgs> {
  const { cmd, args } = splitCommand(input);
  const active = !!state && state.phase === RussianBankPhase.PLAYING && !state.gameEndFlag && state.isHumanTurn;
  const actionError = 'Not your turn.';
  if (cmd === 'pf' || cmd === 'mt') {
    const usage =
      cmd === 'pf' ? 'Usage: pf <zone> <fromOpp:0|1> <col>' : 'Usage: mt <zone> <fromOpp:0|1> <col> <toCol>';
    const needed = cmd === 'pf' ? 3 : 4;
    if (args.length < needed) return { error: usage };
    if (!active) return { error: actionError };
    if (!state) return { error: actionError };
    const [zone, opponent, col, toCol] = args.map(Number);
    if (!Number.isInteger(zone) || zone < 0 || zone > 2 || ![0, 1].includes(opponent) || !Number.isInteger(col))
      return { error: usage };
    if (zone !== 2 && col !== 0) return { error: usage };
    if (zone === 2 && state && (col < 0 || col >= state.tableau.length)) return { error: usage };
    if (zone === 2 && opponent === 1) return { error: 'Cannot move a card from the opponent tableau.' };
    if (cmd === 'mt' && (!Number.isInteger(toCol) || toCol < 0 || (state && toCol >= state.tableau.length)))
      return { error: usage };
    if (opponent === 1 && !state.players.some((p) => !p.isHuman)) return { error: 'No opponent pile.' };
    if (zone === 2 && !state.tableau[col]?.length) return { error: `No card in tableau column ${col}.` };
    const selected = { zone, fromOpp: opponent === 1, col };
    const card = selectedRussianBankCard(state, selected);
    if (cmd === 'pf' && russianBankFoundationTarget(state, card) < 0)
      return { error: 'No legal foundation destination.' };
    if (cmd === 'mt' && !russianBankTableauTargets(state, card, selected).has(toCol))
      return { error: 'Illegal tableau destination.' };
    if (
      zone === 0 &&
      !(opponent ? state.players.find((p) => !p.isHuman)?.reserveTop : state.players.find((p) => p.isHuman)?.reserveTop)
    )
      return { error: 'No reserve card available.' };
    if (
      zone === 1 &&
      !(opponent ? state.players.find((p) => !p.isHuman)?.wasteTop : state.players.find((p) => p.isHuman)?.wasteTop)
    )
      return { error: 'No waste card available.' };
    if (cmd === 'pf') return { args: ['pf', { zone, fromOpp: opponent === 1, col }] };
    return { args: ['mt', { zone, fromOpp: opponent === 1, col, toCol }] };
  }
  switch (cmd) {
    case 'd':
      return active ? { args: ['d'] } : { error: actionError };
    case 's':
      return active && state?.canCallStop ? { args: ['s'] } : { error: active ? 'Cannot call stop now.' : actionError };
    case 'u':
      return active && state?.canUndo ? { args: ['u'] } : { error: active ? 'Cannot undo now.' : actionError };
    case 'h':
    case 'hint':
      return active ? { args: ['hint'] } : { error: actionError };
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
