import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, CaribbeanDrawResponse } from '../../../types/card';
import { CaribbeanDrawPhase } from '../../../types/phases';
import { formatCaribbeandrawState } from './caribbeandrawFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: CaribbeanDrawResponse = {
  playerHand: [],
  dealerHand: [],
  phase: CaribbeanDrawPhase.BET,
  chips: 1000,
  anteBet: 0,
  jackpotBet: 0,
  playBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  jackpotPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  drawCost: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const drawPhaseState: CaribbeanDrawResponse = {
  ...betPhaseState,
  phase: CaribbeanDrawPhase.DRAW,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const actionPhaseState: CaribbeanDrawResponse = {
  ...drawPhaseState,
  phase: CaribbeanDrawPhase.ACTION,
};

const endPhasePlayerWins: CaribbeanDrawResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: CaribbeanDrawPhase.END,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
  playBet: 200,
  result: 1,
  antePayout: 200,
  playPayout: 600,
  jackpotPayout: 0,
  totalPayout: 800,
  dealerQualified: true,
  drawCost: 0,
  playerHandRank: 3,
  dealerHandRank: 1,
  message: 'Player wins!',
};

const endPhaseFold: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  result: -1,
  playBet: 0,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  dealerHand: [],
  message: 'Player folded.',
};

const endPhaseWithJackpot: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  jackpotBet: 10,
  jackpotPayout: 1000,
  totalPayout: 1800,
};

describe('formatCaribbeandrawState', () => {
  it('formats bet phase with chips and phase name', () => {
    const result = formatCaribbeandrawState(betPhaseState);
    expect(result).toContain('chips: 1000');
    expect(result).toContain('phase: BET');
    expect(result).not.toContain('Your hand');
    expect(result).not.toContain('Dealer');
  });

  describe('draw phase', () => {
    it('names the draw phase rather than falling through to ACTION', () => {
      // Caribbean Stud's numbering has ACTION at 2. Keeping it would label the
      // whole draw phase "ACTION" while the only legal command is `d`.
      expect(formatCaribbeandrawState(drawPhaseState)).toContain('phase: DRAW');
    });

    it('states the exchange fee before the player commits to it', () => {
      const result = formatCaribbeandrawState(drawPhaseState);
      expect(result).toContain('Exchange up to 2 cards for 100');
    });

    it('numbers the hand from 1, matching what `d <n>` expects', () => {
      // `d 1` discards playerHand[0]; a `[0]` label next to that card would send
      // the player one position off with no error to show for it.
      const result = formatCaribbeandrawState(drawPhaseState);
      expect(result).toContain('[1]♠10');
      expect(result).toContain('[5]♠7');
      expect(result).not.toContain('[0]');
    });

    it('keeps the dealer hand masked', () => {
      const result = formatCaribbeandrawState(drawPhaseState);
      expect(result).toContain('♥K');
      expect(result).toContain('??');
    });

    it('does not show a fee line before anything has been exchanged', () => {
      expect(formatCaribbeandrawState(drawPhaseState)).not.toContain('draw fee:');
    });
  });

  it('formats action phase with player hand and masked dealer hand', () => {
    const result = formatCaribbeandrawState(actionPhaseState);
    expect(result).toContain('phase: ACTION');
    expect(result).toContain('Your hand:');
    // First dealer card visible
    expect(result).toContain('♥K');
    // Remaining dealer cards hidden
    expect(result).toContain('??');
    expect(result).toContain('ante: 100');
    // The draw is over; the prompt for it must not linger.
    expect(result).not.toContain('Exchange up to');
  });

  it('formats end phase with full dealer hand and qualification', () => {
    const result = formatCaribbeandrawState(endPhasePlayerWins);
    expect(result).toContain('phase: END');
    expect(result).toContain('Dealer:');
    expect(result).toContain('Dealer qualified: yes');
    expect(result).toContain('payout: ante=200 play=600 jackpot=0');
    expect(result).toContain('total: 800');
    expect(result).toContain('Player wins!');
  });

  it('reports the draw fee that was paid, which totalPayout never includes', () => {
    const result = formatCaribbeandrawState({ ...endPhasePlayerWins, drawCost: 100 });
    expect(result).toContain('draw fee: 100');
  });

  it('omits the draw fee line when the player stood pat', () => {
    expect(formatCaribbeandrawState(endPhasePlayerWins)).not.toContain('draw fee');
  });

  it('formats end phase after fold (no dealer hand)', () => {
    const result = formatCaribbeandrawState(endPhaseFold);
    expect(result).toContain('phase: END');
    expect(result).not.toContain('Dealer:');
    expect(result).toContain('Player folded.');
  });

  it('formats jackpot bets and payout', () => {
    const result = formatCaribbeandrawState(endPhaseWithJackpot);
    expect(result).toContain('jackpot: 10');
    expect(result).toContain('jackpot=1000');
    expect(result).toContain('total: 1800');
  });

  it('formats unknown phase gracefully', () => {
    const state = { ...betPhaseState, phase: 99 };
    const result = formatCaribbeandrawState(state);
    expect(result).toContain('phase: UNKNOWN');
  });

  it('includes play bet when set', () => {
    const state = { ...endPhasePlayerWins };
    const result = formatCaribbeandrawState(state);
    expect(result).toContain('play bet: 200');
  });
});
