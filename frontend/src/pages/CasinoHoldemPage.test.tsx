import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { casinoholdemApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CasinoHoldemResponse } from '../types/card';
import { CasinoHoldemPage } from './CasinoHoldemPage';

vi.mock('../api/gameApi', () => ({
  casinoholdemApi: { exec: vi.fn() },
  actionLogApi: { casinoholdem: vi.fn() },
}));

const mockApi = vi.mocked(casinoholdemApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: CasinoHoldemResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  bonusBet: 0,
  callBet: 0,
  result: 0,
  dealerQualify: false,
  antePayout: 0,
  callPayout: 0,
  bonusPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const flopState: CasinoHoldemResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [maskedCard, maskedCard],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10)],
  anteBet: 100,
  chips: 900,
};

const endPlayerWins: CasinoHoldemResponse = {
  ...flopState,
  phase: 3,
  dealerHand: [card('HEART', 7), card('DIAMOND', 5)],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2), card('HEART', 4)],
  callBet: 200,
  result: 1,
  dealerQualify: true,
  antePayout: 100 + 100 * 100, // Royal flush ante
  callPayout: 400,
  totalPayout: 100 + 100 * 100 + 400,
  playerHandRank: 9,
  dealerHandRank: 0,
  message: '勝利！',
  messageCode: 'casinoholdem.result.playerWins',
  chips: 10800,
};

const endDealerWins: CasinoHoldemResponse = {
  ...endPlayerWins,
  result: -1,
  dealerQualify: true,
  antePayout: 0,
  callPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 1,
  message: 'ディーラー勝利！',
  messageCode: 'casinoholdem.result.dealerWins',
};

const endFold: CasinoHoldemResponse = {
  ...flopState,
  phase: 3,
  result: -1,
  message: 'フォールド',
  messageCode: 'casinoholdem.result.fold',
};

const endNoQualify: CasinoHoldemResponse = {
  ...endPlayerWins,
  dealerQualify: false,
  callPayout: 200, // call pushes
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('CasinoHoldemPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<CasinoHoldemPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('toggles the hint tooltip through the settings panel', async () => {
    mockApi.mockResolvedValue({ ...flopState, playerHandRank: 1 });
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /コール/ })).toBeInTheDocument());

    // Hint is off by default, so no tooltip renders.
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    // Open the shared settings panel and enable the hint via its checkbox.
    fireEvent.click(screen.getByText('設定'));
    const checkbox = screen.getByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(checkbox);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());

    // Disabling it hides the hint again.
    fireEvent.click(checkbox);
    await waitFor(() => expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument());
  });

  it('shows flop with call and fold buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /コール/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByText('ディーラークオリファイ')).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValue(endDealerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValue(endFold);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows dealer no-qualify message after call', async () => {
    mockApi.mockResolvedValue(endNoQualify);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('ディーラーノークオリファイ')).toBeInTheDocument());
  });

  it('changes ante and bonus amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const bonusInput = screen.getByLabelText('AAボーナス');
    fireEvent.change(bonusInput, { target: { value: '10' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 10));
  });

  it('shows a validation error and disables Bet for a non-multiple-of-10 ante', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '15' } });
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
  });

  it('steps the ante up and down with the stepper buttons', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const ante = screen.getByLabelText('アンテ') as HTMLInputElement;
    fireEvent.click(screen.getByRole('button', { name: 'アンテ +10' }));
    expect(ante.value).toBe('110');
    fireEvent.click(screen.getByRole('button', { name: 'アンテ −10' }));
    expect(ante.value).toBe('100');
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders board and player cards in flop phase', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    // 2 player face-up + 3 community face-up + 2 dealer face-down = 7 imgs
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(7));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
    expect(screen.getByText('🃏')).toBeInTheDocument();
  });

  it('shows the player current hand at flop (computed client-side)', async () => {
    // flopState is A-K-Q-J-10 all spades → Royal Flush.
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    const hand = await screen.findByTestId('ch-flop-hand');
    expect(hand).toHaveTextContent('現在の役');
    expect(hand).toHaveTextContent('ロイヤルフラッシュ');
  });

  it('shows a non-royal current hand at flop (one pair)', async () => {
    // A♠ + 10♣ with board 10♥ 4♦ 2♠ → a pair of tens.
    const onePairFlop: CasinoHoldemResponse = {
      ...flopState,
      playerHand: [card('SPADE', 1), card('CLOVER', 10)],
      community: [card('HEART', 10), card('DIAMOND', 4), card('SPADE', 2)],
    };
    mockApi.mockResolvedValue(onePairFlop);
    renderWithProviders(<CasinoHoldemPage />);
    const hand = await screen.findByTestId('ch-flop-hand');
    expect(hand).toHaveTextContent('ワンペア');
  });

  it('does not show the flop current-hand readout in the bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ベット/ })).toBeInTheDocument());
    expect(screen.queryByTestId('ch-flop-hand')).not.toBeInTheDocument();
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /コール/ })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled at flop', async () => {
    localStorage.setItem('hint_enabled_casinoholdem', 'true');
    mockApi.mockResolvedValue({ ...flopState, playerHandRank: 1 });
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('next game button executes reset', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });
});
