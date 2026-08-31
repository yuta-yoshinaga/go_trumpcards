import { describe, expect, it } from 'vitest';
import type { KalookiResponse } from '../../../types/card';
import { formatKalookiState } from './kalookiFormatter';

const baseState: KalookiResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 10 },
      ],
      melds: [
        {
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
        },
      ],
      hasOpened: true,
      roundScore: 12,
      cumulativeScore: 40,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      melds: [],
      hasOpened: false,
      roundScore: 0,
      cumulativeScore: 25,
    },
  ],
  phase: 1,
  openingThreshold: 51,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 28,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  message: '',
  messageCode: '',
  config: { cpuDifficulty: 1, playerCount: 2, openingThreshold: 51 },
};

describe('formatKalookiState', () => {
  it('renders header, phase, opening threshold, scores, opened flags, and melds', () => {
    const out = formatKalookiState(baseState);
    expect(out).toContain('Kalooki');
    expect(out).toContain('MELD');
    expect(out).toContain('opening: 51');
    expect(out).toContain('score=40 (+12)');
    expect(out).toContain('cards=13');
    expect(out).toContain('[opened]');
    expect(out).toContain('[not opened]');
    expect(out).toContain('melds:');
    expect(out).toContain('turn:');
  });

  // `layoff <playerIdx> <meldIdx> <cardIdx>` が要求する番号が場に出ていなかった (#6462)。
  it('numbers each meld with the index layoff takes', () => {
    const twoMelds: KalookiResponse = {
      ...baseState,
      players: [
        {
          ...baseState.players[0],
          melds: [
            baseState.players[0].melds[0],
            {
              cards: [
                { design: 'CLOVER', value: 9 },
                { design: 'CLOVER', value: 10 },
                { design: 'CLOVER', value: 11 },
              ],
            },
          ],
        },
        baseState.players[1],
      ],
    };

    const out = formatKalookiState(twoMelds);
    // 番号だけでなく、その番号の後ろに**そのメルドの札**が並ぶことを見る。
    // 1 始まりにしたり順序を入れ替えたりする変異はこれで落ちる。
    expect(out).toContain('[0] ♠5, ♥5, ♣5');
    expect(out).toContain('[1] ♣9, ♣10, ♣J');
    expect(out).not.toContain('[2] ');
  });

  it('shows present discard top', () => {
    const out = formatKalookiState(baseState);
    expect(out).toContain('discard: ♠5');
  });

  it('handles empty discard top', () => {
    const out = formatKalookiState({ ...baseState, discardTop: null });
    expect(out).toContain('[  ]');
  });

  it('shows human winner at game end', () => {
    const out = formatKalookiState({ ...baseState, gameEndFlag: true, phase: 3, winnerIdx: 0 });
    expect(out).toContain('Game Over! Winner:');
    expect(out).not.toContain('turn:');
  });

  it('shows cpu winner at game end', () => {
    const out = formatKalookiState({ ...baseState, gameEndFlag: true, phase: 3, winnerIdx: 1 });
    expect(out).toContain('Game Over! Winner:');
  });

  it('shows draw at game end when no winner', () => {
    const out = formatKalookiState({ ...baseState, gameEndFlag: true, phase: 3, winnerIdx: -1 });
    expect(out).toContain('Game Over! Draw');
  });
});
