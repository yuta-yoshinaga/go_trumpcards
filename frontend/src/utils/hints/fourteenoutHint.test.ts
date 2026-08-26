import { describe, expect, it } from 'vitest';
import type { CardDesign, FourteenOutBoardCell, FourteenOutResponse } from '../../types/card';
import { FourteenOutPhase } from '../../types/phases';
import { getFourteenOutHint } from './fourteenoutHint';

const card = (design: CardDesign, value: number) => ({ design, value });

/** 12 列の state を作る。指定した列だけ中身を持つ。 */
function makeState(cols: number[][], overrides: Partial<FourteenOutResponse> = {}): FourteenOutResponse {
  const columns: FourteenOutBoardCell[][] = Array.from({ length: 12 }, (_, i) =>
    (cols[i] ?? []).map((v) => ({ card: card('SPADE', v) })),
  );
  return {
    columns,
    phase: FourteenOutPhase.PLAYING,
    removedCount: 0,
    removablePairs: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getFourteenOutHint', () => {
  it('names the two columns whose tails make 14', () => {
    const hint = getFourteenOutHint(makeState([[9], [5]]));
    expect(hint?.targetAction).toBe('remove-0-1');
    expect(hint?.confidence).toBe('strong');
  });

  it('scans in the same order as the backend (lowest pair first)', () => {
    // 9 / 5 / 5 → (0,1) が最初。
    expect(getFourteenOutHint(makeState([[9], [5], [5]]))?.targetAction).toBe('remove-0-1');
  });

  // **クローン元は隣接を要求する。**離れた列でも見つけること。
  it('finds a pair in distant columns', () => {
    const cols: number[][] = Array.from({ length: 12 }, () => []);
    cols[0] = [9];
    cols[11] = [5];
    expect(getFourteenOutHint(makeState(cols))?.targetAction).toBe('remove-0-11');
  });

  // **末尾しか見ない。**埋もれた札で組めても提案しない。
  it('ignores buried cards', () => {
    expect(getFourteenOutHint(makeState([[9, 2], [5]]))).toBeNull();
  });

  // **"deal" は返さない。**クローン元は手が無いとき deal を勧めたが、
  // Fourteen Out に山札は無いので、勧める手は存在しない。
  it('returns null when nothing pairs, never a deal suggestion', () => {
    expect(getFourteenOutHint(makeState([[2], [3], [4]]))).toBeNull();
  });

  it('returns null once the game has ended', () => {
    for (const phase of [FourteenOutPhase.GAME_CLEAR, FourteenOutPhase.GAME_OVER]) {
      expect(getFourteenOutHint(makeState([[9], [5]], { phase }))).toBeNull();
    }
  });
});
