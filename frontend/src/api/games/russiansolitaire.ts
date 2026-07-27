// API client for russiansolitaire. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RussianSolitaireResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Russian Solitaire card move. */
export interface RussianSolitaireMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Russian Solitaire /russiansolitaire/exec endpoint. */
export const russianSolitaireApi = createSolitaireMoveApi<
  RussianSolitaireResponse,
  RussianSolitaireMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('russiansolitaire');
