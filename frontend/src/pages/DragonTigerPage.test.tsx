import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { dragontigerApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DragonTigerResponse } from '../types/card';
import { DragonTigerBetType, DragonTigerHistoryResult, DragonTigerPhase } from '../types/phases';
import { DragonTigerPage } from './DragonTigerPage';

vi.mock('../api/gameApi', () => ({
  dragontigerApi: { exec: vi.fn() },
  actionLogApi: { dragontiger: vi.fn() },
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

const mockApi = vi.mocked(dragontigerApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const betState: DragonTigerResponse = {
  phase: DragonTigerPhase.BET,
  chips: 1000,
  betAmount: 0,
  betType: 0,
  result: 0,
  payout: 0,
  history: [],
  message: '',
};

const dragonWinState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 13 },
  tigerCard: { design: 'HEART', value: 5 },
  phase: DragonTigerPhase.END,
  chips: 1100,
  betAmount: 100,
  betType: DragonTigerBetType.DRAGON,
  result: 1,
  payout: 200,
  history: [DragonTigerHistoryResult.DRAGON],
  message: 'Dragon wins!',
  messageCode: 'dragontiger.result.dragonWins',
};

const tigerWinOnTigerBetState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 3 },
  tigerCard: { design: 'HEART', value: 13 },
  phase: DragonTigerPhase.END,
  chips: 1100,
  betAmount: 100,
  betType: DragonTigerBetType.TIGER,
  // GameResult wire value: -1 (GameResultLose = Tiger side won); the player took Tiger so they win.
  result: -1,
  payout: 200,
  history: [DragonTigerHistoryResult.TIGER],
  message: 'Tiger wins!',
  messageCode: 'dragontiger.result.tigerWins',
};

const tieWinOnTieBetState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 7 },
  tigerCard: { design: 'HEART', value: 7 },
  phase: DragonTigerPhase.END,
  chips: 1800,
  betAmount: 100,
  betType: DragonTigerBetType.TIE,
  result: 0,
  payout: 900,
  history: [DragonTigerHistoryResult.TIE],
  message: 'Tie! You win the tie bet.',
  messageCode: 'dragontiger.result.tieWin',
};

const tieRefundOnDragonBetState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 7 },
  tigerCard: { design: 'HEART', value: 7 },
  phase: DragonTigerPhase.END,
  chips: 950,
  betAmount: 100,
  betType: DragonTigerBetType.DRAGON,
  // Tie outcome: half-refund on a Dragon bet means payout < betAmount → no celebration.
  result: 0,
  payout: 50,
  history: [DragonTigerHistoryResult.TIE],
  message: 'Tie. Half of your bet is refunded.',
  messageCode: 'dragontiger.result.tieRefund',
};

describe('DragonTigerPage', () => {
  beforeEach(() => {
    mockApi.mockReset();
    mockApi.mockResolvedValue(betState);
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

  it('anchors the tutorial steps to the bet controls and card area', async () => {
    const { container } = renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="dt-bet-controls"]')).toBeInTheDocument());
    expect(container.querySelector('[data-tutorial="dt-cards"]')).toBeInTheDocument();
  });

  it('runs a tutorial that explains the Tie 8:1 payout', async () => {
    renderWithProviders(<DragonTigerPage />);
    const startBtn = await screen.findByRole('button', { name: 'チュートリアル' });
    fireEvent.click(startBtn);
    const dialog = await screen.findByRole('dialog');
    expect(dialog).toHaveTextContent('8:1');
  });

  it('issues a reset on mount', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders three bet buttons in the bet phase', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ドラゴン' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'タイガー' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /タイ \(8:1\)/ })).toBeInTheDocument();
  });

  it('dispatches a Dragon bet on button click', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ドラゴン' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ドラゴン' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.DRAGON));
  });

  it('dispatches a Tiger bet on button click', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'タイガー' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'タイガー' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.TIGER));
  });

  it('dispatches a Tie bet on button click', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /タイ \(8:1\)/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /タイ \(8:1\)/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.TIE));
  });

  it('renders dragon and tiger cards + Big Road in the end phase', async () => {
    mockApi.mockResolvedValueOnce(dragonWinState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('ドラゴン')).toBeInTheDocument());
    expect(screen.getByText('タイガー')).toBeInTheDocument();
    expect(screen.getByTestId('bigroad')).toBeInTheDocument();
    expect(screen.getByText('払戻し: 200')).toBeInTheDocument();
  });

  // Regression coverage for the gemini/Claude review: winShow must use
  // payout > betAmount, not result===Win-only, otherwise Tiger and Tie wins
  // are silently uncelebrated.
  it('renders payout for a Tiger-bet Tiger-win (player wins)', async () => {
    mockApi.mockResolvedValueOnce(tigerWinOnTigerBetState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('払戻し: 200')).toBeInTheDocument());
    expect(screen.getByTestId('bigroad')).toBeInTheDocument();
  });

  it('renders payout for a Tie-bet Tie outcome (8:1 win)', async () => {
    mockApi.mockResolvedValueOnce(tieWinOnTieBetState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('払戻し: 900')).toBeInTheDocument());
  });

  it('renders payout for a Dragon-bet half-refund tie (no celebration expected)', async () => {
    mockApi.mockResolvedValueOnce(tieRefundOnDragonBetState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('払戻し: 50')).toBeInTheDocument());
  });

  it('shows the payout breakdown — result, ×1 odds badge, and a green profit — for a Dragon win', async () => {
    mockApi.mockResolvedValueOnce(dragonWinState); // bet Dragon 100, payout 200
    renderWithProviders(<DragonTigerPage />);
    const breakdown = await screen.findByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('ドラゴンの勝ち');
    expect(breakdown).toHaveTextContent('ドラゴン ×1');
    const diff = screen.getByTestId('payout-diff');
    expect(diff).toHaveTextContent('+100');
    expect(diff).toHaveClass('text-ds-success');
  });

  it('shows the tiger-win result and ×1 badge for a winning Tiger bet', async () => {
    mockApi.mockResolvedValueOnce(tigerWinOnTigerBetState); // result -1, bet Tiger 100, payout 200
    renderWithProviders(<DragonTigerPage />);
    const breakdown = await screen.findByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('タイガーの勝ち');
    expect(breakdown).toHaveTextContent('タイガー ×1');
    const diff = screen.getByTestId('payout-diff');
    expect(diff).toHaveTextContent('+100');
    expect(diff).toHaveClass('text-ds-success');
  });

  it('shows the ×8 odds badge and a big green profit for a Tie-bet win', async () => {
    mockApi.mockResolvedValueOnce(tieWinOnTieBetState); // bet Tie 100, payout 900
    renderWithProviders(<DragonTigerPage />);
    const breakdown = await screen.findByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('タイ ×8');
    expect(screen.getByTestId('payout-result')).toHaveTextContent('的中'); // tieWin text
    const diff = screen.getByTestId('payout-diff');
    expect(diff).toHaveTextContent('+800');
    expect(diff).toHaveClass('text-ds-success');
  });

  it('distinguishes a tie refund with a red loss diff for a Dragon bet', async () => {
    mockApi.mockResolvedValueOnce(tieRefundOnDragonBetState); // bet Dragon 100, payout 50
    renderWithProviders(<DragonTigerPage />);
    const diff = await screen.findByTestId('payout-diff');
    expect(diff).toHaveTextContent('-50');
    expect(diff).toHaveClass('text-ds-error');
    expect(screen.getByTestId('payout-result')).toHaveTextContent('返還'); // tieRefund text
  });

  it('shows a Rebet button at end-phase after a bet, replaying the same amount and target', async () => {
    renderWithProviders(<DragonTigerPage />);
    const dragonBtn = await screen.findByRole('button', { name: 'ドラゴン' });
    mockApi.mockClear();
    mockApi.mockResolvedValue(dragonWinState); // bet Dragon 100 → END
    fireEvent.click(dragonBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.DRAGON));
    const rebet = await screen.findByTestId('dt-rebet-button');
    // The button advertises the previous target (Dragon) and amount (100).
    expect(rebet).toHaveTextContent('ドラゴン');
    expect(rebet).toHaveTextContent('100');
    expect(rebet).toHaveAttribute('aria-keyshortcuts', 'e');

    mockApi.mockClear();
    mockApi.mockResolvedValueOnce(betState);
    mockApi.mockResolvedValueOnce(dragonWinState);
    fireEvent.click(rebet);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.DRAGON));
  });

  it('does not show the Rebet button at end-phase when no prior bet exists (state seeded at END)', async () => {
    mockApi.mockResolvedValueOnce(dragonWinState); // mount lands directly at END, no bet snapshotted
    renderWithProviders(<DragonTigerPage />);
    await screen.findByTestId('payout-breakdown');
    expect(screen.queryByTestId('dt-rebet-button')).not.toBeInTheDocument();
  });

  it('hides the Rebet button when chips are insufficient to replay', async () => {
    renderWithProviders(<DragonTigerPage />);
    const dragonBtn = await screen.findByRole('button', { name: 'ドラゴン' });
    // End state with only 50 chips left — cannot afford the 100 rebet.
    mockApi.mockResolvedValue({ ...dragonWinState, chips: 50 });
    fireEvent.click(dragonBtn);
    await screen.findByTestId('payout-breakdown');
    expect(screen.queryByTestId('dt-rebet-button')).not.toBeInTheDocument();
  });

  it("the 'e' keyboard shortcut replays the last bet at end phase", async () => {
    renderWithProviders(<DragonTigerPage />);
    const tigerBtn = await screen.findByRole('button', { name: 'タイガー' });
    mockApi.mockResolvedValue(tigerWinOnTigerBetState); // bet Tiger 100 → END
    fireEvent.click(tigerBtn);
    await screen.findByTestId('dt-rebet-button');

    mockApi.mockClear();
    mockApi.mockResolvedValueOnce(betState);
    mockApi.mockResolvedValueOnce(tigerWinOnTigerBetState);
    fireEvent.keyDown(document, { key: 'e' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.TIGER));
  });
});
