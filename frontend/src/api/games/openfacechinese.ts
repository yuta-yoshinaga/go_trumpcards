// API client for openfacechinese. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OpenFaceChineseResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the Open Face Chinese Poker (OFC) /openfacechinese/exec endpoint. */
export type OpenFaceChineseCommand = 'reset' | 'place' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Open Face Chinese Poker (OFC) /openfacechinese/exec endpoint.
 *
 * OFC is a solo-vs-CPU game where each dealt card must be committed to one of
 * three rows — Top (3 cards), Middle (5 cards) or Bottom (5 cards) — and once
 * placed a card cannot be moved. Commands:
 *   - `place` -> `{ row }` where row is 0=Top, 1=Middle, 2=Bottom
 *   - `reset` / `nextround` / `hint` / `log` carry no extra fields
 */
export const openfacechineseApi = {
  exec: (command: OpenFaceChineseCommand, opts?: { row?: number }) =>
    gameExec<OpenFaceChineseResponse>('openfacechinese', {
      command,
      row: opts?.row,
    }),
};
