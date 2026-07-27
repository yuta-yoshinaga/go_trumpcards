// API client for flowergarden. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FlowerGardenMoveZone, FlowerGardenResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Flower Garden /flowergarden/exec endpoint. */
export const flowerGardenApi = createSolitaireMoveApi<
  FlowerGardenResponse,
  FlowerGardenMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('flowergarden');
