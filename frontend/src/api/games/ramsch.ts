// API client for ramsch. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RamschConfig as RamschConfigType, RamschResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Ramsch game settings. */
export interface RamschConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** API client for the Ramsch /ramsch/exec endpoint. */
export const ramschApi = {
  exec: (
    // **入札系のコマンドは無い。** Skat クローンの 'bid' / 'pickramsch' /
    // 'discard' / 'game' はバックエンドの dispatcher にも CLI のパーサにも無く、
    // 残すと型は通るのに実行時に unknown command になる（OpenAPI の
    // RamschRequest とも食い違う）。
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    args?: {
      cardIndex?: number;
      config?: RamschConfigInput;
    },
  ) =>
    gameExec<RamschResponse>('ramsch', {
      command,
      ...(args || {}),
    }),
};

// RamschConfigType import is used only for type re-export; ensure it's referenced.
export type { RamschConfigType };
