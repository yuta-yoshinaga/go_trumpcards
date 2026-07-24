import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, fourcardpokerApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FourCardPokerResponse } from '../types/card';
import { FourCardPokerPage } from './FourCardPokerPage';

vi.mock('../hooks/useGameHint');

vi.mock('../api/gameApi', () => ({
  fourcardpokerApi: { exec: vi.fn() },
  actionLogApi: { fourcardpoker: vi.fn() },
}));

const mockExec = vi.mocked(fourcardpokerApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: FourCardPokerResponse = {
  playerHand: [],
  dealerHand: [],
  playerBest: [],
  dealerBest: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  acesUpBet: 0,
  playBet: 0,
  playMultiplier: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  anteBonusPayout: 0,
  acesUpPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const actionPhaseState: FourCardPokerResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 7)], // upcard only
  anteBet: 100,
  chips: 900,
};

const endPhasePlayerWins: FourCardPokerResponse = {
  playerHand: [card('SPADE', 9), card('CLOVER', 9), card('HEART', 9), card('DIAMOND', 5), card('CLOVER', 2)],
  dealerHand: [
    card('CLOVER', 7),
    card('SPADE', 2),
    card('HEART', 4),
    card('DIAMOND', 6),
    card('SPADE', 8),
    card('CLOVER', 10),
  ],
  playerBest: [card('SPADE', 9), card('CLOVER', 9), card('HEART', 9), card('DIAMOND', 5)],
  dealerBest: [card('CLOVER', 7), card('SPADE', 8), card('CLOVER', 10), card('DIAMOND', 6)],
  phase: 3,
  chips: 1400,
  anteBet: 100,
  acesUpBet: 0,
  playBet: 100,
  playMultiplier: 1,
  result: 1,
  antePayout: 200,
  playPayout: 200,
  anteBonusPayout: 200,
  acesUpPayout: 0,
  totalPayout: 600,
  playerHandRank: 6,
  dealerHandRank: 1,
  message: 'You Win!',
  messageCode: 'fourcardpoker.result.playerWins',
};

const endPhaseDealerWins: FourCardPokerResponse = {
  ...endPhasePlayerWins,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  anteBonusPayout: 0,
  totalPayout: 0,
  message: 'Dealer Wins!',
  messageCode: 'fourcardpoker.result.dealerWins',
};

const endPhaseFold: FourCardPokerResponse = {
  ...endPhasePlayerWins,
  result: -1,
  playBet: 0,
  playMultiplier: 0,
  antePayout: 0,
  playPayout: 0,
  anteBonusPayout: 0,
  totalPayout: 0,
  dealerHand: [],
  message: 'Folded',
  messageCode: 'fourcardpoker.result.fold',
};

const endPhasePush: FourCardPokerResponse = {
  ...endPhasePlayerWins,
  result: 0,
  antePayout: 100,
  playPayout: 100,
  anteBonusPayout: 0,
  totalPayout: 200,
  message: 'Push!',
  messageCode: 'fourcardpoker.result.push',
};

const endPhaseWithAcesUp: FourCardPokerResponse = {
  ...endPhasePlayerWins,
  acesUpBet: 50,
  acesUpPayout: 200,
  totalPayout: 800,
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('FourCardPokerPage', () => {
  it('renders bet phase on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<FourCardPokerPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('sends bet command with ante and acesUp amounts', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0));
  });

  it('renders action phase with per-multiplier play buttons and fold', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByTestId('play-1x')).toBeInTheDocument());
    expect(screen.getByTestId('play-2x')).toBeInTheDocument();
    expect(screen.getByTestId('play-3x')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    // The select-box multiplier control is gone.
    expect(screen.queryByLabelText(/プレイ倍率/)).not.toBeInTheDocument();
  });

  it('sends play command with the chosen multiplier button', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    fireEvent.click(await screen.findByTestId('play-1x'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 1));
    fireEvent.click(screen.getByTestId('play-2x'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 2));
    fireEvent.click(screen.getByTestId('play-3x'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 3));
  });

  it('plays via the 1/2/3 keyboard shortcuts', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await screen.findByTestId('play-1x');
    for (const key of ['1', '2', '3'] as const) {
      fireEvent.keyDown(document, { key });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, Number(key)));
    }
  });

  it('renders the dealer concealed cards as backs during the action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await screen.findByTestId('play-1x');
    // Dealer holds 6 cards but only the upcard is revealed; the other 5 are backs.
    expect(screen.getAllByRole('img', { name: '非公開のカード' })).toHaveLength(5);
  });

  it('reveals all dealer cards face-up (no backs) at the end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('You Win!')).toBeInTheDocument());
    expect(screen.queryByRole('img', { name: '非公開のカード' })).not.toBeInTheDocument();
  });

  it('sends fold command', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('renders end phase with player win result', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('You Win!')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('renders dealer wins', async () => {
    mockExec.mockResolvedValue(endPhaseDealerWins);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('Dealer Wins!')).toBeInTheDocument());
  });

  it('renders push', async () => {
    mockExec.mockResolvedValue(endPhasePush);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('Push!')).toBeInTheDocument());
  });

  it('renders folded state', async () => {
    mockExec.mockResolvedValue(endPhaseFold);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('Folded')).toBeInTheDocument());
  });

  it('shows aces up payout when sidebet wins', async () => {
    mockExec.mockResolvedValue(endPhaseWithAcesUp);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/エースズアップ:.*200/)).toBeInTheDocument();
  });

  it('uses the actionLogApi mock', () => {
    expect(actionLogApi).toBeDefined();
  });

  it('renders ante and aces-up as ChipBetInputs with steppers', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アンテ +10' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'エースズアップ +10' })).toBeInTheDocument();
  });

  it('disables Bet and shows an alert when the combined wager exceeds the balance', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<FourCardPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Ante defaults to 100; pushing Aces Up to the full balance makes the sum exceed it.
    const acesUp = screen.getByLabelText('エースズアップ') as HTMLInputElement;
    fireEvent.change(acesUp, { target: { value: '1000' } });

    expect(screen.getByRole('alert')).toHaveTextContent(/残高/);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
  });
});
