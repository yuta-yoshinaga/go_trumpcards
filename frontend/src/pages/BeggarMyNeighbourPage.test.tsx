import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beggarmyneighbourApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BeggarMyNeighbourResponse } from '../types/card';
import { BeggarMyNeighbourPhase } from '../types/phases';
import { BeggarMyNeighbourPage } from './BeggarMyNeighbourPage';

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

const mockUseCliMode = vi.mocked(useCliMode);

vi.mock('../api/gameApi', () => ({
  beggarmyneighbourApi: { exec: vi.fn() },
  actionLogApi: { beggarmyneighbour: vi.fn() },
}));

const mockExec = vi.mocked(beggarmyneighbourApi.exec);

const baseState: BeggarMyNeighbourResponse = {
  players: [
    { id: 0, isHuman: true, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
    { id: 1, isHuman: false, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
  ],
  phase: BeggarMyNeighbourPhase.PLAY,
  gameEndFlag: false,
  winnerIdx: -1,
  currentPlayerIdx: 0,
  penaltyOwnerIdx: -1,
  penaltyRemaining: 0,
  centralPileSize: 0,
  lastCardPlayed: null,
  roundsPlayed: 0,
  config: { maxRounds: 2000 },
  message: '',
  messageCode: 'play',
};

const penaltyState: BeggarMyNeighbourResponse = {
  ...baseState,
  phase: BeggarMyNeighbourPhase.PAY_PENALTY,
  currentPlayerIdx: 1,
  penaltyOwnerIdx: 0,
  penaltyRemaining: 3,
  centralPileSize: 3,
  lastCardPlayed: { design: 'SPADE', value: 13 },
  messageCode: 'penalty',
};

const collectState: BeggarMyNeighbourResponse = {
  ...baseState,
  phase: BeggarMyNeighbourPhase.COLLECT,
  penaltyOwnerIdx: 0,
  centralPileSize: 5,
  messageCode: 'collect',
};

const gameEndState: BeggarMyNeighbourResponse = {
  ...baseState,
  phase: BeggarMyNeighbourPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { id: 0, isHuman: true, drawPileSize: 0, discardPileSize: 52, totalCards: 52 },
    { id: 1, isHuman: false, drawPileSize: 0, discardPileSize: 0, totalCards: 0 },
  ],
  messageCode: 'gameEnd',
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
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

describe('BeggarMyNeighbourPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders pile info after state loads', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    const pileLines = screen.getAllByText(/26/);
    expect(pileLines.length).toBeGreaterThan(0);
  });

  it('step button calls exec with step', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('autoplay button calls exec with autoplay', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('autoplay-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('autoplay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autoplay'));
  });

  it('shows penalty remaining during PAY_PENALTY phase', async () => {
    mockExec.mockResolvedValueOnce(penaltyState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getAllByText(/3/).length).toBeGreaterThan(0);
  });

  it('shows central pile stack when centralPileSize > 0', async () => {
    mockExec.mockResolvedValueOnce(penaltyState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    const stack = await screen.findByTestId('bmn-central-pile-stack');
    expect(stack).toHaveAttribute('data-pile-size', '3');
  });

  it('announces the current phase in a polite live region', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />); // baseState: PLAY phase
    const region = await screen.findByTestId('bmn-phase-announce');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('フェーズ: プレイ中');
  });

  it('announces the penalty countdown during PAY_PENALTY', async () => {
    mockExec.mockResolvedValueOnce(penaltyState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    const region = await screen.findByTestId('bmn-phase-announce');
    await waitFor(() => expect(region).toHaveTextContent('フェーズ: ペナルティ支払い中。残りペナルティ 3 枚'));
  });

  it('conveys pile counts via accessible text and hides the decorative visuals', async () => {
    mockExec.mockResolvedValueOnce(penaltyState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    // Counts are readable as plain text (no redundant aria-label on the card visuals).
    await waitFor(() => expect(screen.getByText(/場の山.*3/)).toBeInTheDocument());
    expect(screen.getAllByText(/山札.*26/).length).toBeGreaterThan(0);
    // The decorative pile stack is hidden from assistive tech.
    const stack = screen.getByTestId('bmn-central-pile-stack');
    expect(stack).toHaveAttribute('aria-hidden', 'true');
  });

  it('renders a role=alert ErrorAlert (not a bare button) on error', async () => {
    // Mount succeeds so the page (not the skeleton) is shown, then a step fails.
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('caps the central pile stack at 10 cards visually', async () => {
    mockExec.mockResolvedValueOnce({ ...penaltyState, centralPileSize: 15 });
    renderWithProviders(<BeggarMyNeighbourPage />);
    const stack = await screen.findByTestId('bmn-central-pile-stack');
    expect(stack).toHaveAttribute('data-pile-size', '15');
    expect(stack.querySelectorAll('[data-testid="animated-card-back"]')).toHaveLength(10);
  });

  it('shows last card played when present', async () => {
    mockExec.mockResolvedValueOnce(penaltyState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
  });

  it('shows collect phase prompt', async () => {
    mockExec.mockResolvedValueOnce(collectState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getAllByText(/5/).length).toBeGreaterThan(0);
  });

  it('disables step button on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
  });

  it('disables autoplay button on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByTestId('autoplay-button')).toBeDisabled());
  });

  it('renders CLI terminal when enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.queryByTestId('step-button')).not.toBeInTheDocument());
  });

  it('shows each player held-card total in the summary readout', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      players: [
        { id: 0, isHuman: true, drawPileSize: 30, discardPileSize: 9, totalCards: 39 },
        { id: 1, isHuman: false, drawPileSize: 10, discardPileSize: 3, totalCards: 13 },
      ],
    });
    renderWithProviders(<BeggarMyNeighbourPage />);
    const counts = await screen.findByTestId('bmn-card-counts');
    expect(counts).toHaveTextContent(/あなた.*計 39 枚/);
    expect(counts).toHaveTextContent(/CPU.*計 13 枚/);
  });

  it('reflects the held-card ratio in the count bar', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      players: [
        { id: 0, isHuman: true, drawPileSize: 30, discardPileSize: 9, totalCards: 39 },
        { id: 1, isHuman: false, drawPileSize: 10, discardPileSize: 3, totalCards: 13 },
      ],
    });
    renderWithProviders(<BeggarMyNeighbourPage />);
    // 39 / (39 + 13) = 75%.
    const bar = await screen.findByTestId('bmn-count-bar');
    expect(bar).toHaveAttribute('data-you-pct', '75');
    expect(bar).toHaveAttribute('role', 'img');
    expect(bar).toHaveAccessibleName('持ち札 あなた 39 枚、CPU 13 枚');
  });

  it('shows round progress toward the max-rounds cap', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      roundsPlayed: 250,
      config: { maxRounds: 1000 },
    });
    renderWithProviders(<BeggarMyNeighbourPage />);
    const progress = await screen.findByTestId('bmn-round-progress');
    expect(progress).toHaveTextContent('ラウンド 250 / 1000');
  });

  it('exposes accessible help for the max-rounds setting', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />);
    await waitFor(() => expect(screen.getByText('最大ラウンド数')).toBeInTheDocument());
    fireEvent.click(screen.getByText('(?)'));
    expect(screen.getByRole('tooltip')).toHaveTextContent('ペナルティカードの連鎖でゲームが長引く場合の上限です');
  });
});
