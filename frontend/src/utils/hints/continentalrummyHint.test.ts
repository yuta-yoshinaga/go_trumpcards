import { describe, expect, it } from 'vitest';
import { makeContinentalRummyState } from '../../test/stateFactories';
import { getContinentalrummyHint } from './continentalrummyHint';

describe('getContinentalrummyHint', () => {
  // **上がれるときは札ではなく「上がる」を指す。**
  it('points at the go-out button when the server says the hand is complete', () => {
    expect(getContinentalrummyHint(makeContinentalRummyState())).toEqual({
      targetAction: 'goout',
      reason: 'frontendHint.continentalrummy_go_out',
      confidence: 'strong',
    });
  });

  // **どの札を捨てるかまで言う。** 「緩い札を」だけだと player が探し直す。
  it('names the card to throw when the hand cannot go out', () => {
    expect(
      getContinentalrummyHint(
        makeContinentalRummyState({ goOutIdx: -1, hintDiscardIdx: 4, hintReason: 'discard_loose' }),
      ),
    ).toEqual({
      targetAction: 'discard',
      targetPos: 4,
      reason: 'frontendHint.continentalrummy_discard_loose',
      confidence: 'moderate',
    });
  });

  // **引かずに上がれるならそれを指す。** 引くと 10 点が 7 点に落ちる。
  it('points at the deal go-out ahead of either draw', () => {
    expect(
      getContinentalrummyHint(
        makeContinentalRummyState({ phase: 'draw', canGoOutOnDeal: true, hintReason: 'take_discard' }),
      ),
    ).toEqual({
      targetAction: 'gooutdeal',
      reason: 'frontendHint.continentalrummy_go_out_on_deal',
      confidence: 'strong',
    });
  });

  it.each([
    ['take_discard', 'take', 'strong'],
    ['draw_stock', 'stock', 'moderate'],
  ] as const)('on the draw it points at %s', (reason, action, confidence) => {
    expect(getContinentalrummyHint(makeContinentalRummyState({ phase: 'draw', hintReason: reason }))).toEqual({
      targetAction: action,
      reason: `frontendHint.continentalrummy_${reason}`,
      confidence,
    });
  });

  it.each([
    ['the game has ended', { gameEndFlag: true }],
    ['it is not the human turn', { isHumanTurn: false }],
    ['the round is over', { phase: 'roundEnd' }],
    ['the server sent no reason', { hintReason: '' }],
  ])('is silent when %s', (_label, overrides) => {
    expect(getContinentalrummyHint(makeContinentalRummyState(overrides))).toBeNull();
  });

  it('is silent on the discard when there is neither a go-out nor a card to name', () => {
    expect(getContinentalrummyHint(makeContinentalRummyState({ goOutIdx: -1, hintDiscardIdx: -1 }))).toBeNull();
  });
});
