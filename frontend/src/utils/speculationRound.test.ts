import { describe, expect, it } from 'vitest';
import { SpeculationPhase } from '../types/phases';
import { speculationDisplayRound } from './speculationRound';

describe('speculationDisplayRound', () => {
  // **決着した回は、その回の番号を出す。** `roundNo` は決着で 1 進むので、
  // そのまま +1 すると「まだ始めていない回の結果」を見せることになる (#6607)。
  it('shows the round that just finished on the result and game-end screens', () => {
    expect(speculationDisplayRound(SpeculationPhase.RESULT, 1)).toBe(1);
    expect(speculationDisplayRound(SpeculationPhase.GAME_END, 5)).toBe(5);
  });

  // 進行中は次に打つ回を出す。ここを落とすと全画面が 1 つ手前になる。
  it('shows the round about to be played in every other phase', () => {
    expect(speculationDisplayRound(SpeculationPhase.FLIP, 0)).toBe(1);
    expect(speculationDisplayRound(SpeculationPhase.AUCTION, 2)).toBe(3);
    expect(speculationDisplayRound(undefined, 0)).toBe(1);
  });
});
