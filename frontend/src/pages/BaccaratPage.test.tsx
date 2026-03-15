import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, baccaratApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BaccaratResponse, Card, CardDesign } from '../types/card';
import { BaccaratPage } from './BaccaratPage';

vi.mock('../api/gameApi', () => ({
  baccaratApi: { exec: vi.fn() },
  actionLogApi: { baccarat: vi.fn() },
}));

const mockExec = vi.mocked(baccaratApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: BaccaratResponse = {
  playerHand: [],
  bankerHand: [],
  playerHandValue: 0,
  bankerHandValue: 0,
  phase: 1,
  chips: 1000,
  betAmount: 0,
  betType: 0,
  result: 0,
  payout: 0,
  message: '',
};

const endPhasePlayerWins: BaccaratResponse = {
  playerHand: [card('SPADE', 9), card('HEART', 3)],
  bankerHand: [card('CLOVER', 5), card('DIAMOND', 2)],
  playerHandValue: 2,
  bankerHandValue: 7,
  phase: 2,
  chips: 1100,
  betAmount: 100,
  betType: 0,
  result: 1,
  payout: 200,
  message: 'プレイヤーの勝ち！',
  messageCode: 'baccarat.result.playerWins',
};

const endPhaseBankerWins: BaccaratResponse = {
  ...endPhasePlayerWins,
  result: -1,
  payout: 0,
  message: 'バンカーの勝ち！',
  messageCode: 'baccarat.result.bankerWins',
};

const endPhaseTie: BaccaratResponse = {
  ...endPhasePlayerWins,
  result: 0,
  payout: 900,
  betType: 2,
  message: '引き分け！',
  messageCode: 'baccarat.result.tie',
};

const errorState: BaccaratResponse = {
  ...betPhaseState,
  message: 'Invalid bet amount.',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('BaccaratPage', () => {
  it('renders bet phase on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('returns null before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {})); // never resolves
    const { container } = renderWithProviders(<BaccaratPage />);
    expect(container.firstChild).toBeNull();
  });

  it('shows end phase with player wins', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    expect(screen.getByText('配当: 200')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('shows end phase with banker wins', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhaseBankerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('バンカーの勝ち！')).toBeInTheDocument());
  });

  it('shows end phase with tie', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhaseTie);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
    expect(screen.getByText('配当: 900')).toBeInTheDocument();
  });

  it('can change bet amount and bet type', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const amountInput = screen.getByRole('spinbutton');
    fireEvent.change(amountInput, { target: { value: '200' } });

    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '1' } }); // banker

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 1));
  });

  it('resets after end phase', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(endPhasePlayerWins)
      .mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(endPhasePlayerWins)
      .mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows error message', async () => {
    mockExec.mockResolvedValue(errorState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('Invalid bet amount.')).toBeInTheDocument());
  });

  it('shows network error', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('shows action log', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    vi.mocked(actionLogApi.baccarat).mockResolvedValue({ entries: [] as never[] });
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('renders player and banker cards', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(4));
    expect(screen.getByText(/値: 2/)).toBeInTheDocument();
    expect(screen.getByText(/値: 7/)).toBeInTheDocument();
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  // --- Keyboard navigation tests ---

  it('pressing b triggers bet in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0));
  });

  it('pressing r triggers reset in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤーの勝ち/)).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.keyDown(document, { key: 'r' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('pressing b does not trigger bet in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤーの勝ち/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('pressing r does not trigger reset in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'r' });
    expect(mockExec).not.toHaveBeenCalled();
  });
});
