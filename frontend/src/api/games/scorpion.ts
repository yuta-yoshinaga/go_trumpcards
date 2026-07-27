// API client for scorpion. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ScorpionResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Scorpion card move. */
export interface ScorpionMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Scorpion /scorpion/exec endpoint. */
export const scorpionApi = createSolitaireMoveApi<
  ScorpionResponse,
  ScorpionMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('scorpion');
