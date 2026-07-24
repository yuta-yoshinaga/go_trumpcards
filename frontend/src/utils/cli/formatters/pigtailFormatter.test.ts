import { describe, expect, it } from 'vitest';
import type { PigsTailResponse } from '../../../types/card';
import { formatPigtailState } from './pigtailFormatter';

function makeState(overrides: Partial<PigsTailResponse> = {}): PigsTailResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 3 },
      { id: 1, isHuman: false, cardCount: 5 },
    ],
    circleCount: 40,
    centerCount: 4,
    gameEndFlag: false,
    lastDrawCard: null,
    lastPenalty: false,
    message: '',
    ...overrides,
  } as PigsTailResponse;
}

describe('formatPigtailState', () => {
  it('renders localized phase, circle/center summary, and player rows', () => {
    const out = formatPigtailState(makeState());
    expect(out).toContain('フェーズ: プレイ中 | 輪: 40 | 中央: 4');
    expect(out).toContain('あなた: 3枚');
    expect(out).toContain('CPU 1: 5枚');
  });

  it('shows End phase when the game has ended', () => {
    const out = formatPigtailState(makeState({ gameEndFlag: true }));
    expect(out).toContain('フェーズ: 終了');
  });

  it('renders a penalty draw line and any message', () => {
    const out = formatPigtailState(
      makeState({
        lastDrawCard: { design: 'HEART', value: 7 } as PigsTailResponse['lastDrawCard'],
        lastPenalty: true,
        message: 'テストメッセージ',
      }),
    );
    expect(out).toContain('最後: H7 （ペナルティ）');
    expect(out).toContain('テストメッセージ');
  });

  it('marks a safe draw', () => {
    const out = formatPigtailState(
      makeState({ lastDrawCard: { design: 'SPADE', value: 2 } as PigsTailResponse['lastDrawCard'] }),
    );
    expect(out).toContain('（セーフ）');
  });
});
