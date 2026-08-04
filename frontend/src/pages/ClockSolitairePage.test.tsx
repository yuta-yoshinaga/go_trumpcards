import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { clocksolitaireApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ClockSolitaireCard, ClockSolitaireResponse } from '../types/card';
import { ClockSolitairePage } from './ClockSolitairePage';

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
  clocksolitaireApi: { exec: vi.fn() },
  actionLogApi: { clocksolitaire: vi.fn() },
}));

const mockExec = vi.mocked(clocksolitaireApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTestPiles(): ClockSolitaireCard[][] {
  const piles: ClockSolitaireCard[][] = [];
  for (let i = 0; i < 13; i++) {
    const pile: ClockSolitaireCard[] = [];
    for (let j = 0; j < 4; j++) {
      pile.push({ card: card('SPADE', (i % 13) + 1), faceUp: false });
    }
    piles.push(pile);
  }
  piles[12][3].faceUp = true;
  return piles;
}

function makeFaceUpCount(): number[] {
  const fuc = Array(13).fill(0);
  fuc[12] = 1;
  return fuc;
}

const playingState: ClockSolitaireResponse = {
  piles: makeTestPiles(),
  faceUpCount: makeFaceUpCount(),
  phase: 0,
  stepCount: 0,
  currentCard: card('SPADE', 5),
  message: '',
  messageCode: 'clocksolitaire.playing',
};

const gameClearState: ClockSolitaireResponse = {
  ...playingState,
  phase: 1,
  stepCount: 48,
  currentCard: undefined,
  messageCode: 'clocksolitaire.gameClear',
  messageParams: { stepCount: '48' },
};

const gameOverState: ClockSolitaireResponse = {
  ...playingState,
  phase: 2,
  stepCount: 30,
  currentCard: undefined,
  messageCode: 'clocksolitaire.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

afterEach(() => {
  localStorage.removeItem('clocksolitaire:autoPlaySpeed');
});

describe('ClockSolitairePage', () => {
  it('renders heading', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders step count', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ClockSolitairePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders the hint toggle inside the settings panel', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('設定')).toBeInTheDocument());
    expect(screen.getByLabelText('ヒント表示')).toBeInTheDocument();
  });

  it('shows game clear phase name', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームクリア'));
  });

  it('mirrors the resolved message in a polite live region', async () => {
    renderWithProviders(<ClockSolitairePage />);
    const live = await screen.findByTestId('cs-live-region');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveAttribute('role', 'status');
    await waitFor(() => expect(live).toHaveTextContent('プレイ中'));
  });

  it('announces the game-clear result (with step count) in the live region', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<ClockSolitairePage />);
    const live = await screen.findByTestId('cs-live-region');
    await waitFor(() => expect(live).toHaveTextContent('ゲームクリア'));
    expect(live.textContent).toMatch(/48/);
  });

  it('falls back to the raw message when there is no messageCode', async () => {
    mockExec.mockResolvedValue({ ...playingState, messageCode: undefined, message: 'カスタムメッセージ' });
    renderWithProviders(<ClockSolitairePage />);
    const live = await screen.findByTestId('cs-live-region');
    await waitFor(() => expect(live).toHaveTextContent('カスタムメッセージ'));
  });

  it('falls back to the raw message when the messageCode has no translation', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      messageCode: 'clocksolitaire.unknownCode',
      message: '生メッセージ',
    });
    renderWithProviders(<ClockSolitairePage />);
    const live = await screen.findByTestId('cs-live-region');
    await waitFor(() => expect(live).toHaveTextContent('生メッセージ'));
  });

  it('shows game over phase name', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
  });

  it('renders clock position labels', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('12')).toBeInTheDocument());
    expect(screen.getByText('6')).toBeInTheDocument();
  });

  it('renders center K label', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('K')).toBeInTheDocument());
  });

  it('renders current card section when playing', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('手持ちカード')).toBeInTheDocument());
  });

  it('does not render current card section when game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    expect(screen.queryByText('手持ちカード')).not.toBeInTheDocument();
  });

  it('renders face-up count for piles', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getAllByText('0/4').length).toBeGreaterThan(0));
  });

  it('renders playing phase indicator', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('プレイ中'));
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
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
    expect(screen.queryByText('手持ちカード')).not.toBeInTheDocument();
  });

  // Helper: the CLI-mode test above leaves `mockUseCliMode` returning `cliEnabled: true`
  // because vi.clearAllMocks() only clears call history, not the implementation. Restore
  // the default cliEnabled: false at the top of each remaining test so the clock face renders.
  const restoreCliModeDefault = (): void => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: false,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
  };

  it('marks the matching hour pile as the flight target while a 1-12 card is waiting', async () => {
    restoreCliModeDefault();
    // playingState already has currentCard = ♠5, so the 5 o'clock pile (index 4) should be the
    // single flight target. Other piles must not carry the data attribute.
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
    const targets = document.querySelectorAll('[data-flight-target="true"]');
    expect(targets).toHaveLength(1);
    const target = targets[0] as HTMLElement;
    expect(target.className).toContain('ring-ds-warning');
    expect(target.className).toContain('animate-pulse');
  });

  it('marks the center K pile as the flight target when a King is waiting', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...playingState, currentCard: card('SPADE', 13) });
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
    const targets = document.querySelectorAll('[data-flight-target="true"]');
    // Only the center K pile should match for value=13.
    expect(targets).toHaveLength(1);
    expect((targets[0] as HTMLElement).className).toContain('ring-ds-warning');
  });

  it('renders no flight target when there is no card awaiting placement', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...playingState, currentCard: undefined });
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
    expect(document.querySelectorAll('[data-flight-target="true"]')).toHaveLength(0);
  });

  it('autoplay auto-advances via repeated step calls and stops at game clear', async () => {
    restoreCliModeDefault();
    vi.useFakeTimers();
    try {
      // reset (mount) → playing, then two client-driven steps; the second clears the game.
      mockExec
        .mockReset()
        .mockResolvedValueOnce(playingState) // reset on mount
        .mockResolvedValueOnce(playingState) // step 1
        .mockResolvedValueOnce(gameClearState) // step 2 → game clear
        .mockResolvedValue(gameClearState);
      renderWithProviders(<ClockSolitairePage />);
      // Flush the mount reset.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      fireEvent.click(screen.getByTestId('cs-autoplay-button'));
      // Two normal-speed ticks (450ms each) place two cards, then the game clears.
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
      // Client-driven autoplay never hits the single-shot server command.
      expect(mockExec).not.toHaveBeenCalledWith('autoplay');
    } finally {
      vi.useRealTimers();
    }
  });

  it('faster speed shortens the auto-advance interval', async () => {
    restoreCliModeDefault();
    vi.useFakeTimers();
    try {
      mockExec.mockReset().mockResolvedValue(playingState);
      renderWithProviders(<ClockSolitairePage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      // Switch to fast (150ms delay) before starting autoplay.
      fireEvent.change(screen.getByTestId('autoplay-speed-select'), { target: { value: 'fast' } });
      fireEvent.click(screen.getByTestId('cs-autoplay-button'));
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
    restoreCliModeDefault();
    renderWithProviders(<ClockSolitairePage />);
    const select = (await screen.findByTestId('autoplay-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('normal');
    fireEvent.change(select, { target: { value: 'slow' } });
    expect(select.value).toBe('slow');
    expect(localStorage.getItem('clocksolitaire:autoPlaySpeed')).toBe('slow');
  });

  it('toggles the autoplay button label and disables step while autoplaying', async () => {
    restoreCliModeDefault();
    vi.useFakeTimers();
    try {
      mockExec.mockReset().mockResolvedValue(playingState);
      renderWithProviders(<ClockSolitairePage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      const autoBtn = screen.getByTestId('cs-autoplay-button');
      expect(autoBtn).toHaveTextContent('オートプレイ');
      expect(screen.getByTestId('cs-step-button')).not.toBeDisabled();
      fireEvent.click(autoBtn);
      expect(autoBtn).toHaveTextContent('停止');
      expect(autoBtn).toHaveAttribute('aria-pressed', 'true');
      expect(screen.getByTestId('cs-step-button')).toBeDisabled();
    } finally {
      vi.useRealTimers();
    }
  });

  it('disables the undo button when canUndo is false', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<ClockSolitairePage />);
    const undoBtn = await screen.findByTestId('cs-undo-button');
    expect(undoBtn).toBeDisabled();
  });

  it('enables the undo button and dispatches undo on click', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<ClockSolitairePage />);
    const undoBtn = await screen.findByTestId('cs-undo-button');
    await waitFor(() => expect(undoBtn).not.toBeDisabled());
    fireEvent.click(undoBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('shows an enabled undo button after game over so the last move can be reverted', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...gameOverState, canUndo: true });
    renderWithProviders(<ClockSolitairePage />);
    const undoBtn = await screen.findByTestId('cs-undo-button');
    await waitFor(() => expect(undoBtn).not.toBeDisabled());
    // The step/autoplay controls are hidden once the game has ended.
    expect(screen.queryByTestId('cs-step-button')).not.toBeInTheDocument();
  });

  it('announces where the drawn card is heading', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue({ ...playingState, currentCard: { design: 'SPADE', value: 3 } });
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('cs-live-region')).toHaveTextContent(/3時/));
  });

  it('announces the centre pile for a king', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue({ ...playingState, currentCard: { design: 'SPADE', value: 13 } });
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('cs-live-region')).toHaveTextContent(/中央/));
  });
});
