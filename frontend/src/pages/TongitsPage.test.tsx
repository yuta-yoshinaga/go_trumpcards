import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tongitsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TongitsResponse } from '../types/card';
import { TongitsPhase } from '../types/phases';
import { TongitsPage } from './TongitsPage';

vi.mock('../api/gameApi', () => ({ tongitsApi: { exec: vi.fn() }, actionLogApi: { tongits: vi.fn() } }));
const mockExec = vi.mocked(tongitsApi.exec);
const card = (design: string, value: number): Card => ({ design, value }) as Card;
function state(overrides: Partial<TongitsResponse> = {}): TongitsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 3), card('HEART', 3), card('CLOVER', 3), card('DIAMOND', 9), card('SPADE', 11)],
        melds: [],
        roundScore: 0,
        cumulativeScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 5,
        cards: [],
        melds: [{ cards: [card('HEART', 5), card('HEART', 6), card('HEART', 7)] }],
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    phase: TongitsPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 2),
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    isTongits: false,
    remainingPoints: 3,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 50 },
    ...overrides,
  };
}
beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(state());
});

describe('TongitsPage', () => {
  it('declares a challenge without a selected card', async () => {
    renderWithProviders(<TongitsPage />);
    fireEvent.click(await screen.findByRole('button', { name: /チャレンジ/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('challenge', undefined, undefined, undefined, undefined, undefined, [
        true,
        true,
      ]),
    );
  });
  it('melds the selected cards', async () => {
    renderWithProviders(<TongitsPage />);
    for (const index of [0, 1, 2]) fireEvent.click(await screen.findByTestId(`tongits-hand-${index}`));
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, [0, 1, 2]));
  });
  it('shows public melds and sapaw controls', async () => {
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByTestId('tongits-melds')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /サパウ/ })).toBeInTheDocument();
  });
  it('shows remaining points and keeps challenge available at any score', async () => {
    mockExec.mockResolvedValue(state({ remainingPoints: 6 }));
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByTestId('tongits-remaining-points')).toHaveTextContent('6');
    expect(await screen.findByRole('button', { name: /チャレンジ/ })).toBeEnabled();
  });
});
