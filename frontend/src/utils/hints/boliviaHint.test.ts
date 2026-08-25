import { describe, expect, it } from 'vitest';
import { makeBoliviaState } from '../../test/stateFactories';
import { BoliviaPhase } from '../../types/phases';
import { getBoliviaHint } from './boliviaHint';

const meldPhase = (overrides: Record<string, unknown> = {}) => {
  const base = makeBoliviaState();
  return makeBoliviaState({
    phase: BoliviaPhase.MELD,
    currentPlayerIdx: 0,
    players: base.players.map((p, i) => (i === 0 ? { ...p, hasInitMeld: true, ...overrides } : p)),
  });
};

describe('getBoliviaHint', () => {
  // **上がりに要るのはエスカレラ。** ここを `hasBolivia` で見ていると、
  // 「ボリビアはあるがエスカレラが無い」──いちばん助言が要る局面──で黙る。
  it('tells a player with a bolivia but no escalera to build the escalera', () => {
    const hint = getBoliviaHint(meldPhase({ hasBolivia: true, hasEscalera: false }));
    expect(hint?.reason).toBe('hint.buildEscalera');
    expect(hint?.confidence).toBe('strong');
  });

  // 負のコントロール: エスカレラを持っている人に「エスカレラを作れ」と言わない。
  it('does not tell a player who already has an escalera to build one', () => {
    const hint = getBoliviaHint(meldPhase({ hasEscalera: true, hasBolivia: false }));
    expect(hint?.reason).not.toBe('hint.buildEscalera');
    expect(hint?.reason).toBe('hint.buildBolivia');
  });

  it('suggests extending once both signature melds are down', () => {
    expect(getBoliviaHint(meldPhase({ hasEscalera: true, hasBolivia: true }))?.reason).toBe('hint.meldExtend');
  });

  it('asks for the initial meld before anything else', () => {
    const base = makeBoliviaState();
    const state = makeBoliviaState({
      phase: BoliviaPhase.MELD,
      currentPlayerIdx: 0,
      players: base.players.map((p, i) => (i === 0 ? { ...p, hasInitMeld: false, hasEscalera: false } : p)),
    });
    expect(getBoliviaHint(state)?.reason).toBe('hint.meldInitial');
  });

  it.each([
    ['the game has ended', { gameEndFlag: true }],
    ['it is not the human turn', { currentPlayerIdx: 1 }],
  ])('is silent when %s', (_label, overrides) => {
    expect(getBoliviaHint(makeBoliviaState(overrides))).toBeNull();
  });

  it('advises a draw in the draw phase and a discard in the discard phase', () => {
    const base = makeBoliviaState();
    const at = (phase: number) => makeBoliviaState({ phase, currentPlayerIdx: 0, players: base.players });
    expect(at(BoliviaPhase.DRAW)?.phase).toBe(BoliviaPhase.DRAW);
    expect(getBoliviaHint(at(BoliviaPhase.DRAW))?.targetAction).toBe('draw');
    expect(getBoliviaHint(at(BoliviaPhase.DISCARD))?.targetAction).toBe('discard');
  });
});
