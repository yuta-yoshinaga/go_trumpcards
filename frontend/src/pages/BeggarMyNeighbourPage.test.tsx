import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
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

afterEach(() => {
  localStorage.removeItem('beggarmyneighbour:autoPlaySpeed');
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

  it('autoplay auto-advances via repeated step calls and stops at game end', async () => {
    vi.useFakeTimers();
    try {
      // reset (mount) → base, then two client-driven steps; the second ends the game.
      mockExec
        .mockReset()
        .mockResolvedValueOnce(baseState) // reset on mount
        .mockResolvedValueOnce(baseState) // step 1
        .mockResolvedValueOnce(gameEndState) // step 2 → game end
        .mockResolvedValue(gameEndState);
      renderWithProviders(<BeggarMyNeighbourPage />);
      // Flush the mount reset.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      fireEvent.click(screen.getByTestId('autoplay-button'));
      // Two normal-speed ticks (450ms each) advance two rounds, then the game ends.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(450);
      });
      await act(async () => {
        await vi.advanceTimersByTimeAsync(450);
      });
      // Any further ticks must not fire more steps once the game has ended.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2000);
      });
      const stepCalls = mockExec.mock.calls.filter(([cmd]) => cmd === 'step');
      expect(stepCalls).toHaveLength(2);
      // autoplay never hits the single-shot server command.
      expect(mockExec).not.toHaveBeenCalledWith('autoplay');
    } finally {
      vi.useRealTimers();
    }
  });

  it('faster speed shortens the auto-advance interval', async () => {
    vi.useFakeTimers();
    try {
      mockExec.mockReset().mockResolvedValue(baseState);
      renderWithProviders(<BeggarMyNeighbourPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      // Switch to fast (150ms delay) before starting autoplay.
      fireEvent.change(screen.getByTestId('autoplay-speed-select'), { target: { value: 'fast' } });
      fireEvent.click(screen.getByTestId('autoplay-button'));
      // 150ms is enough at fast speed but would be too short at the 450ms normal delay.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(150);
      });
      expect(mockExec).toHaveBeenCalledWith('step');
    } finally {
      vi.useRealTimers();
    }
  });

  it('renders the playback speed selector persisted to localStorage', async () => {
    renderWithProviders(<BeggarMyNeighbourPage />);
    const select = (await screen.findByTestId('autoplay-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('normal');
    fireEvent.change(select, { target: { value: 'slow' } });
    expect(select.value).toBe('slow');
    expect(localStorage.getItem('beggarmyneighbour:autoPlaySpeed')).toBe('slow');
  });

  it('stops autoplay when the stop button is toggled', async () => {
    vi.useFakeTimers();
    try {
      mockExec.mockReset().mockResolvedValue(baseState);
      renderWithProviders(<BeggarMyNeighbourPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      // Start autoplay: button flips to the stop label.
      fireEvent.click(screen.getByTestId('autoplay-button'));
      expect(screen.getByTestId('autoplay-button')).toHaveAttribute('aria-pressed', 'true');
      await act(async () => {
        await vi.advanceTimersByTimeAsync(450);
      });
      const afterStart = mockExec.mock.calls.filter(([cmd]) => cmd === 'step').length;
      expect(afterStart).toBeGreaterThan(0);
      // Toggle off; no further steps should fire.
      fireEvent.click(screen.getByTestId('autoplay-button'));
      expect(screen.getByTestId('autoplay-button')).toHaveAttribute('aria-pressed', 'false');
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2000);
      });
      const afterStop = mockExec.mock.calls.filter(([cmd]) => cmd === 'step').length;
      expect(afterStop).toBe(afterStart);
    } finally {
      vi.useRealTimers();
    }
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
    // The first help toggle belongs to the max-rounds setting (the speed setting adds a second).
    fireEvent.click(screen.getAllByText('(?)')[0]);
    expect(screen.getByRole('tooltip')).toHaveTextContent('ペナルティカードの連鎖でゲームが長引く場合の上限です');
  });
});
