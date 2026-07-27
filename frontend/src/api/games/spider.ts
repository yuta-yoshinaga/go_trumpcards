// API client for spider. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpiderResponse } from '../../types/card';
import { createSolitaireMoveApiWithConfig } from '../gameExec';

/** Source or target zone for a Spider card move. */
export interface SpiderMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** Configuration options for Spider game settings. */
export interface SpiderConfigInput {
  difficulty?: number;
}

/** API client for the Spider /spider/exec endpoint. */
export const spiderApi = createSolitaireMoveApiWithConfig<
  SpiderResponse,
  SpiderMoveZone,
  SpiderConfigInput,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('spider');
