import { describe, expect, it } from 'vitest';
import type { SlapjackResponse } from '../../../types/card';
import { SlapjackPhase } from '../../../types/phases';
import { formatSlapjackState } from './slapjackFormatter';

function makeState(overrides: Partial<SlapjackResponse> = {}): SlapjackResponse {
  return {
    phase: SlapjackPhase.PLAY,
    centerPileSize: 3,
    topCard: { value: 11 } as SlapjackResponse['topCard'],
    currentTurnIdx: 0,
    players: [
      { isHuman: true, stockSize: 20 },
      { isHuman: false, stockSize: 32 },
    ],
    message: '',
    ...overrides,
  } as SlapjackResponse;
}

describe('formatSlapjackState', () => {
  it('renders localized phase, pile summary, and player stock rows', () => {
    const out = formatSlapjackState(makeState());
    expect(out).toContain('フェーズ: プレイ中 | 場札: 3 | トップ: 11 | 手番: P0');
    expect(out).toContain('あなた: ストック=20');
    expect(out).toContain('CPU: ストック=32');
  });

  it('shows End phase and a placeholder top when the pile is empty', () => {
    const out = formatSlapjackState(makeState({ phase: SlapjackPhase.GAME_END, topCard: null }));
    expect(out).toContain('フェーズ: 終了');
    expect(out).toContain('トップ: --');
  });

  it('appends any message', () => {
    const out = formatSlapjackState(makeState({ message: 'テストメッセージ' }));
    expect(out).toContain('テストメッセージ');
  });
});
