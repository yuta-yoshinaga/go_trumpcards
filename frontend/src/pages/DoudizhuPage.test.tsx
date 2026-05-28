import { fireEvent, screen, waitFor } from '@testing-library/react';
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

  it('renders peasant-win message when landlord loses', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'end',
      gameEndFlag: true,
      scores: [-2, 1, 1],
    });
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByText('農民の勝利！')).toBeInTheDocument();
    });
  });

  it('renders bid buttons above the highest bid during bid phase', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      phase: 'bid',
      landlordIdx: -1,
      highestBid: 1,
      currentTurn: 0,
    });
    renderWithProviders(<DoudizhuPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /2/ })).toBeInTheDocument();
    });
    // Bid 1 is not offered because highestBid is already 1.
    expect(screen.queryByRole('button', { name: '1で叫ぶ' })).not.toBeInTheDocument();
  });

  it('disables play until a card is selected, then plays on click', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          isLandlord: true,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 5 }],
        },
        { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
        { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      ],
    });
    renderWithProviders(<DoudizhuPage />);

    const playButton = await screen.findByRole('button', { name: '出す' });
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByAltText(/5/));
    expect(playButton).toBeEnabled();

    fireEvent.click(playButton);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p', indices: [0] }));
    });
  });
});
