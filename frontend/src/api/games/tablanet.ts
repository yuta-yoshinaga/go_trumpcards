// API client for tablanet. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TablanetResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tablanet (Tablić) game settings (CPU difficulty only). */
export interface TablanetConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Tablanet /tablanet/exec endpoint. */
export type TablanetCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tablanet (Tablić) /tablanet/exec endpoint.
 *
 * Tablanet is a 4-player 52-card fishing/capture game.
 *   - `play` → `{ cardIndex, tableIndices? }` (tableIndices = table cards to
 *     capture; omit to trail, a Jack always sweeps)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tablanetApi = {
  exec: (
    command: TablanetCommand,
    opts?: {
      cardIndex?: number;
      tableIndices?: number[];
      config?: TablanetConfigInput;
    },
  ) =>
    gameExec<TablanetResponse>('tablanet', {
      command,
      cardIndex: opts?.cardIndex,
      tableIndices: opts?.tableIndices,
      config: opts?.config,
    }),
};
