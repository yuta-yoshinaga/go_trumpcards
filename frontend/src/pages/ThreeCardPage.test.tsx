import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, threecardApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ThreeCardResponse } from '../types/card';
import { ThreeCardPage } from './ThreeCardPage';

vi.mock('../hooks/useGameHint');

vi.mock('../api/gameApi', () => ({
  threecardApi: { exec: vi.fn() },
  actionLogApi: { threecard: vi.fn() },
}));

const mockExec = vi.mocked(threecardApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: ThreeCardResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  pairPlusBet: 0,
  playBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  anteBonusPayout: 0,
  pairPlusPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const actionPhaseState: ThreeCardResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13)],
  dealerHand: [],
  anteBet: 100,
  chips: 900,
};

const endPhasePlayerWins: ThreeCardResponse = {
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 2), card('HEART', 7)],
  phase: 3,
  chips: 1200,
  anteBet: 100,
  pairPlusBet: 0,
  playBet: 100,
  result: 1,
  antePayout: 200,
  playPayout: 200,
  anteBonusPayout: 0,
  pairPlusPayout: 0,
  totalPayout: 400,
  dealerQualified: true,
  playerHandRank: 1,
  dealerHandRank: 1,
  message: '勝利！',
  messageCode: 'threecard.result.playerWins',
};

const endPhaseDealerWins: ThreeCardResponse = {
  ...endPhasePlayerWins,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  message: 'ディーラー勝利！',
  messageCode: 'threecard.result.dealerWins',
};

const endPhaseFold: ThreeCardResponse = {
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
  messageCode: 'threecard.result.fold',
};

const endPhasePush: ThreeCardResponse = {
  ...endPhasePlayerWins,
  result: 0,
  antePayout: 100,
  playPayout: 100,
  totalPayout: 200,
  message: '引き分け！',
  messageCode: 'threecard.result.push',
};

const endPhaseDealerNotQualified: ThreeCardResponse = {
  ...endPhasePlayerWins,
  dealerQualified: false,
  antePayout: 100,
  playPayout: 0,
  totalPayout: 100,
  message: 'ディーラー未クオリファイ！',
  messageCode: 'threecard.result.dealerNotQualified',
};

const endPhaseWithPairPlus: ThreeCardResponse = {
  ...endPhasePlayerWins,
  pairPlusBet: 50,
  pairPlusPayout: 200,
  anteBonusPayout: 100,
  totalPayout: 700,
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('ThreeCardPage', () => {
  it('renders bet phase on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {})); // never resolves
    renderWithProviders(<ThreeCardPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('does not expand card area with flex-1 during bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).not.toHaveClass('flex-1');
  });

  it('expands card area with flex-1 during action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).toHaveClass('flex-1');
  });

  it('shows action phase with play and fold buttons', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows the ante and play-required amounts during action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByTestId('action-bet-slip')).toBeInTheDocument());
    const slip = screen.getByTestId('action-bet-slip');
    expect(slip).toHaveTextContent('アンテ: 100');
    expect(slip).toHaveTextContent('プレイに必要: 100');
    // pairPlusBet is 0 in this fixture → the pair-plus row is omitted
    expect(slip).not.toHaveTextContent('ペアプラス');
  });

  it('shows the pair-plus row in the action bet slip when pair plus is wagered', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, pairPlusBet: 50 });
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByTestId('action-bet-slip')).toBeInTheDocument());
    expect(screen.getByTestId('action-bet-slip')).toHaveTextContent('ペアプラス: 50');
  });

  it('does not show the action bet slip during bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.queryByTestId('action-bet-slip')).not.toBeInTheDocument();
  });

  it('shows end phase with player wins', async () => {
    mockExec.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ' }));
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockExec.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ' }));
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockExec.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseFold);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockExec.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePush);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ' }));
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('shows end phase with dealer not qualified', async () => {
    mockExec.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerNotQualified);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ' }));
    await waitFor(() => expect(screen.getAllByText(/未クオリファイ/).length).toBeGreaterThanOrEqual(1));
  });

  it('shows payout breakdown with pair plus and ante bonus', async () => {
    mockExec.mockResolvedValue(endPhaseWithPairPlus);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/アンテボーナス: 100/)).toBeInTheDocument();
    expect(screen.getByText(/ペアプラス: 200/)).toBeInTheDocument();
    expect(screen.getByText(/合計: 700/)).toBeInTheDocument();
  });

  it('can change ante and pair plus amounts', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const ppInput = screen.getByLabelText('ペアプラス');
    fireEvent.change(ppInput, { target: { value: '50' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 50));
  });

  it('steps the ante and pair plus amounts with the chip steppers', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Default ante 100 → +10 → 110; pair plus 0 → +10 → 10.
    fireEvent.click(screen.getByRole('button', { name: 'アンテ +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ペアプラス +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 110, 10));
  });

  it('disables the pair plus minus stepper at zero', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ペアプラス −10' })).toBeDisabled();
  });

  it('resets after end phase', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(endPhasePlayerWins)
      .mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ── Next game button at end phase ──────────────────────────────────────────

  it('next game button triggers reset without confirm dialog', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows network error', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('shows action log', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    vi.mocked(actionLogApi.threecard).mockResolvedValue({ entries: [] as never[] });
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('renders player and dealer cards in end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(6));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  it('shows hand rank names in end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getAllByText(/ハイカード/).length).toBeGreaterThanOrEqual(1));
  });

  it('shows dealer qualification status', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ/)).toBeInTheDocument());
  });

  // --- Keyboard navigation tests ---

  it('pressing b triggers bet in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0));
  });

  it('pressing p triggers play in ACTION phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.keyDown(document, { key: 'p' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play'));
  });

  it('pressing f triggers fold in ACTION phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseFold);
    fireEvent.keyDown(document, { key: 'f' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('pressing r triggers reset in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText(/勝利/)).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.keyDown(document, { key: 'r' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('pressing b does not trigger bet in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText(/勝利/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('pressing r does not trigger reset in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'r' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // --- Payout table tests ---

  it('shows payout table in bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('配当表')).toBeInTheDocument());
  });

  it('payout table contains ante bonus and pair plus sections', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('配当表')).toBeInTheDocument());
    expect(screen.getByText('アンテボーナス')).toBeInTheDocument();
    // 'ペアプラス' appears in both the payout header and bet label; check for payout-specific entries
    expect(screen.getByText('ペア: 1:1')).toBeInTheDocument();
  });

  it('does not show payout table in action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    expect(screen.queryByText('配当表')).not.toBeInTheDocument();
  });

  // --- Hint UI tests ---

  it('shows hint tooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'play', reason: 'hintReason.strongHand', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('hides hint tooltip when hintEnabled is false', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'play', reason: 'hintReason.strongHand', confidence: 'strong' },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ' })).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('hint settings checkbox calls setHintEnabled on toggle', async () => {
    const mockSetHintEnabled = vi.fn();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: mockSetHintEnabled });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    const checkbox = screen.getByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(checkbox);
    expect(mockSetHintEnabled).toHaveBeenCalledWith(true);
  });

  // --- Rebet (same-amount replay) tests ---

  it('shows a rebet button in end phase after a bet and replays the prior ante', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByTestId('tc-rebet-button')).toBeInTheDocument());
    // Default ante 100 + pair plus 0 → the button advertises a total of 100.
    expect(screen.getByTestId('tc-rebet-button')).toHaveTextContent('100');

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    fireEvent.click(screen.getByTestId('tc-rebet-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0));
  });

  it('rebet replays the previous ante and pair plus amounts', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '200' } });
    fireEvent.change(screen.getByLabelText('ペアプラス'), { target: { value: '50' } });
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByTestId('tc-rebet-button')).toBeInTheDocument());
    // 200 ante + 50 pair plus → advertised total of 250.
    expect(screen.getByTestId('tc-rebet-button')).toHaveTextContent('250');

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    fireEvent.click(screen.getByTestId('tc-rebet-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 50));
  });

  it('does not show a rebet button in end phase when no prior bet exists', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByTestId('tc-rebet-button')).not.toBeInTheDocument();
  });

  it('hides the rebet button when chips are below the previous bet', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce({ ...endPhaseDealerWins, chips: 50 });
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // Snapshotted ante was 100 but only 50 chips remain → rebet is unaffordable.
    expect(screen.queryByTestId('tc-rebet-button')).not.toBeInTheDocument();
  });

  it('pressing n replays the previous bet in END phase', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByTestId('tc-rebet-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    fireEvent.keyDown(document, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0));
  });

  it('pressing n does nothing in END phase without a prior bet', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<ThreeCardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'n' });
    expect(mockExec).not.toHaveBeenCalled();
  });
});
