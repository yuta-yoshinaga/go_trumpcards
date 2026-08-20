import { describe, expect, it } from 'vitest';
import type { EgyptianRatscrewResponse } from '../../../types/card';
import { EgyptianRatscrewPhase } from '../../../types/phases';
import { formatEgyptianRatscrewState } from './egyptianratscrewFormatter';

function baseState(overrides: Partial<EgyptianRatscrewResponse> = {}): EgyptianRatscrewResponse {
  return {
    phase: EgyptianRatscrewPhase.PLAY,
    gameEndFlag: false,
    winnerIdx: -1,
    currentTurnIdx: 0,
    isHumanTurn: true,
    isTopFaceCard: false,
    isSlappable: false,
    centerPileSize: 0,
    topCard: null,
    players: [
      { name: 'You', isHuman: true, stockSize: 26 },
      { name: 'CPU', isHuman: false, stockSize: 26 },
    ],
    cpuDifficulty: 1,
    chanceRemaining: 0,
    faceChances: { jack: 1, queen: 2, king: 3, ace: 4 },
    chanceFromIdx: -1,
    pendingKind: 0,
    pendingDeadlineMs: 0,
    lastEventKind: 0,
    lastEventPlayerIdx: 0,
    lastSlapReason: 0,
    message: '',
    ...overrides,
  };
}

describe('formatEgyptianRatscrewState', () => {
  it('renders Play phase, empty pile, no top card, no slap, no chance, no message', () => {
    const out = formatEgyptianRatscrewState(baseState());
    expect(out).toContain('Phase: Play');
    expect(out).toContain('Pile: 0');
    expect(out).toContain('Top: --');
    expect(out).not.toContain('[SLAP!]');
    expect(out).not.toContain('Chance:');
    expect(out).toContain('Turn: P0');
    expect(out).toContain('You: stock=26');
    expect(out).toContain('CPU: stock=26');
  });

  it('renders End phase when game ended', () => {
    const out = formatEgyptianRatscrewState(baseState({ phase: EgyptianRatscrewPhase.GAME_END, gameEndFlag: true }));
    expect(out).toContain('Phase: End');
  });

  it('renders top card value when present', () => {
    const out = formatEgyptianRatscrewState(baseState({ centerPileSize: 1, topCard: { design: 'SPADE', value: 11 } }));
    expect(out).toContain('Top: 11');
    expect(out).toContain('Pile: 1');
  });

  it('renders SLAP marker when slappable', () => {
    const out = formatEgyptianRatscrewState(
      baseState({ isSlappable: true, centerPileSize: 2, topCard: { design: 'HEART', value: 7 } }),
    );
    expect(out).toContain('[SLAP!]');
  });

  it('renders Chance count during a chance battle', () => {
    const out = formatEgyptianRatscrewState(baseState({ chanceRemaining: 3 }));
    expect(out).toContain('Chance: 3');
  });

  it('appends message line when message is non-empty', () => {
    const out = formatEgyptianRatscrewState(baseState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  it('omits message line when message is empty', () => {
    const out = formatEgyptianRatscrewState(baseState({ message: '' }));
    expect(out.split('\n').filter((l) => l === '').length).toBe(0);
  });
});
