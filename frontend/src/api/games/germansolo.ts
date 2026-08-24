// API client for germansolo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GermanSoloResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for German Solo game settings. */
export interface GermanSoloConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the German Solo /germansolo/exec endpoint. */
export type GermanSoloCommand = 'reset' | 'bid' | 'ace' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the German Solo /germansolo/exec endpoint.
 *
 * German Solo is a 4-player soloist-vs-defenders trick-taker on a 32-card Skat
 * pack. The auction climbs Frage → Solo → Tout; the two partner contracts then
 * enter an ace-call phase where the declarer names the ace that picks a hidden
 * partner.
 *   - `bid` → `{ bid, trumpSuit? }` (bid 0=pass, 2=Frage, 3=Solo, 4=Tout —
 *     **1 (Mussfrage) is never sent**, it is forced by the server when everyone
 *     passes; trumpSuit 1=♠ 2=♣ 3=♥ 4=♦, required for any bid but a pass)
 *   - `ace` → `{ aceSuit }` (1=♠ 2=♣ 3=♥ 4=♦; must be an ace the declarer does
 *     not hold and not of the trump suit)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const germansoloApi = {
  exec: (
    command: GermanSoloCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      trumpSuit?: number;
      aceSuit?: number;
      config?: GermanSoloConfigInput;
    },
  ) =>
    gameExec<GermanSoloResponse>('germansolo', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      trumpSuit: opts?.trumpSuit,
      aceSuit: opts?.aceSuit,
      config: opts?.config,
    }),
};
