import { describe, expect, it } from 'vitest';
import { makeHorseState } from '../../test/stateFactories';
import { getHorseHint } from './horseHint';

describe('getHorseHint', () => {
  it('says nothing once the match is over', () => {
    expect(getHorseHint(makeHorseState({ gameEndFlag: true }))).toBeNull();
  });

  it('points at the next hand when the hand is settled', () => {
    const hint = getHorseHint(makeHorseState({ phase: 1, isHumanTurn: false }));
    expect(hint).toEqual({ targetAction: 'next', reason: 'frontendHint.horseNextHand', confidence: 'strong' });
  });

  it('says nothing while a CPU is acting', () => {
    expect(getHorseHint(makeHorseState({ isHumanTurn: false }))).toBeNull();
  });

  // **種目が変わった直後だけ強く告げる。** 同じ操作で規則だけが変わるのが
  // ミックスゲームの唯一の落とし穴。
  it('warns on the first hand of a discipline', () => {
    const hint = getHorseHint(makeHorseState({ handInDiscipline: 1 }));
    expect(hint?.reason).toBe('frontendHint.horseDisciplineChanged');
    expect(hint?.confidence).toBe('strong');
  });

  it('recalls the two-card rule in Omaha', () => {
    const hint = getHorseHint(
      makeHorseState({ handInDiscipline: 2, discipline: 1, disciplineName: 'omahaHiLo', communityCards: [] }),
    );
    expect(hint?.reason).toBe('frontendHint.horseOmahaUsesTwo');
  });

  it('recalls that Razz wants a low hand', () => {
    const hint = getHorseHint(makeHorseState({ handInDiscipline: 2, discipline: 2, disciplineName: 'razz' }));
    expect(hint?.reason).toBe('frontendHint.horseRazzWantsLow');
  });

  it('falls back to the running discipline', () => {
    const hint = getHorseHint(makeHorseState({ handInDiscipline: 2 }));
    expect(hint?.reason).toBe('frontendHint.horsePlayTheDiscipline');
  });
});
