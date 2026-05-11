import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ultimatetexasholdemApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, UltimateTexasHoldemResponse } from '../types/card';
import { UltimateTexasHoldemPage } from './UltimateTexasHoldemPage';

vi.mock('../api/gameApi', () => ({
  ultimatetexasholdemApi: { exec: vi.fn() },
  actionLogApi: { ultimatetexasholdem: vi.fn() },
}));

const mockApi = vi.mocked(ultimatetexasholdemApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: UltimateTexasHoldemResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  blindBet: 0,
  tripsBet: 0,
  playBet: 0,
  folded: false,
  result: 0,
  dealerQualified: false,
  antePayout: 0,
  blindPayout: 0,
  playPayout: 0,
  tripsPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const preFlopState: UltimateTexasHoldemResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [maskedCard, maskedCard],
  anteBet: 100,
  blindBet: 100,
  chips: 800,
};

const flopState: UltimateTexasHoldemResponse = {
  ...preFlopState,
  phase: 3,
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10)],
};

const riverState: UltimateTexasHoldemResponse = {
  ...flopState,
  phase: 4,
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2), card('HEART', 4)],
};

const endPlayerWins: UltimateTexasHoldemResponse = {
  ...riverState,
  phase: 5,
  dealerHand: [card('HEART', 7), card('DIAMOND', 5)],
  playBet: 100,
  result: 1,
  dealerQualified: true,
  antePayout: 200,
  blindPayout: 100 + 100 * 500, // royal flush blind
  playPayout: 200,
  tripsPayout: 0,
  totalPayout: 200 + (100 + 100 * 500) + 200,
  playerHandRank: 9,
  dealerHandRank: 0,
  message: '勝利！',
  messageCode: 'ultimatetexasholdem.result.playerWins',
  chips: 60000,
};

const endDealerWins: UltimateTexasHoldemResponse = {
  ...endPlayerWins,
  result: -1,
  dealerQualified: true,
  antePayout: 0,
  blindPayout: 0,
  playPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 1,
  message: 'ディーラー勝利！',
  messageCode: 'ultimatetexasholdem.result.dealerWins',
};

const endFold: UltimateTexasHoldemResponse = {
  ...riverState,
  phase: 5,
  folded: true,
  result: -1,
  message: 'フォールド',
  messageCode: 'ultimatetexasholdem.result.fold',
};

const endPush: UltimateTexasHoldemResponse = {
  ...endPlayerWins,
  result: 0,
  antePayout: 100,
  blindPayout: 100,
  playPayout: 100,
  totalPayout: 300,
  message: '引き分け！',
  messageCode: 'ultimatetexasholdem.result.push',
};

const endDealerNotQualifiedAntePushes: UltimateTexasHoldemResponse = {
  ...endPlayerWins,
  dealerQualified: false,
  // Ante pushes (returned only) while Blind and Play still pay out normally.
  antePayout: 100,
  blindPayout: 100 + 100 * 500,
  playPayout: 200,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('UltimateTexasHoldemPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<UltimateTexasHoldemPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows pre-flop with play 4x / 3x / check buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'プレイ 3×' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
  });

  it('shows flop with play 2x / check buttons after preflop check', async () => {
    mockApi.mockResolvedValueOnce(preFlopState).mockResolvedValueOnce(flopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 2×' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
  });

  it('shows river with play 1x / fold buttons after flop check', async () => {
    mockApi.mockResolvedValueOnce(flopState).mockResolvedValueOnce(riverState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 2×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 1×' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('sends multiplier 4 when Play 4x is pressed', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ 4×' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, undefined, 4));
  });

  it('sends multiplier 3 when Play 3x is pressed', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 3×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ 3×' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, undefined, 3));
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('shows dealer-qualified label at end phase', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ（ペア以上）/)).toBeInTheDocument());
  });

  it('shows dealer-not-qualified label at end phase', async () => {
    mockApi.mockResolvedValue(endDealerNotQualifiedAntePushes);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイなし/)).toBeInTheDocument());
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValue(endDealerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValue(endFold);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockApi.mockResolvedValue(endPush);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('changes ante and trips amounts when betting', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const tripsInput = screen.getByLabelText('トリップス');
    fireEvent.change(tripsInput, { target: { value: '20' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 20));
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders board and player cards on flop', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    // 2 player face-up + 3 community face-up + 2 dealer face-down (back image)
    await waitFor(() => expect(screen.getAllByRole('img').length).toBeGreaterThanOrEqual(5));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
    expect(screen.getByText('🃏')).toBeInTheDocument();
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });
});
