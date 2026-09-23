import { describe, expect, it, vi } from 'vitest';
import { makeTappTarockState } from '../../test/stateFactories';
import { getTappTarockHint } from './tapptarockHint';

vi.mock('../../types/phases', () => ({
  TappTarockPhase: { BID: 0, TALON: 1, PLAY: 2, TRICK_END: 3, ROUND_END: 4, GAME_END: 5 },
}));

describe('getTappTarockHint', () => {
  it('advises nothing once the match is over', () => {
    expect(getTappTarockHint(makeTappTarockState({ gameEndFlag: true }))).toBeNull();
  });

  it('advises advancing at a trick or deal boundary', () => {
    expect(getTappTarockHint(makeTappTarockState({ phase: 3 }))?.targetAction).toBe('next');
    expect(getTappTarockHint(makeTappTarockState({ phase: 4 }))?.targetAction).toBe('nextround');
  });

  it('advises nothing while a CPU is thinking', () => {
    expect(getTappTarockHint(makeTappTarockState({ phase: 2, isHumanTurn: false }))).toBeNull();
  });

  it('advises on the auction and the talon exchange', () => {
    expect(getTappTarockHint(makeTappTarockState({ phase: 0 }))?.reason).toBe('frontendHint.tapptarockBidNeedsTrumps');
    expect(getTappTarockHint(makeTappTarockState({ phase: 1 }))?.reason).toBe('frontendHint.tapptarockBuryCheap');
  });

  // **トリシャーケンだけ助言が逆向き。** 点を取ると負けるので、勝ちにいかせない。
  it('flips the advice under Trischaken', () => {
    const trischaken = makeTappTarockState({ phase: 2, contractName: 'trischaken' });
    expect(getTappTarockHint(trischaken)?.reason).toBe('frontendHint.tapptarockAvoidPoints');

    const rufer = makeTappTarockState({ phase: 2, contractName: 'rufer' });
    expect(getTappTarockHint(rufer)?.reason).toBe('frontendHint.tapptarockFollowSuit');
  });
});
