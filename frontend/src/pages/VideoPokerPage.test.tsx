import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { videopokerApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, VideoPokerResponse } from '../types/card';
import { VideoPokerPage } from './VideoPokerPage';

vi.mock('../api/gameApi', () => ({
  videopokerApi: { exec: vi.fn() },
  actionLogApi: { videopoker: vi.fn() },
}));

const mockExec = vi.mocked(videopokerApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: VideoPokerResponse = {
  hand: [],
  phase: 1,
  chips: 1000,
  betAmount: 0,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: '',
};

const drawPhaseState: VideoPokerResponse = {
  hand: [card('SPADE', 1), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 13)],
  phase: 2,
  chips: 997,
  betAmount: 3,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: '',
};

const resultPhaseWin: VideoPokerResponse = {
  hand: [card('SPADE', 11), card('CLOVER', 11), card('HEART', 3), card('DIAMOND', 5), card('SPADE', 9)],
  phase: 3,
  chips: 1001,
  betAmount: 1,
  result: 1,
  payout: 1,
  handRank: 1,
  handName: 'Jacks or Better',
  heldIndices: [true, true, false, false, false],
  message: 'Jacks or Better! You win!',
  messageCode: 'videopoker.result.win',
  variantName: 'jacksorbetter',
  messageParams: { handName: 'Jacks or Better', payout: '1' },
};

const resultPhaseLose: VideoPokerResponse = {
  hand: [card('SPADE', 2), card('CLOVER', 5), card('HEART', 7), card('DIAMOND', 9), card('SPADE', 11)],
  phase: 3,
  chips: 999,
  betAmount: 1,
  result: -1,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  message: 'No winning hand.',
  variantName: 'jacksorbetter',
  messageCode: 'videopoker.result.lose',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('VideoPokerPage', () => {
  it('calls reset on mount and renders bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/チップ.*1000/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument();
  });

  it('renders draw phase with 5 cards', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
  });

  it('renders result phase with win message', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のハンド/)).toBeInTheDocument());
  });

  it('renders result phase with lose message', async () => {
    mockExec.mockResolvedValue(resultPhaseLose);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のハンド/)).toBeInTheDocument());
  });

  it('reset button shows confirmation dialog', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のハンド/)).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /次のハンド/ }));
    await waitFor(() => expect(screen.getByText(/リセットしますか/)).toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
