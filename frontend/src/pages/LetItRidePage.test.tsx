import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { letitrideApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, LetItRideResponse, MaskedCard } from '../types/card';
import { LetItRidePhase } from '../types/phases';
import { LetItRidePage } from './LetItRidePage';

vi.mock('../api/gameApi', () => ({
  letitrideApi: { exec: vi.fn() },
  actionLogApi: { letitride: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(letitrideApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard: MaskedCard = { design: '', value: 0 };

const betPhaseState: LetItRideResponse = {
  playerHand: [],
  communityCards: [],
  phase: LetItRidePhase.BET,
  chips: 1000,
  betAmount: 0,
  bet1Active: true,
  bet2Active: true,
  bet3Active: true,
  result: 0,
  handRank: 0,
  bet1Payout: 0,
  bet2Payout: 0,
  bet3Payout: 0,
  totalPayout: 0,
  message: '',
};

const firstDecisionState: LetItRideResponse = {
  ...betPhaseState,
  phase: LetItRidePhase.FIRST_DECISION,
  playerHand: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11)],
  communityCards: [maskedCard, maskedCard],
  betAmount: 100,
  chips: 700,
};

const secondDecisionState: LetItRideResponse = {
  ...firstDecisionState,
  phase: LetItRidePhase.SECOND_DECISION,
  communityCards: [card('CLOVER', 12), maskedCard],
  bet1Active: false,
};

const endPhaseWin: LetItRideResponse = {
  ...betPhaseState,
  phase: LetItRidePhase.END,
  playerHand: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11)],
  communityCards: [card('CLOVER', 12), card('SPADE', 13)],
  betAmount: 100,
  chips: 1300,
  result: 1,
  handRank: 9,
  bet1Active: true,
  bet2Active: true,
  bet3Active: true,
  bet1Payout: 100,
  bet2Payout: 200,
  bet3Payout: 300,
  totalPayout: 600,
  message: '勝利！',
  messageCode: 'letitride.result.win',
};

const endPhaseLoss: LetItRideResponse = {
  ...endPhaseWin,
  result: -1,
  chips: 700,
  bet1Payout: 0,
  bet2Payout: 0,
  bet3Payout: 0,
  totalPayout: 0,
  message: '役なし',
  messageCode: 'letitride.result.lose',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

afterEach(() => {
  localStorage.clear();
});

describe('LetItRidePage', () => {
  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<LetItRidePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders bet phase on mount with chips and bet button', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('applies min=10 and step=10 guardrails to the bet input', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    const input = screen.getByLabelText('ベット') as HTMLInputElement;
    expect(input).toHaveAttribute('min', '10');
    expect(input).toHaveAttribute('step', '10');
    // max is capped at 1/3 of the chip balance (1000 → 333).
    expect(input).toHaveAttribute('max', '333');
  });

  it('clamps a below-min bet up to the minimum of 10', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    const input = screen.getByLabelText('ベット') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '5' } });
    expect(input.value).toBe('10');
  });

  it('clamps an over-max bet down to the chip-capped maximum', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    const input = screen.getByLabelText('ベット') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '99999' } });
    // 1000 / 3 floored = 333.
    expect(input.value).toBe('333');
  });

  it('submits a valid clamped bet amount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    const input = screen.getByLabelText('ベット') as HTMLInputElement;
    fireEvent.change(input, { target: { value: '200' } });
    expect(input.value).toBe('200');
    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200));
  });

  it('shows payout reference panel in bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByText('配当表')).toBeInTheDocument();
  });

  it('explains the 3-bet split and pull-back rule in bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    const panel = screen.getByTestId('bet-structure');
    expect(panel).toBeInTheDocument();
    expect(panel).toHaveTextContent('なぜ3口に分かれるの？');
    // Accurate to the domain: 3x deduction caps a single bet at 1/3 of chips.
    expect(panel).toHaveTextContent('1/3');
    expect(panel).toHaveTextContent('ベット3を引き戻せます');
    expect(panel).toHaveTextContent('ベット2を引き戻せます');
    expect(panel).toHaveTextContent('ベット1は必ず残ります');
  });

  it('hides the bet-structure explanation outside the bet phase', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    expect(screen.queryByTestId('bet-structure')).not.toBeInTheDocument();
  });

  it('shows FIRST_DECISION phase with pull and letitride buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レットイットライド' })).toBeInTheDocument();
  });

  it('shows bet status indicators in FIRST_DECISION phase', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    expect(screen.getByTestId('bet-status')).toHaveTextContent('有効');
  });

  it('shows SECOND_DECISION phase with pull and letitride buttons', async () => {
    mockApi.mockResolvedValue(secondDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レットイットライド' })).toBeInTheDocument();
  });

  it('shows bet pulled status when a bet is withdrawn', async () => {
    mockApi.mockResolvedValue(secondDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    expect(screen.getByTestId('bet-status')).toHaveTextContent('引き戻し');
  });

  it('shows the per-bet amount and total current risk when all bets are live', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    // betAmount 100 in each of the 3 boxes.
    expect(screen.getByTestId('bet-box-bet1')).toHaveTextContent('100');
    // All three live → risk 300, and the active box carries the success ring.
    expect(screen.getByTestId('current-risk')).toHaveTextContent('300');
    expect(screen.getByTestId('bet-box-bet1').className).toContain('ring-ds-success');
  });

  it('dims a pulled bet box and lowers the current risk total', async () => {
    mockApi.mockResolvedValue(secondDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    // bet1 pulled → dimmed, risk drops to 2 × 100 = 200.
    expect(screen.getByTestId('bet-box-bet1').className).toContain('opacity-40');
    expect(screen.getByTestId('current-risk')).toHaveTextContent('200');
  });

  it('exposes the current risk total as a polite live region so pulls are announced', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    const risk = screen.getByTestId('current-risk');
    expect(risk).toHaveAttribute('role', 'status');
    expect(risk).toHaveAttribute('aria-live', 'polite');
  });

  it('shows END phase with reset button and payout breakdown on win', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('shows hand rank in END phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    // handRank=9 → 'handRank.9' → 'ロイヤルフラッシュ'
    await waitFor(() => expect(screen.getByText(/ロイヤルフラッシュ/)).toBeInTheDocument());
  });

  it('does not label an unevaluated hand rank as High Card', async () => {
    mockApi.mockResolvedValue({ ...endPhaseWin, handRank: -1 });
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // handRank of -1 (unevaluated) must not fall back to the "High Card" label.
    expect(screen.queryByText(/ハイカード/)).not.toBeInTheDocument();
  });

  it('shows individual bet payouts when non-zero', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    const breakdown = screen.getByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('ベット1: 100');
    expect(breakdown).toHaveTextContent('ベット2: 200');
    expect(breakdown).toHaveTextContent('ベット3: 300');
    expect(breakdown).toHaveTextContent('合計: 600');
  });

  it('hides zero-value bet payouts and shows total in END phase', async () => {
    mockApi.mockResolvedValue(endPhaseLoss);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    const breakdown = screen.getByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('合計: 0');
    // Individual bet lines with 0 should not render
    expect(breakdown).not.toHaveTextContent('ベット1:');
    expect(breakdown).not.toHaveTextContent('ベット2:');
    expect(breakdown).not.toHaveTextContent('ベット3:');
  });

  it('confirms before pulling, showing the return amount and new risk', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());

    // Clicking Pull opens a confirmation (does not immediately call the API).
    fireEvent.click(screen.getByRole('button', { name: 'プル' }));
    expect(screen.getByText('ベットを引き下げますか？')).toBeInTheDocument();
    // firstDecisionState: betAmount 100, all 3 active → risk 300, newRisk 200.
    expect(screen.getByText(/100 が戻り、総リスクは 300 → 200/)).toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalledWith('pull');

    mockApi.mockResolvedValue(secondDecisionState);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('pull'));
  });

  it('does not pull when the confirmation is cancelled', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'プル' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalledWith('pull');
  });

  it('opens the pull confirmation via the "p" keyboard shortcut', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());
    fireEvent.keyDown(document, { key: 'p' });
    expect(screen.getByText('ベットを引き下げますか？')).toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalledWith('pull');
  });

  it('disables keyboard shortcuts while the pull confirmation is open', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'プル' }));
    // With the dialog open, the 'l' shortcut must not bypass it.
    fireEvent.keyDown(document, { key: 'l' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalledWith('letitride');
  });

  it('calls execApi with letitride when letitride button clicked', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レットイットライド' })).toBeInTheDocument());

    mockApi.mockResolvedValue(secondDecisionState);
    fireEvent.click(screen.getByRole('button', { name: 'レットイットライド' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('letitride'));
  });

  it('calls execApi with bet and default amount when bet button clicked', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockApi.mockResolvedValue(firstDecisionState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('can change bet amount via input', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    const input = screen.getByLabelText('ベット');
    fireEvent.change(input, { target: { value: '200' } });

    mockApi.mockResolvedValue(firstDecisionState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200));
  });

  it('next game button at end phase fires reset without dialog', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    mockApi.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows network error alert', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('shows player hand section with cards in FIRST_DECISION phase', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
  });

  it('shows community card section in FIRST_DECISION phase', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByText('コミュニティカード')).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in decision phase', async () => {
    localStorage.setItem('hint_enabled_letitride', 'true');
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('renders CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    // CLI terminal uses role="log" and an input with aria-label for command prompt
    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('shows view action log button in END phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // The "view log" button (棋譜を見る) should also be present in END phase
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });
});
