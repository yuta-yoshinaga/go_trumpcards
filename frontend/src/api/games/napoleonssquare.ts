// API client for napoleonssquare. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { NapoleonsSquareMoveZone, NapoleonsSquareResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Napoleon's Square /napoleonssquare/exec endpoint. */
export const napoleonsSquareApi = createSolitaireMoveApi<
  NapoleonsSquareResponse,
  NapoleonsSquareMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('napoleonssquare');
