import { describe, expect, it } from 'vitest';
import { makeSambaState } from '../../test/stateFactories';
import type { SambaPlayerData } from '../../types/card';
import { SambaPhase } from '../../types/phases';
import { getSambaHint } from './sambaHint';

function withHuman(overrides: Partial<SambaPlayerData>) {
  const base = makeSambaState();
  return makeSambaState({ players: [{ ...base.players[0], ...overrides }, ...base.players.slice(1)] });
}

describe('getSambaHint', () => {
  it('returns null when game has ended', () => {
    expect(getSambaHint(makeSambaState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getSambaHint(makeSambaState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends drawing from stock at start of turn', () => {
    expect(getSambaHint(makeSambaState({ discardTop: null, discardPileCount: 0 }))?.reason).toBe('hint.drawStock');
  });

  it('recommends taking the discard pile when the human has melded and it is not frozen', () => {
    const state = withHuman({ hasInitMeld: true });
    expect(getSambaHint({ ...state, discardPileCount: 3 })?.reason).toBe('hint.takeDiscardPile');
  });

  it('recommends the initial meld in the meld phase when the human has none', () => {
    const hint = getSambaHint(makeSambaState({ phase: SambaPhase.MELD }));
    expect(hint?.reason).toBe('hint.meldInitial');
  });

  it('recommends building a samba once melded but without a samba yet', () => {
    const state = withHuman({ hasInitMeld: true });
    expect(getSambaHint({ ...state, phase: SambaPhase.MELD })?.reason).toBe('hint.buildSamba');
  });

  it('recommends extending melds once a samba exists', () => {
    const state = withHuman({ hasInitMeld: true, hasSamba: true });
    expect(getSambaHint({ ...state, phase: SambaPhase.MELD })?.reason).toBe('hint.meldExtend');
  });

  it('recommends discarding a high safe card in the discard phase', () => {
    const hint = getSambaHint(makeSambaState({ phase: SambaPhase.DISCARD }));
    expect(hint?.reason).toBe('hint.discardHighSafe');
  });
});
