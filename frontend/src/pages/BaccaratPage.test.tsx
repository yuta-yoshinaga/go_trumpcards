import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, baccaratApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BaccaratResponse, Card, CardDesign } from '../types/card';
import { BaccaratPage } from './BaccaratPage';

vi.mock('../hooks/useGameHint');

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
  history: [],
  playerPairBet: 0,
  bankerPairBet: 0,
  sideBetResults: [],
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
  history: [0],
  playerPairBet: 0,
  bankerPairBet: 0,
  sideBetResults: [],
  message: 'プレイヤーの勝ち！',
  messageCode: 'baccarat.result.playerWins',
};

const endPhaseBankerWins: BaccaratResponse = {
  ...endPhasePlayerWins,
  result: -1,
  payout: 0,
  history: [1],
  message: 'バンカーの勝ち！',
  messageCode: 'baccarat.result.bankerWins',
};

const endPhaseTie: BaccaratResponse = {
  ...endPhasePlayerWins,
  result: 0,
  payout: 900,
  betType: 2,
  history: [2],
  message: '引き分け！',
  messageCode: 'baccarat.result.tie',
};

const errorState: BaccaratResponse = {
  ...betPhaseState,
  message: 'Invalid bet amount.',
};

const endPhaseWithSideBets: BaccaratResponse = {
  ...endPhasePlayerWins,
  playerPairBet: 10,
  bankerPairBet: 20,
  sideBetResults: [
    { betType: 1, resultType: 1, resultName: 'Pair', betAmount: 10, payout: 120 },
    { betType: 2, resultType: 0, resultName: '', betAmount: 20, payout: 0 },
  ],
};

const endPhaseWithHistory: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [0, 1, 0, 0, 1, 2, 1],
};

const endPhaseWithDragonTail: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [0, 0, 0, 0, 0, 0, 0, 1], // 7 player wins (dragon tail) + 1 banker
};

const endPhaseWithLeadingTie: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [2, 0, 1], // tie first, then player, then banker
};

const endPhaseWithOnlyTies: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [2, 2, 2], // only ties
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('BaccaratPage', () => {
  it('renders bet phase on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('does not expand card area with flex-1 during bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('ベットしてゲーム開始')).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).not.toHaveClass('flex-1');
  });

  it('expands card area with flex-1 during end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).toHaveClass('flex-1');
  });

  it('renders skeleton before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {})); // never resolves
    renderWithProviders(<BaccaratPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
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

    const amountInput = screen.getByLabelText('ベット額:');
    fireEvent.change(amountInput, { target: { value: '200' } });

    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '1' } }); // banker

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 1, 0, 0));
  });

  it('can set side bet amounts', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const ppInput = screen.getByLabelText('プレイヤーペア');
    fireEvent.change(ppInput, { target: { value: '10' } });

    const bpInput = screen.getByLabelText('バンカーペア');
    fireEvent.change(bpInput, { target: { value: '20' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 10, 20));
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

  // --- Big Road tests ---

  it('renders Big Road grid when history is present', async () => {
    mockExec.mockResolvedValue(endPhaseWithHistory);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
    expect(screen.getByText('罫線')).toBeInTheDocument();
    expect(screen.getByText('履歴クリア')).toBeInTheDocument();
  });

  it('does not render Big Road when history is empty', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.queryByTestId('big-road')).not.toBeInTheDocument();
    expect(screen.queryByText('罫線')).not.toBeInTheDocument();
  });

  it('renders Big Road with dragon tail (7+ same side)', async () => {
    mockExec.mockResolvedValue(endPhaseWithDragonTail);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
    // Dragon tail: 7 player wins should create overflow columns, then banker in separate column
    expect(screen.getByText('罫線')).toBeInTheDocument();
  });

  it('renders Big Road with leading tie (tie dropped silently)', async () => {
    mockExec.mockResolvedValue(endPhaseWithLeadingTie);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
  });

  it('does not render Big Road when history is only ties', async () => {
    mockExec.mockResolvedValue(endPhaseWithOnlyTies);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    // Only ties means columns is empty, so BigRoadGrid returns null
    // But history.length > 0, so the title and clear button should show
    expect(screen.queryByTestId('big-road')).not.toBeInTheDocument();
  });

  it('calls clearhistory when clear button is clicked', async () => {
    mockExec.mockResolvedValueOnce(endPhaseWithHistory).mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('履歴クリア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('履歴クリア'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('clearhistory'));
  });

  // --- Side bet tests ---

  it('shows side bet results in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseWithSideBets);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('side-bet-results')).toBeInTheDocument());
    expect(screen.getByText(/プレイヤーペア.*ペア.*配当: 120/)).toBeInTheDocument();
    expect(screen.getByText(/バンカーペア.*ペアなし.*配当: 0/)).toBeInTheDocument();
  });

  it('does not show side bet results when empty', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    expect(screen.queryByTestId('side-bet-results')).not.toBeInTheDocument();
  });

  // --- Keyboard navigation tests ---

  it('pressing b triggers bet in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 0, 0));
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

  it('renders tutorial button', async () => {
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // --- Hint UI tests ---

  it('shows hint tooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'banker', reason: 'hintReason.bankerBestOdds', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('hides hint tooltip when hintEnabled is false', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'banker', reason: 'hintReason.bankerBestOdds', confidence: 'strong' },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('hint settings checkbox calls setHintEnabled on toggle', async () => {
    const mockSetHintEnabled = vi.fn();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: mockSetHintEnabled });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    const checkbox = screen.getByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(checkbox);
    expect(mockSetHintEnabled).toHaveBeenCalledWith(true);
  });

  it('payout table is rendered as a collapsible details element in bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    const { container } = renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('配当表')).toBeInTheDocument());
    const details = container.querySelector('details');
    expect(details).toBeInTheDocument();
    const summary = details?.querySelector('summary');
    expect(summary).toHaveTextContent('配当表');
  });
});
