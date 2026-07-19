import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { warApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { WarResponse } from '../types/card';
import { WarPhase } from '../types/phases';
import { WarPage } from './WarPage';

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
  warApi: { exec: vi.fn() },
  actionLogApi: { war: vi.fn() },
}));

const mockExec = vi.mocked(warApi.exec);

const baseState: WarResponse = {
  players: [
    { id: 0, isHuman: true, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
    { id: 1, isHuman: false, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
  ],
  phase: WarPhase.REVEAL,
  gameEndFlag: false,
  winnerIdx: -1,
  playerRevealed: null,
  cpuRevealed: null,
  warPotSize: 0,
  lastWinnerIdx: -1,
  lastBurialCount: 0,
  roundsPlayed: 0,
  config: { maxRounds: 500 },
  message: '',
  messageCode: 'reveal',
};

const warPhaseState: WarResponse = {
  ...baseState,
  phase: WarPhase.WAR_BURY,
  playerRevealed: { design: 'SPADE', value: 7 },
  cpuRevealed: { design: 'HEART', value: 7 },
  warPotSize: 2,
  messageCode: 'war',
};

const gameEndState: WarResponse = {
  ...baseState,
  phase: WarPhase.GAME_END,
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
  localStorage.removeItem('war:autoPlaySpeed');
});

describe('WarPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('exposes accessible help for the max-rounds setting', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByText('最大ラウンド数')).toBeInTheDocument());
    // The (?) toggle reveals an accessible tooltip explaining the end condition.
    // The first help toggle belongs to the max-rounds setting.
    fireEvent.click(screen.getAllByText('(?)')[0]);
    expect(screen.getByRole('tooltip')).toHaveTextContent('無限');
  });

  it('renders pile info after state loads', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    // Both players show draw pile = 26 (rendered as "山札: 26")
    const pileLines = screen.getAllByText(/26/);
    expect(pileLines.length).toBeGreaterThan(0);
  });

  it('step button calls exec with step', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('shows war phase pot count when tie occurred', async () => {
    mockExec.mockResolvedValueOnce(warPhaseState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    // Pot size 2 should appear in the label
    expect(screen.getAllByText(/2/).length).toBeGreaterThan(0);
  });

  it('highlights the winner card and dims the loser when the player wins a round', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      phase: WarPhase.RESOLVED,
      playerRevealed: { design: 'SPADE', value: 13 },
      cpuRevealed: { design: 'HEART', value: 5 },
      lastWinnerIdx: 0,
    });
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
    expect(screen.getByAltText('♠ K')).toHaveClass('ring-ds-success');
    expect(screen.getByAltText('♥ 5')).toHaveClass('opacity-60');
  });

  it('highlights the CPU card when the CPU wins a round', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      phase: WarPhase.RESOLVED,
      playerRevealed: { design: 'SPADE', value: 3 },
      cpuRevealed: { design: 'HEART', value: 9 },
      lastWinnerIdx: 1,
    });
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 9')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 9')).toHaveClass('ring-ds-success');
    expect(screen.getByAltText('♠ 3')).toHaveClass('opacity-60');
  });

  it('marks both cards with a warning ring during a war', async () => {
    mockExec.mockResolvedValueOnce(warPhaseState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 7')).toBeInTheDocument());
    expect(screen.getByAltText('♠ 7')).toHaveClass('ring-ds-warning');
    expect(screen.getByAltText('♥ 7')).toHaveClass('ring-ds-warning');
  });

  it('shows no emphasis during the reveal phase', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      playerRevealed: { design: 'SPADE', value: 4 },
      cpuRevealed: { design: 'HEART', value: 8 },
    });
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 4')).toBeInTheDocument());
    for (const alt of ['♠ 4', '♥ 8']) {
      expect(screen.getByAltText(alt)).not.toHaveClass('ring-ds-success');
      expect(screen.getByAltText(alt)).not.toHaveClass('ring-ds-warning');
      expect(screen.getByAltText(alt)).not.toHaveClass('opacity-60');
    }
  });

  it('disables step button on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
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
      renderWithProviders(<WarPage />);
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
      renderWithProviders(<WarPage />);
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

  it('renders the animation speed selector persisted to localStorage', async () => {
    renderWithProviders(<WarPage />);
    const select = (await screen.findByTestId('autoplay-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('normal');
    fireEvent.change(select, { target: { value: 'slow' } });
    expect(select.value).toBe('slow');
    expect(localStorage.getItem('war:autoPlaySpeed')).toBe('slow');
  });

  it('disables autoplay button on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<WarPage />);
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
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.queryByTestId('step-button')).not.toBeInTheDocument());
  });

  it('does not render the pot stack when warPotSize is 2 or fewer', async () => {
    // warPhaseState carries warPotSize: 2 — the visual stack only appears when burial cards push
    // the pot past the initial tied pair (> 2).
    mockExec.mockResolvedValueOnce(warPhaseState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.queryByTestId('war-pot-stack')).not.toBeInTheDocument();
  });

  it('renders the pot stack with data-pot-size matching warPotSize when > 2', async () => {
    mockExec.mockResolvedValueOnce({ ...warPhaseState, warPotSize: 4 });
    renderWithProviders(<WarPage />);
    const stack = await screen.findByTestId('war-pot-stack');
    expect(stack).toHaveAttribute('data-pot-size', '4');
    // 4 cards → 4 face-down placeholders inside the stack.
    expect(stack.querySelectorAll('[data-testid="animated-card-back"]')).toHaveLength(4);
  });

  it('hides the decorative pot stack from AT while keeping the count text readable', async () => {
    mockExec.mockResolvedValueOnce({ ...warPhaseState, warPotSize: 4 });
    renderWithProviders(<WarPage />);
    const stack = await screen.findByTestId('war-pot-stack');
    expect(stack).toHaveAttribute('aria-hidden', 'true');
    // The count is still conveyed by the adjacent text (not hidden).
    expect(screen.getByText(/4/)).toBeInTheDocument();
  });

  it('caps the visual pot stack at 10 cards even when warPotSize is larger', async () => {
    mockExec.mockResolvedValueOnce({ ...warPhaseState, warPotSize: 15 });
    renderWithProviders(<WarPage />);
    const stack = await screen.findByTestId('war-pot-stack');
    expect(stack).toHaveAttribute('data-pot-size', '15');
    // The render cap keeps the layout stable on long war chains.
    expect(stack.querySelectorAll('[data-testid="animated-card-back"]')).toHaveLength(10);
  });
});
