import { describe, expect, it } from 'vitest';
import type { Card, GoofspielResponse } from '../../../types/card';
import { formatGoofspielState } from './goofspielFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 12,
  cards: id === 0 ? [card('SPADE', 1), card('SPADE', 2)] : [card('CLOVER', 1), card('CLOVER', 2)],
  score: 0,
  hasBid: false,
  ...over,
});

const state = (over: Partial<GoofspielResponse> = {}): GoofspielResponse =>
  ({
    players: [seat(0), seat(1)],
    phase: 0,
    validPlays: [0, 1],
    currentPrize: card('DIAMOND', 9),
    carriedPrizes: [],
    prizeValue: 9,
    prizeRemaining: 11,
    lastWinnerIdx: -1,
    lastGained: 0,
    roundNumber: 2,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 2, tieRule: 0 },
    message: '',
    ...over,
  }) as unknown as GoofspielResponse;

describe('formatGoofspielState', () => {
  it('reports loading for a null state', () => {
    expect(formatGoofspielState(null)).toBe('Loading...');
  });

  it('shows the round, prizes left and the rule', () => {
    const out = formatGoofspielState(state());
    expect(out).toContain('round 2');
    expect(out).toContain('11 prizes left');
    expect(out).toMatch(/everyone bids face down at the same time/);
    expect(out).toMatch(/prize: .* — 9 points/);
  });

  // **持ち越しは「今回の賞が増える」こと。**
  it('notes a carry-over in the prize line', () => {
    expect(formatGoofspielState(state({ carriedPrizes: [card('DIAMOND', 4)], prizeValue: 13 }))).toMatch(
      /13 points \(incl\. 1 carried\)/,
    );
    expect(formatGoofspielState(state())).not.toMatch(/carried/);
  });

  // **伏せたことは見せますが、中身は公開まで見せません。**
  it('shows that a seat has bid without showing the card', () => {
    const out = formatGoofspielState(state({ players: [seat(0, { hasBid: true }), seat(1)] }));
    expect(out).toContain('[has bid]');
    expect(out).not.toContain('[played');
    expect(out).toMatch(/you have bid/);
  });

  it('shows the revealed bids once they are face up', () => {
    const out = formatGoofspielState(
      state({
        phase: 1,
        currentPrize: undefined,
        lastWinnerIdx: 1,
        lastGained: 9,
        players: [seat(0, { revealedBid: card('SPADE', 3) }), seat(1, { revealedBid: card('CLOVER', 11), score: 9 })],
      }),
    );
    expect(out).toMatch(/\[played /);
    expect(out).toMatch(/CPU 1 takes 9 points/);
    expect(out).toMatch(/next — turn the next prize card/);
  });

  // **同点は誰も取りません。**
  it('reports a tie as nobody taking the prize', () => {
    const out = formatGoofspielState(state({ phase: 1, currentPrize: undefined, lastWinnerIdx: -1, lastGained: 0 }));
    expect(out).toMatch(/a tie — nobody takes this prize/);
  });

  // **残り札は全員分を出す。** 使った札は場に出るので隠せていません。
  it('lists every seat remaining cards, CPUs included', () => {
    const out = formatGoofspielState(state());
    expect(out).toMatch(/\[0\]♠A/);
    expect(out).toMatch(/♣A/);
    expect(out).toMatch(/12 left, 0 points/);
  });

  it('names the winner once the game ends', () => {
    const out = formatGoofspielState(
      state({ phase: 2, gameEndFlag: true, winnerIdx: 0, players: [seat(0, { score: 50 }), seat(1, { score: 41 })] }),
    );
    expect(out).toMatch(/game over — あなた took the most \(50\)/);
  });

  it('appends the server message', () => {
    expect(formatGoofspielState(state({ message: 'boom' }))).toContain('boom');
  });
});
