import { describe, expect, it } from 'vitest';
import { makeSakuraState } from '../../test/stateFactories';
import { getSakuraHint } from './sakuraHint';

describe('getSakuraHint', () => {
  it('advises nothing once the game is over', () => {
    expect(getSakuraHint(makeSakuraState({ gameEndFlag: true }))).toBeNull();
  });

  it('advises moving on when the round is over', () => {
    const hint = getSakuraHint(makeSakuraState({ phase: 1 }));
    expect(hint).toEqual({
      targetAction: 'next',
      reason: 'frontendHint.sakuraRoundIsOver',
      confidence: 'strong',
    });
  });

  it('advises nothing while a CPU is thinking', () => {
    expect(getSakuraHint(makeSakuraState({ isHumanTurn: false }))).toBeNull();
  });

  it('advises taking when a match exists', () => {
    const hint = getSakuraHint(makeSakuraState({ captureOptions: { 1: [0] } }));
    expect(hint?.reason).toBe('frontendHint.sakuraTakeTheMatch');
    expect(hint?.confidence).toBe('strong');
  });

  // **空の候補表は「取れる」ではない。** 0 件のキーがあるだけで取れると答えると、
  // 合う札が 1 枚も無い局面で「取れ」と言うことになる。
  it('advises discarding when nothing matches', () => {
    const empties: Record<number, number[]>[] = [{}, { 0: [] }];
    for (const captureOptions of empties) {
      const hint = getSakuraHint(makeSakuraState({ captureOptions }));
      expect(hint?.reason).toBe('frontendHint.sakuraDiscardCheapest');
      expect(hint?.confidence).toBe('moderate');
    }
  });
});
