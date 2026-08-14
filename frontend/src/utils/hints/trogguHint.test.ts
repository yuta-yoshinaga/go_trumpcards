import { describe, expect, it } from 'vitest';
import { makeTrogguState } from '../../test/stateFactories';
import { getTrogguHint } from './trogguHint';

describe('getTrogguHint', () => {
  it('advises nothing once the match is over', () => {
    expect(getTrogguHint(makeTrogguState({ gameEndFlag: true }))).toBeNull();
  });

  it('advises advancing at a trick or deal boundary', () => {
    expect(getTrogguHint(makeTrogguState({ phase: 2 }))?.targetAction).toBe('next');
    expect(getTrogguHint(makeTrogguState({ phase: 3 }))?.targetAction).toBe('nextround');
  });

  it('advises nothing while a CPU is thinking', () => {
    expect(getTrogguHint(makeTrogguState({ phase: 1, isHumanTurn: false }))).toBeNull();
  });

  it('advises on the auction', () => {
    expect(getTrogguHint(makeTrogguState({ phase: 0 }))?.reason).toBe('frontendHint.trogguBidPickTheContract');
  });

  // **取ってはいけない契約は、デクレアラーにだけ効く。** 防御側は取りにいってよい。
  it('flips the advice only for the declarer of a negative contract', () => {
    const base = makeTrogguState({ phase: 1, contractName: 'misere' });
    const asDeclarer = makeTrogguState({
      ...base,
      players: base.players.map((p, i) => (i === 0 ? { ...p, isDeclarer: true } : p)),
    });
    expect(getTrogguHint(asDeclarer)?.reason).toBe('frontendHint.trogguAvoidTricks');

    // 同じミゼールでも、人間が防御側なら取りにいく助言のまま。
    const asDefender = makeTrogguState({
      ...base,
      players: base.players.map((p, i) => (i === 1 ? { ...p, isDeclarer: true } : p)),
    });
    expect(getTrogguHint(asDefender)?.reason).toBe('frontendHint.trogguFollowSuit');
  });

  it('does not flip for a positive contract', () => {
    const solo = makeTrogguState({
      phase: 1,
      contractName: 'solo',
      players: makeTrogguState().players.map((p, i) => (i === 0 ? { ...p, isDeclarer: true } : p)),
    });
    expect(getTrogguHint(solo)?.reason).toBe('frontendHint.trogguFollowSuit');
  });
});
