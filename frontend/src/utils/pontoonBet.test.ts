import { describe, expect, it } from 'vitest';
import { PONTOON_MIN_BET, pontoonBuyChoices, pontoonClampBuy, pontoonMaxBuy } from './pontoonBet';

describe('pontoonMaxBuy', () => {
  it('is twice the current bet', () => {
    expect(pontoonMaxBuy(100)).toBe(200);
  });

  it('never drops below the floor, even at a zero bet', () => {
    expect(pontoonMaxBuy(0)).toBe(PONTOON_MIN_BET);
  });
});

describe('pontoonBuyChoices', () => {
  it('walks the floor to twice the bet in floor-sized steps', () => {
    expect(pontoonBuyChoices(30)).toEqual([10, 20, 30, 40, 50, 60]);
  });

  it('still offers the floor when the bet is below it', () => {
    expect(pontoonBuyChoices(0)).toEqual([PONTOON_MIN_BET]);
  });
});

describe('pontoonClampBuy', () => {
  it('follows the current bet when nothing was chosen', () => {
    expect(pontoonClampBuy(null, 100)).toBe(100);
  });

  it('keeps a legal choice', () => {
    expect(pontoonClampBuy(150, 100)).toBe(150);
  });

  it('pulls a choice down when the next hand has a smaller stake', () => {
    // Picked 150 on a 100 hand, then the stake drops to 20: 40 is the new ceiling.
    expect(pontoonClampBuy(150, 20)).toBe(40);
  });

  it('pulls a choice up to the floor', () => {
    expect(pontoonClampBuy(5, 100)).toBe(PONTOON_MIN_BET);
  });
});
