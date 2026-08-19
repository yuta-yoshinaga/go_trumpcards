import { describe, expect, it } from 'vitest';
import { TWENTYNINE_TOTAL_POINTS, twentyNineContractProgress } from './twentyNineContract';

describe('twentyNineContractProgress', () => {
  it('is null until a contract exists', () => {
    expect(twentyNineContractProgress(-1, 20, [0, 0])).toBeNull();
    expect(twentyNineContractProgress(0, 0, [0, 0])).toBeNull();
  });

  it('reports made once the declaring team reaches the contract', () => {
    expect(twentyNineContractProgress(0, 16, [16, 5])).toEqual({
      declarerTeam: 0,
      points: 16,
      contract: 16,
      remaining: 0,
      status: 'made',
    });
  });

  it('reads the declaring team, not team 0', () => {
    const p = twentyNineContractProgress(1, 16, [5, 16]);
    expect(p?.declarerTeam).toBe(1);
    expect(p?.status).toBe('made');
  });

  // 場に残るのは 29-(10+15)=4 点。10+4=14 < 20 なので、この時点で不成立が確定する。
  it('reports failed once the points still in play cannot close the gap', () => {
    const p = twentyNineContractProgress(0, 20, [10, 15]);
    expect(p?.status).toBe('failed');
    expect(p?.remaining).toBe(10);
  });

  // ちょうど届く境界: 残り 4 点で 16+4 = 20。まだ failed にしてはいけない。
  it('keeps needMore when the remaining points exactly close the gap', () => {
    expect(twentyNineContractProgress(0, 20, [16, 9])?.status).toBe('needMore');
  });

  it('sums to the game its name comes from', () => {
    expect(TWENTYNINE_TOTAL_POINTS).toBe(29);
  });
});
