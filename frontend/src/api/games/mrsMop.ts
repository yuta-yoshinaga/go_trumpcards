// API client for mrsmop. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MrsMopResponse } from '../../types/card';
import { createSolitaireMoveApiWithConfig } from '../gameExec';

/** Source or target zone for a MrsMop card move. */
export interface MrsMopMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** Configuration options for MrsMop game settings. */
export interface MrsMopConfigInput {
  difficulty?: number;
}

/** API client for the MrsMop /mrsMop/exec endpoint. */
export const mrsMopApi = createSolitaireMoveApiWithConfig<
  MrsMopResponse,
  MrsMopMoveZone,
  MrsMopConfigInput,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('mrsmop');
