import { describe, expect, it } from 'vitest';
import type { CucumberResponse } from '../../types/card';
import { getCucumberHint } from './cucumberHint';

const state = (hint?: CucumberResponse['hint']): CucumberResponse => ({ hint }) as CucumberResponse;

describe('getCucumberHint', () => {
  it('returns null without a hint', () => {
    expect(getCucumberHint(state())).toBeNull();
  });

  // **更新できない場面は選択の余地がない。**
  it('is certain when the play is forced', () => {
    expect(getCucumberHint(state({ cardIndex: 0, reason: 'cucumberForced' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.cucumberForced',
      confidence: 'strong',
    });
  });

  it.each(['cucumberLead', 'cucumberBeat'])('is only moderate for %s', (reason) => {
    expect(getCucumberHint(state({ cardIndex: 2, reason }))).toEqual({
      targetAction: 'card-2',
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  it('returns null when the hint names no card', () => {
    expect(getCucumberHint(state({ reason: 'cucumberBeat' }))).toBeNull();
  });
});
