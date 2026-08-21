// API client for mao. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type {
  BakersDozenMoveZone,
  BeleagueredCastleMoveZone,
  CrescentMoveZone,
  FlowerGardenMoveZone,
  FortressMoveZone,
  FortyAndEightMoveZone,
  FortyThievesMoveZone,
  KingAlbertMoveZone,
  MaoResponse,
  PerseveranceMoveZone,
  RankAndFileMoveZone,
  SomersetMoveZone,
  StHelenaMoveZone,
  StreetsAndAlleysMoveZone,
  SultanMoveZone,
} from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Mao game settings. */
export interface MaoConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/**
 * API client for the Mao /mao/exec endpoint.
 *
 * Mao is a Crazy Eights-style shedding game with a secret hidden rule. The
 * `declareword` command carries the player's compliance utterance (`word`);
 * the server never reveals the rule itself.
 */
export const maoApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'suit' | 'declare' | 'skipdeclare' | 'declareword' | 'nextround',
    cardIndex?: number,
    suit?: number,
    config?: MaoConfigInput,
    word?: string,
  ) =>
    gameExec<MaoResponse>('mao', {
      command,
      cardIndex,
      suit,
      config,
      word,
    }),
};

export type {
  BakersDozenMoveZone,
  BeleagueredCastleMoveZone,
  CrescentMoveZone,
  FlowerGardenMoveZone,
  FortressMoveZone,
  FortyAndEightMoveZone,
  FortyThievesMoveZone,
  KingAlbertMoveZone,
  PerseveranceMoveZone,
  RankAndFileMoveZone,
  SomersetMoveZone,
  StHelenaMoveZone,
  StreetsAndAlleysMoveZone,
  SultanMoveZone,
};
