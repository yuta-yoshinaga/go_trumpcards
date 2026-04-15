import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { letitrideApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
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

  it('shows payout reference panel in bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByText('配当表')).toBeInTheDocument();
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

  it('shows END phase with reset button and payout breakdown on win', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('shows hand rank in END phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    // handRank=9 → 'handRank.9' → 'ロイヤルフラッシュ'
    await waitFor(() => expect(screen.getByText(/ロイヤルフラッシュ/)).toBeInTheDocument());
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

  it('calls execApi with pull when pull button clicked', async () => {
    mockApi.mockResolvedValue(firstDecisionState);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プル' })).toBeInTheDocument());

    mockApi.mockResolvedValue(secondDecisionState);
    fireEvent.click(screen.getByRole('button', { name: 'プル' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('pull'));
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

  it('shows confirm dialog when reset is clicked', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('calls reset on confirm in dialog', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<LetItRidePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    mockApi.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
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
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    // The "view log" button (棋譜を見る) should also be present in END phase
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });
});
