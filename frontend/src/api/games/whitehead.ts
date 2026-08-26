// API client for whitehead. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WhiteheadResponse } from '../../types/card';
import { createSolitaireMoveApiWithConfig } from '../gameExec';

/** Source or target zone for a Whitehead card move. */
export interface WhiteheadMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** Configuration options for Whitehead game settings. */
export interface WhiteheadConfigInput {
  drawCount?: number;
  scoringMode?: number;
}

/** API client for the Whitehead /whitehead/exec endpoint. */
export const whiteheadApi = createSolitaireMoveApiWithConfig<
  WhiteheadResponse,
  WhiteheadMoveZone,
  WhiteheadConfigInput,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('whitehead');
