import type { barbuApi } from '../../../api/gameApi';
import type { BarbuResponse } from '../../../types/card';
import type { CliParseResult } from '../types';

/** Args tuple accepted by barbuApi.exec. */
export type BarbuCliArgs = Parameters<typeof barbuApi.exec>;

/** Parses a single CLI command line for the Barbu game.
 *
 * Accepted syntax:
 *   r / reset               reset game
 *   n / next                start next deal
 *   c <0-6> [trump 1-4]     pick a contract (trump only for Trumps=5)
 *   p <hand>                play hand card (-1 to pass in Dominoes)
 *   l / log                 show action log
 */
export function parseBarbuCommand(input: string): CliParseResult<BarbuCliArgs> {
  const parts = input.trim().toLowerCase().split(/\s+/).filter(Boolean);
  const cmd = parts[0] ?? '';
  if (cmd === 'reset' || cmd === 'r') return { args: ['r'] };
  if (cmd === 'next' || cmd === 'n') return { args: ['n'] };
  if (cmd === 'log' || cmd === 'l') return { args: ['log'] };

  if (cmd === 'c' || cmd === 'contract') {
    const rest = parts.slice(1);
    if (rest.length === 0) return { error: 'Usage: c <0-6> [trump 1-4]' };
    const contract = Number.parseInt(rest[0], 10);
    if (Number.isNaN(contract)) return { error: 'Invalid contract' };
    const trumpSuit = rest.length > 1 ? Number.parseInt(rest[1], 10) : -1;
    if (Number.isNaN(trumpSuit)) return { error: 'Invalid trump suit' };
    return { args: ['c', { contract, trumpSuit }] };
  }

  if (cmd === 'p' || cmd === 'play') {
    const rest = parts.slice(1);
    if (rest.length === 0) return { error: 'Usage: p <hand> (-1 to pass)' };
    const hand = Number.parseInt(rest[0], 10);
    if (Number.isNaN(hand)) return { error: 'Invalid hand index' };
    return { args: ['p', { handIndex: hand }] };
  }

  if (cmd === 'hint' || cmd === 'h') return { args: ['hint'] };

  return { error: `Unknown command: ${cmd}` };
}

/** Contract display names indexed by contract id. */
const CONTRACT_NAMES = ['No Tricks', 'No Hearts', 'No Queens', 'Barbu', 'No Last Trick', 'Trumps', 'Dominoes'];

/** Renders a Barbu game state as CLI-friendly text. */
export function formatBarbuState(s: BarbuResponse): string {
  const lines: string[] = [];
  lines.push(`Deal ${s.dealNumber + 1}/${s.totalDeals} / Dealer: ${s.dealerIdx} / Phase: ${s.phase}`);
  if (s.currentContract >= 0) {
    lines.push(`Contract: ${CONTRACT_NAMES[s.currentContract] ?? '?'}`);
  }
  lines.push(`Turn: ${s.gameEndFlag ? 'End' : `Player ${s.currentTurn}`}`);
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : `CPU${p.id}`;
    lines.push(`${tag}: hand=${p.cardCount} tricks=${p.trickCount} score=${p.totalScore}`);
  }
  if (s.currentTrick.length > 0) {
    lines.push(`Trick: ${s.currentTrick.map((t) => `${t.card.value}${t.card.design}`).join(' ')}`);
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for Barbu. */
export const BARBU_HELP = [
  'c <0-6> [trump]  - Pick a contract (trump 1-4 only for Trumps=5)',
  'p <hand>         - Play hand card (-1 to pass in Dominoes)',
  'n/next           - Start next deal',
  'r/reset          - Reset game',
  'l/log            - Show action log',
  'h/hint           - Get a hint',
] as const;
