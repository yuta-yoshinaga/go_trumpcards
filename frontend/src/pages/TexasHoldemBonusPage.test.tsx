import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { texasholdembonusApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, TexasHoldemBonusResponse } from '../types/card';
import { TexasHoldemBonusPage } from './TexasHoldemBonusPage';

vi.mock('../api/gameApi', () => ({
  texasholdembonusApi: { exec: vi.fn() },
  actionLogApi: { texasholdembonus: vi.fn() },
}));

const mockApi = vi.mocked(texasholdembonusApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: TexasHoldemBonusResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  bonusBet: 0,
  flopBet: 0,
  turnBet: 0,
  riverBet: 0,
  totalPlayBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  bonusPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const preFlopState: TexasHoldemBonusResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [maskedCard, maskedCard],
  community: [],
  anteBet: 100,
  chips: 900,
};

const flopState: TexasHoldemBonusResponse = {
  ...preFlopState,
  phase: 3,
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10)],
  flopBet: 200,
  totalPlayBet: 200,
  chips: 700,
};

const turnState: TexasHoldemBonusResponse = {
  ...flopState,
  phase: 4,
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2)],
};

const endPlayerWins: TexasHoldemBonusResponse = {
  ...flopState,
  phase: 5,
  dealerHand: [card('HEART', 7), card('DIAMOND', 5)],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2), card('HEART', 4)],
  result: 1,
  antePayout: 200 + 100 * 1000,
  playPayout: 400,
  totalPayout: 200 + 100 * 1000 + 400,
  playerHandRank: 9,
  dealerHandRank: 0,
  message: '勝利！',
  messageCode: 'texasholdembonus.result.playerWins',
  chips: 100400,
};

const endDealerWins: TexasHoldemBonusResponse = {
  ...endPlayerWins,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 1,
  message: 'ディーラー勝利！',
  messageCode: 'texasholdembonus.result.dealerWins',
};

const endFold: TexasHoldemBonusResponse = {
  ...preFlopState,
  phase: 5,
  community: [],
  result: -1,
  message: 'フォールド',
  messageCode: 'texasholdembonus.result.fold',
};

const endPush: TexasHoldemBonusResponse = {
  ...endPlayerWins,
  result: 0,
  antePayout: 100 + 100 * 1, // ante returned + straight ante bonus
  playPayout: 200,
  totalPayout: 100 + 100 * 1 + 200,
  message: '引き分け！',
  messageCode: 'texasholdembonus.result.push',
};

const endWithBonus: TexasHoldemBonusResponse = {
  ...endPlayerWins,
  bonusBet: 10,
  bonusPayout: 310, // 10 + 10*30 (AA)
  totalPayout: endPlayerWins.totalPayout + 310,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('TexasHoldemBonusPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('explains that ante and bonus payouts are judged independently', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<TexasHoldemBonusPage />);
    const intro = await screen.findByTestId('payout-intro');
    expect(intro).toHaveTextContent('独立して');
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<TexasHoldemBonusPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows pre-flop with play and fold buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(preFlopState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /プレイ/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('previews the Play bet cost (2× ante) on the pre-flop button', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<TexasHoldemBonusPage />);
    // anteBet 100 → flop Play bet = 2× = 200 chips.
    const playButton = await screen.findByTestId('thb-play-button');
    expect(playButton).toHaveTextContent('プレイ (2× = 200)');
  });

  it('shows flop with check and raise buttons', async () => {
    mockApi.mockResolvedValueOnce(preFlopState).mockResolvedValueOnce(flopState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /プレイ/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /プレイ/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /レイズ/ })).toBeInTheDocument();
  });

  it('previews the Raise bet cost (1× ante) on the flop and turn buttons', async () => {
    mockApi.mockResolvedValueOnce(flopState).mockResolvedValueOnce(turnState);
    renderWithProviders(<TexasHoldemBonusPage />);
    // anteBet 100 → each Raise = 1× = 100 chips, on both flop and turn.
    const flopRaise = await screen.findByTestId('thb-raise-button');
    expect(flopRaise).toHaveTextContent('レイズ (1× = 100)');

    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    const turnRaise = await screen.findByTestId('thb-raise-button');
    expect(turnRaise).toHaveTextContent('レイズ (1× = 100)');
  });

  it('shows turn after flop check', async () => {
    mockApi.mockResolvedValueOnce(flopState).mockResolvedValueOnce(turnState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('check'));
  });

  it('shows end phase with player wins (royal flush ante bonus)', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValue(endDealerWins);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('does not label an unevaluated hand rank as High Card', async () => {
    mockApi.mockResolvedValue({ ...endPlayerWins, playerHandRank: -1, dealerHandRank: -1 });
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    // Ranks of -1 (unevaluated) must not fall back to the handRank.0 "High Card" label.
    expect(screen.queryByText(/ハイカード/)).not.toBeInTheDocument();
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValue(endFold);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockApi.mockResolvedValue(endPush);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('shows payout breakdown including bonus side bet', async () => {
    mockApi.mockResolvedValue(endWithBonus);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/ボーナスサイド: 310/)).toBeInTheDocument();
  });

  it('changes ante and bonus amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const bonusInput = screen.getByLabelText('ボーナス');
    fireEvent.change(bonusInput, { target: { value: '10' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 10));
  });

  it('steps the ante and bonus amounts with the chip steppers', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Default ante 100 → +10 → 110; bonus 0 → +10 → 10.
    fireEvent.click(screen.getByRole('button', { name: 'アンテ +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ボーナス +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 110, 10));
  });

  it('disables the bonus minus stepper at zero', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ボーナス −10' })).toBeDisabled();
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders board and player cards in flop phase', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<TexasHoldemBonusPage />);
    // 2 dealer face-down cards are each announced as "hidden card".
    await waitFor(() => expect(screen.getAllByRole('img', { name: '非公開のカード' })).toHaveLength(2));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
    expect(screen.getByText('🃏')).toBeInTheDocument();
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /プレイ/ })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in pre-flop', async () => {
    localStorage.setItem('hint_enabled_texasholdembonus', 'true');
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('next game button executes reset', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<TexasHoldemBonusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });
});
