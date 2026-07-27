// API client for courtpiece. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CourtPieceResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Court Piece (Rang) game settings. */
export interface CourtPieceConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Commands accepted by the Court Piece (Rang) /courtpiece/exec endpoint. */
export type CourtPieceCommand = 'reset' | 'trump' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Court Piece (Rang) /courtpiece/exec endpoint.
 *
 * Court Piece is a 4-player, 2-team (seats 0&2 vs 1&3) trick-taker with no
 * numeric bidding. The caller (Hakim) peeks at 5 cards and declares a trump
 * suit; the teams then play 13 tricks. A team taking 7+ tricks wins the round
 * (Sar = +1 point); sweeps and consecutive wins add a Court bonus (+2). The
 * first team to reach the point limit (default 7) wins.
 *   - `trump` → `{ trumpSuit: number }` (1=♠ 2=♣ 3=♥ 4=♦)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const courtPieceApi = {
  exec: (
    command: CourtPieceCommand,
    opts?: {
      trumpSuit?: number;
      cardIndex?: number;
      config?: CourtPieceConfigInput;
    },
  ) =>
    gameExec<CourtPieceResponse>('courtpiece', {
      command,
      trumpSuit: opts?.trumpSuit,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
