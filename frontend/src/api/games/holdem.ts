// API client for holdem. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HoldemResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Texas Hold'em game settings. */
export interface HoldemConfigInput {
  smallBlind?: number;
  bigBlind?: number;
  tournamentMode?: boolean;
  blindLevelHands?: number;
  blindMultiplier?: number;
  bettingLimit?: number;
  tableSize?: number;
  rebuyEnabled?: boolean;
  rebuyMaxCount?: number;
  rebuyChips?: number;
  rebuyPeriodHands?: number;
  addonEnabled?: boolean;
  addonChips?: number;
  addonAfterHand?: number;
  cpuMetaAI?: boolean;
}

/** Command set shared by Hold'em-family games. */
export type HoldemLikeCommand =
  | 'reset'
  | 'fold'
  | 'check'
  | 'call'
  | 'bet'
  | 'raise'
  | 'allin'
  | 'rebuy'
  | 'skiprebuy'
  | 'addon'
  | 'skipaddon'
  | 'muck'
  | 'show';

/** Factory for Hold'em-family APIs that share the same exec pattern. */
export function createHoldemLikeApi<T, C = HoldemConfigInput>(game: string) {
  return {
    exec: (command: HoldemLikeCommand, amount?: number, config?: C, humanPlayMs?: number, profile?: unknown) =>
      gameExec<T>(game, { command, amount, humanPlayMs, profile, ...(config as Record<string, unknown>) }),
  };
}

/** The exec signature shared by every `createHoldemLikeApi` client (issue #4301). */
export type HoldemLikeExec<T> = ReturnType<typeof createHoldemLikeApi<T>>['exec'];

/** API client for the Texas Hold'em /holdem/exec endpoint. */
export const holdemApi = createHoldemLikeApi<HoldemResponse>('holdem');
