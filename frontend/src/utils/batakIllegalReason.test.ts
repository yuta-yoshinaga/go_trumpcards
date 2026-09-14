import { describe, expect, it } from 'vitest';
import type { Card } from '../types/common';
import type { BatakTrickCard } from '../types/games/batak';
import { batakIllegalReason } from './batakIllegalReason';

const spadeAce: Card = { design: 'SPADE', value: 1 };
const heartJack: Card = { design: 'HEART', value: 11 };
const diamondThree: Card = { design: 'DIAMOND', value: 3 };

const reason = (
  cards: Card[],
  currentTrick: BatakTrickCard[] = [],
  spadesBroken = false,
  validPlayIndices = [1],
  cardIndex = 0,
) => batakIllegalReason(cardIndex, cards, currentTrick, spadesBroken, validPlayIndices);

describe('batakIllegalReason', () => {
  it('explains an unbroken spade lead when a non-spade is in hand', () => {
    expect(reason([spadeAce, heartJack], [], false, [1])).toBe('spadesNotBroken');
  });

  it('explains when a player must follow the lead suit', () => {
    expect(reason([diamondThree, heartJack], [{ playerIdx: 1, card: diamondThree }], false, [0], 1)).toBe('followSuit');
  });

  it('explains when a void player must trump with a spade', () => {
    expect(reason([heartJack, spadeAce], [{ playerIdx: 1, card: diamondThree }], false, [1])).toBe('mustTrumpSpade');
  });

  it('does not restrict a lead when the hand contains only spades', () => {
    expect(reason([spadeAce, { design: 'SPADE', value: 2 }], [], false, [0, 1])).toBeUndefined();
  });

  it('does not require a spade cut when spades were led', () => {
    expect(reason([heartJack, spadeAce], [{ playerIdx: 1, card: spadeAce }], false, [0])).toBeUndefined();
  });

  it('does not explain a card that is in validPlayIndices', () => {
    expect(reason([spadeAce, heartJack], [], false, [0, 1])).toBeUndefined();
  });

  it('does not explain a card index outside the hand', () => {
    expect(batakIllegalReason(2, [spadeAce, heartJack], [], false, [])).toBeUndefined();
  });

  it('does not restrict a spade lead after spades have been broken', () => {
    expect(reason([spadeAce, heartJack], [], true, [1])).toBeUndefined();
  });

  it('does not explain an excluded card that follows the lead suit', () => {
    expect(reason([diamondThree, heartJack], [{ playerIdx: 1, card: diamondThree }], false, [], 0)).toBeUndefined();
  });

  it('does not explain an excluded spade cut when the hand is void in the lead suit', () => {
    expect(reason([heartJack, spadeAce], [{ playerIdx: 1, card: diamondThree }], false, [], 1)).toBeUndefined();
  });

  it('does not explain an excluded discard when the hand has neither lead suit nor spades', () => {
    expect(
      reason([heartJack, diamondThree], [{ playerIdx: 1, card: { design: 'CLOVER', value: 2 } }], false, [], 1),
    ).toBeUndefined();
  });
});
