// API client for alaska. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AlaskaResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Russian Solitaire card move. */
export interface AlaskaMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Russian Solitaire /alaska/exec endpoint. */
export const alaskaHintApi = createSolitaireMoveApi<
  AlaskaResponse,
  AlaskaMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('alaska');
