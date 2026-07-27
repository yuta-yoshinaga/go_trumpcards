// API client for king. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KingResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for King game settings. */
export interface KingConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the King /king/exec endpoint. */
export type KingCommand = 'reset' | 'contract' | 'play' | 'next' | 'hint' | 'log';

/**
 * API client for the King /king/exec endpoint.
 *
 * King is a 4-player 52-card compendium trick-avoidance game. Each match runs
 * exactly seven deals; the dealer of each deal selects one of seven unused
 * contracts and all four seats play thirteen must-follow tricks.
 *   - `contract` → `{ contract, trumpSuit }` (the dealer picks the deal's
 *     contract 0..6; `trumpSuit` is 1..4 for contract 6 "King (Trump)", else -1)
 *   - `play` → `{ handIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `hint` / `log` carry no extra fields.
 */
export const kingApi = {
  exec: (
    command: KingCommand,
    opts?: {
      contract?: number;
      trumpSuit?: number;
      handIndex?: number;
      config?: KingConfigInput;
    },
  ) =>
    gameExec<KingResponse>('king', {
      command,
      contract: opts?.contract,
      trumpSuit: opts?.trumpSuit,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};
