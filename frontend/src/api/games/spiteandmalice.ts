// API client for spiteandmalice. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpiteAndMaliceMoveZone, SpiteAndMaliceResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Command verbs accepted by the Spite & Malice /spiteandmalice/exec endpoint. */
export type SpiteAndMaliceCommand = 'reset' | 'move' | 'discard' | 'cpu' | 'autocomplete' | 'hint' | 'log';

/** API client for the Spite & Malice /spiteandmalice/exec endpoint. */
export const spiteAndMaliceApi = {
  exec: (command: SpiteAndMaliceCommand, from?: SpiteAndMaliceMoveZone, to?: SpiteAndMaliceMoveZone) =>
    gameExec<SpiteAndMaliceResponse>('spiteandmalice', { command, from, to }),
};
