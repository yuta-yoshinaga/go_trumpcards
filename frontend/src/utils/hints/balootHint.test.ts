import { describe, expect, it } from 'vitest';
import type { BalootResponse } from '../../types/card';
import { getBalootHint } from './balootHint';

const base = (hint?: BalootResponse['hint']): BalootResponse => ({ hint }) as unknown as BalootResponse;

describe('getBalootHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getBalootHint(base())).toBeNull();
  });

  // **宣言のヒントは札を指さない。** cardIndex が無くても null にしない。
  it.each([
    ['balootDeclareSun', 'declare-sun'],
    ['balootPassDeclare', 'pass-declare'],
  ])('turns %s into the %s decision', (reason, targetAction) => {
    expect(getBalootHint(base({ reason, suit: 0 }))).toEqual({
      targetAction,
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  // **Hokom はスートまで指す。** モードだけでは切り札が決まらず、序列も決まらない。
  it('names the suit when it recommends Hokom', () => {
    expect(getBalootHint(base({ reason: 'balootDeclareHokom', suit: 3 }))).toEqual({
      targetAction: 'declare-hokom-3',
      reason: 'hint.balootDeclareHokom',
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getBalootHint(base({ cardIndex: 0, reason: 'balootWinTrick', suit: 0 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.balootWinTrick',
      confidence: 'moderate',
    });
  });

  // 味方に点を乗せる手はほぼ一択。両側を踏む。
  it('is more confident about feeding a winning partner', () => {
    expect(getBalootHint(base({ cardIndex: 2, reason: 'balootFeedPartner', suit: 0 }))?.confidence).toBe('strong');
    expect(getBalootHint(base({ cardIndex: 2, reason: 'balootWinTrick', suit: 0 }))?.confidence).toBe('moderate');
  });
});
