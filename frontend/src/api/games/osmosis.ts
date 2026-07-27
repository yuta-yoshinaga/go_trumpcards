// API client for osmosis. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OsmosisResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for an Osmosis card move. */
export interface OsmosisMoveZone {
  zone: string;
  col?: number;
}

/** API client for the Osmosis /osmosis/exec endpoint. */
export const osmosisApi = createSolitaireMoveApi<
  OsmosisResponse,
  OsmosisMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('osmosis');
