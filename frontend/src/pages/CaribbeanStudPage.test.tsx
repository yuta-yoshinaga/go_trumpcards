import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { caribbeanstudApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CaribbeanStudResponse } from '../types/card';
import { CaribbeanStudPage } from './CaribbeanStudPage';

vi.mock('../api/gameApi', () => ({
  caribbeanstudApi: { exec: vi.fn() },
  actionLogApi: { caribbeanstud: vi.fn() },
}));

const mockApi = vi.mocked(caribbeanstudApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: CaribbeanStudResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
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
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const maskedCard = { design: '' as CardDesign, value: 0 };

const actionPhaseState: CaribbeanStudResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  // 1 face-up dealer card + 4 masked cards (security: dealer hand not fully revealed)
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const endPhasePlayerWins: CaribbeanStudResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: 3,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
  playBet: 200,
  result: 1,
  antePayout: 200,
  playPayout: 800,
  jackpotPayout: 0,
  totalPayout: 1000,
  dealerQualified: true,
  playerHandRank: 3,
  dealerHandRank: 1,
  message: '勝利！',
  messageCode: 'caribbeanstud.result.playerWins',
};

const endPhaseDealerWins: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  message: 'ディーラー勝利！',
  messageCode: 'caribbeanstud.result.dealerWins',
};

const endPhaseFold: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  result: -1,
  playBet: 0,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  dealerHand: [],
  dealerQualified: false,
  dealerHandRank: 0,
  message: 'フォールド',
  messageCode: 'caribbeanstud.result.fold',
};

const endPhasePush: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  result: 0,
  antePayout: 100,
  playPayout: 200,
  totalPayout: 300,
  message: '引き分け！',
  messageCode: 'caribbeanstud.result.push',
};

const endPhaseDealerNotQualified: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  dealerQualified: false,
  antePayout: 200,
  playPayout: 200,
  totalPayout: 400,
  message: 'ディーラー未クオリファイ！',
  messageCode: 'caribbeanstud.result.dealerNotQualified',
};

const endPhaseWithJackpot: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  jackpotBet: 10,
  jackpotPayout: 1000,
  totalPayout: 2000,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('CaribbeanStudPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<CaribbeanStudPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows action phase with call and fold buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseFold);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePush);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('shows end phase with dealer not qualified', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerNotQualified);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getAllByText(/未クオリファイ/).length).toBeGreaterThanOrEqual(1));
  });

  it('shows payout breakdown with jackpot', async () => {
    mockApi.mockResolvedValue(endPhaseWithJackpot);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/ジャックポット: 1000/)).toBeInTheDocument();
    expect(screen.getByText(/合計: 2000/)).toBeInTheDocument();
  });

  it('can change ante and jackpot amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const jpInput = screen.getByLabelText('ジャックポット');
    fireEvent.change(jpInput, { target: { value: '10' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 10));
  });

  it('next game button executes reset without confirm dialog', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders player and dealer cards in end phase', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(10));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  it('renders 1 face-up and 4 face-down dealer cards in action phase', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    // 5 player face-up + 1 dealer face-up + 4 dealer face-down (CardBack images) = 10 imgs
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBe(10);
    // Dealer section is shown
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  it('shows dealer qualification status', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ/)).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in action phase', async () => {
    localStorage.setItem('hint_enabled_caribbeanstud', 'true');
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });
});
