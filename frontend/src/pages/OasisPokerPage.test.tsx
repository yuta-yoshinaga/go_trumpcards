import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { oasispokerApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, OasisPokerResponse } from '../types/card';
import { OasisPokerPage } from './OasisPokerPage';

vi.mock('../api/gameApi', () => ({
  oasispokerApi: { exec: vi.fn() },
  actionLogApi: { oasispoker: vi.fn() },
}));

const mockApi = vi.mocked(oasispokerApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: OasisPokerResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  jackpotBet: 0,
  exchangeCount: 0,
  exchangeFee: 0,
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

const exchangePhaseState: OasisPokerResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const actionPhaseState: OasisPokerResponse = {
  ...exchangePhaseState,
  phase: 3,
  exchangeCount: 1,
  exchangeFee: 100,
  chips: 800,
};

const endPhasePlayerWins: OasisPokerResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: 4,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
  exchangeCount: 1,
  exchangeFee: 100,
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
  messageCode: 'oasispoker.result.playerWins',
};

const endPhaseFold: OasisPokerResponse = {
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
  messageCode: 'oasispoker.result.fold',
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('OasisPokerPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<OasisPokerPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows the dealer-qualification-pending note during the exchange phase', async () => {
    mockApi.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByTestId('dealer-qualify-pending')).toBeInTheDocument());
  });

  it('labels the masked dealer cards as hidden for assistive tech', async () => {
    mockApi.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<OasisPokerPage />);
    // 4 masked dealer cards each announced as "hidden card" (1 dealer card is face-up).
    const hidden = await screen.findAllByRole('img', { name: '非公開のカード' });
    expect(hidden).toHaveLength(4);
  });

  it('hides the pending note and shows the qualification state at the end phase', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByText('クオリファイ')).toBeInTheDocument());
    expect(screen.queryByTestId('dealer-qualify-pending')).not.toBeInTheDocument();
  });

  it('transitions bet → exchange and shows exchange UI', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(exchangePhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'ステイ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument();
  });

  it('exchange button is disabled until a card is selected', async () => {
    mockApi.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const exchangeBtn = screen.getByRole('button', { name: '交換' });
    expect(exchangeBtn).toBeDisabled();

    // Click a card to select it
    fireEvent.click(screen.getByTestId('player-card-0'));
    expect(exchangeBtn).not.toBeDisabled();
  });

  it('clicking a selected card toggles it off', async () => {
    mockApi.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const card0 = screen.getByTestId('player-card-0');
    fireEvent.click(card0);
    expect(card0).toHaveAttribute('data-selected', 'true');
    fireEvent.click(card0);
    expect(card0).toHaveAttribute('data-selected', 'false');
  });

  it('exchange sends selected indices to the API', async () => {
    mockApi.mockResolvedValueOnce(exchangePhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('player-card-0'));
    fireEvent.click(screen.getByTestId('player-card-2'));
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('exchange', undefined, undefined, [0, 2]));
  });

  it('stand transitions to action phase without exchanging', async () => {
    mockApi.mockResolvedValueOnce(exchangePhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ステイ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ステイ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('stand'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
  });

  it('action phase shows call/fold buttons', async () => {
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('end phase player wins shows payout breakdown', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('end phase shows fold message', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseFold);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('can change ante and jackpot amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const jpInput = screen.getByLabelText('ジャックポット');
    fireEvent.change(jpInput, { target: { value: '10' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 10));
  });

  it('shows a play hint in the action phase when hints are enabled (strong hand)', async () => {
    const strongAction: OasisPokerResponse = { ...actionPhaseState, playerHandRank: 1 };
    mockApi.mockResolvedValue(strongAction);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    // No hint until the toggle is enabled.
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(screen.getByLabelText('ヒント表示'));
    expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('ペア以上 — プレイ推奨');
  });

  it('shows a fold hint in the action phase for a weak hand when hints are enabled', async () => {
    // exchangePhaseState hand (10/J/K/5/7 offsuit) has rank 0 and no Ace-King pair.
    const weakAction: OasisPokerResponse = { ...actionPhaseState, playerHandRank: 0 };
    mockApi.mockResolvedValue(weakAction);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('ヒント表示'));
    expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('弱い手札 — フォールド推奨');
  });

  it('shows an exchange hint in the exchange phase for a weak hand when hints are enabled', async () => {
    mockApi.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('ヒント表示'));
    expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('弱い手札 — カード交換で改善を狙う');
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<OasisPokerPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
