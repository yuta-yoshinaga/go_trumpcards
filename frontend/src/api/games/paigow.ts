// API client for paigow. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PaiGowResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Pai Gow Poker /paigow/exec endpoint. */
export const paigowApi = {
  exec: (command: 'reset' | 'bet' | 'set' | 'log' | 'hint', amount?: number, low0?: number, low1?: number) =>
    gameExec<PaiGowResponse>('paigow', { command, amount, low0, low1 }),
};
