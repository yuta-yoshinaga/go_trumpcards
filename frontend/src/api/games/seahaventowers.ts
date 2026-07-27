// API client for seahaventowers. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SeahavenTowersResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Seahaven Towers card move. */
export interface SeahavenTowersMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Seahaven Towers /seahaventowers/exec endpoint. */
export const seahaventowersApi = createSolitaireMoveApi<
  SeahavenTowersResponse,
  SeahavenTowersMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('seahaventowers');
