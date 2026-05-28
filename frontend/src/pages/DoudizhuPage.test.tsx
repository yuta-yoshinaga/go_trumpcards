import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doudizhuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DoudizhuResponse } from '../types/card';
import { DoudizhuPage } from './DoudizhuPage';

vi.mock('../api/gameApi', () => ({
  doudizhuApi: { exec: vi.fn() },
  actionLogApi: { doudizhu: vi.fn() },
}));

const mockExec = vi.mocked(doudizhuApi.exec);

const defaultState: DoudizhuResponse = {
  players: [
    { id: 0, isHuman: true, isFinished: false, isLandlord: true, cardCount: 20, cards: [] },
    { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
    { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
  ],
  phase: 'play',
  currentTurn: 0,
  tableCards: [],
  tableCombo: '',
  kittyCards: [],
  landlordIdx: 0,
  baseBid: 1,
  highestBid: 1,
  bombCount: 0,
  scores: [0, 0, 0],
  gameEndFlag: false,
  config: { cpuDifficulty: 0 },
  cpuActions: [],
  humanAction: null,
  message: '',
};

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(defaultState);
});

describe('DoudizhuPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DoudizhuPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'reset' }));
    });
  });

  it('renders CPU player areas after load', async () => {
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    });
  });

  it('renders result message when game ends', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'end',
      gameEndFlag: true,
      scores: [2, -1, -1],
    });
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByText('地主の勝利！')).toBeInTheDocument();
    });
  });
});
