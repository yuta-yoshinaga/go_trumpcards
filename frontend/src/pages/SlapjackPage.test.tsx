import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { slapjackApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SlapjackResponse } from '../types/card';
import { SlapjackEventKind, SlapjackPendingKind, SlapjackPhase } from '../types/phases';
import { SlapjackPage } from './SlapjackPage';

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
  slapjackApi: { exec: vi.fn() },
  actionLogApi: { slapjack: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  // AnimatedCard and the central taps (useGameApi / GamePageShell) consume
  // useOptionalSound; route it to the same spy and assert on specific sound
  // names so per-card deal sounds don't interfere.
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(slapjackApi.exec);

const baseState: SlapjackResponse = {
  phase: SlapjackPhase.PLAY,
  gameEndFlag: false,
  winnerIdx: -1,
  currentTurnIdx: 0,
  isHumanTurn: true,
  isTopJack: false,
  centerPileSize: 0,
  topCard: null,
  players: [
    { name: 'You', isHuman: true, stockSize: 26 },
    { name: 'CPU', isHuman: false, stockSize: 26 },
  ],
  cpuDifficulty: 1,
  pendingKind: 0,
  pendingDeadlineMs: 0,
  lastEventKind: 0,
  lastEventPlayerIdx: 0,
  message: '',
};

const jackOnTopState: SlapjackResponse = {
  ...baseState,
  isTopJack: true,
  centerPileSize: 1,
  topCard: { design: 'SPADE', value: 11 },
};

const gameEndState: SlapjackResponse = {
  ...baseState,
  phase: SlapjackPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { name: 'You', isHuman: true, stockSize: 52 },
    { name: 'CPU', isHuman: false, stockSize: 0 },
  ],
};

beforeEach(() => {
  mockPlaySound.mockClear();
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

describe('SlapjackPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the GameSkeleton while state is null', () => {
    // Keep exec pending so `state` stays null and the loading guard renders.
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<SlapjackPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByTestId('step-button')).not.toBeInTheDocument();
  });

  it('renders an ErrorAlert with a retry button when an action fails', async () => {
    // Mount reset resolves so state loads; a subsequent action rejects.
    mockExec.mockResolvedValueOnce(baseState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    mockExec.mockRejectedValue(new Error('boom'));
    fireEvent.click(screen.getByTestId('step-button'));
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('通信エラー');
    expect(screen.getByRole('button', { name: /再試行/i })).toBeInTheDocument();
  });

  it('announces pending CPU flips without a countdown and clears the visible status when complete', async () => {
    const pendingState = {
      ...baseState,
      pendingKind: SlapjackPendingKind.STEP,
      pendingDeadlineMs: Date.now() + 1500,
    };
    mockExec.mockResolvedValueOnce(pendingState);
    renderWithProviders(<SlapjackPage />);

    const liveRegion = await screen.findByTestId('cpu-pending-announcement');
    await waitFor(() => expect(liveRegion).toHaveTextContent('CPUのめくりを待っています'));
    expect(liveRegion.textContent).not.toMatch(/\d/);
    const status = screen.getByTestId('cpu-pending-status');
    expect(status).toHaveTextContent(/CPUのめくり実行待ち（残り約\d+秒）/);
    expect(screen.getByRole('progressbar', { name: 'CPUの予約アクション実行までの進捗' })).toBeInTheDocument();

    mockExec.mockResolvedValue(baseState);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('tick'));
    await waitFor(() => expect(liveRegion).toBeEmptyDOMElement());
    expect(screen.queryByTestId('cpu-pending-status')).not.toBeInTheDocument();
    expect(screen.queryByRole('progressbar', { name: 'CPUの予約アクション実行までの進捗' })).not.toBeInTheDocument();
  });

  it('announces pending CPU slaps with the slap-specific message', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      pendingKind: SlapjackPendingKind.SLAP,
      pendingDeadlineMs: Date.now() + 1500,
    });
    renderWithProviders(<SlapjackPage />);

    const liveRegion = await screen.findByTestId('cpu-pending-announcement');
    await waitFor(() => expect(liveRegion).toHaveTextContent('CPUのスラップを待っています'));
    expect(liveRegion.textContent).not.toMatch(/\d/);
    expect(screen.getByTestId('cpu-pending-status')).toHaveTextContent(/CPUのスラップ実行待ち（残り約\d+秒）/);
    expect(screen.getByRole('progressbar', { name: 'CPUの予約アクション実行までの進捗' })).toBeInTheDocument();
  });

  it('starts a new pending countdown below 100 percent, including after switching deadlines', async () => {
    vi.useFakeTimers();
    try {
      vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
      const firstPending = {
        ...baseState,
        pendingKind: SlapjackPendingKind.STEP,
        pendingDeadlineMs: Date.now() + 5000,
      };
      const nextPending = { ...firstPending, pendingDeadlineMs: Date.now() + 8000 };
      mockExec.mockResolvedValueOnce(firstPending).mockResolvedValue(nextPending);

      renderWithProviders(<SlapjackPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      const progress = screen.getByRole('progressbar', { name: 'CPUの予約アクション実行までの進捗' });
      expect(Number(progress.getAttribute('aria-valuenow'))).toBeLessThan(100);

      await act(async () => {
        await vi.advanceTimersByTimeAsync(1000);
      });
      await vi.waitFor(() => expect(mockExec).toHaveBeenCalledWith('tick'));
      expect(Number(progress.getAttribute('aria-valuenow'))).toBeLessThan(100);
    } finally {
      vi.useRealTimers();
    }
  });

  it('keeps the pending countdown at about one second after its deadline', async () => {
    vi.useFakeTimers();
    try {
      vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
      mockExec.mockResolvedValue({
        ...baseState,
        pendingKind: SlapjackPendingKind.STEP,
        pendingDeadlineMs: Date.now() + 500,
      });
      renderWithProviders(<SlapjackPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      const status = screen.getByTestId('cpu-pending-status');
      await act(async () => {
        await vi.advanceTimersByTimeAsync(1000);
      });
      expect(status).toHaveTextContent(/残り約1秒/);
    } finally {
      vi.useRealTimers();
    }
  });

  it('renders stock counts after state loads', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getAllByText(/26/).length).toBeGreaterThan(0);
  });

  it('announces newly revealed non-J cards and leaves Jack announcements to the slap notice', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    const liveRegion = screen.getByTestId('sj-card-announce');
    expect(liveRegion).toBeEmptyDOMElement();

    mockExec.mockResolvedValueOnce({
      ...baseState,
      centerPileSize: 1,
      topCard: { design: 'HEART', value: 5 },
    });
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(liveRegion).toHaveTextContent('カードが出ました。♥ 5'));

    mockExec.mockResolvedValueOnce(jackOnTopState);
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(screen.getByTestId('sj-jack-announce')).toHaveTextContent('スラップ可能です'));
    expect(liveRegion).toBeEmptyDOMElement();
  });

  it('shows accessible stock progress for both players and clamps an empty stock to zero', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<SlapjackPage />);
    const cpuProgress = await screen.findByRole('progressbar', { name: 'CPU 1のストック残り 0枚' });
    const humanProgress = screen.getByRole('progressbar', { name: 'あなたのストック残り 52枚' });
    expect(cpuProgress).toHaveAttribute('aria-valuemin', '0');
    expect(cpuProgress).toHaveAttribute('aria-valuemax', '52');
    expect(cpuProgress).toHaveAttribute('aria-valuenow', '0');
    expect(cpuProgress.firstElementChild).toHaveStyle({ width: '0%' });
    expect(humanProgress).toHaveAttribute('aria-valuenow', '52');
    expect(humanProgress.firstElementChild).toHaveStyle({ width: '100%' });
  });

  it('step button calls exec with step', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('slap button calls exec with slap', async () => {
    mockExec.mockResolvedValueOnce(jackOnTopState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('slap-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('slap'));
  });

  it('disables slap button when pile is empty', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeDisabled());
  });

  it('disables step and slap buttons on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
    expect(screen.getByTestId('slap-button')).toBeDisabled();
  });

  it('disables step button when it is not the human turn', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, isHumanTurn: false, currentTurnIdx: 1 });
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
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
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.queryByTestId('step-button')).not.toBeInTheDocument());
  });

  it('renders the Jack-on-top callout and a flashing slap button', async () => {
    mockExec.mockResolvedValueOnce(jackOnTopState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    const slap = screen.getByTestId('slap-button');
    expect(slap).not.toBeDisabled();
    expect(slap.className).toMatch(/animate-pulse/);
    // Buttons meet the full WCAG 2.5.5 44x44px tap-target minimum.
    expect(slap.className).toContain('min-h-[44px]');
    expect(slap.className).toContain('min-w-[44px]');
    const step = screen.getByTestId('step-button');
    expect(step.className).toContain('min-h-[44px]');
    expect(step.className).toContain('min-w-[44px]');
  });

  it('announces the slap chance via an atomic assertive live region when a Jack is on top', async () => {
    mockExec.mockResolvedValueOnce(jackOnTopState);
    renderWithProviders(<SlapjackPage />);
    const announce = await screen.findByTestId('sj-jack-announce');
    expect(announce).toHaveAttribute('aria-live', 'assertive');
    expect(announce).toHaveAttribute('aria-atomic', 'true');
    expect(announce).toHaveTextContent('スラップ可能です');
  });

  it('announces only slap-availability changes and clears the notice when the Jack leaves', async () => {
    mockExec.mockResolvedValueOnce(jackOnTopState).mockResolvedValueOnce(baseState);
    renderWithProviders(<SlapjackPage />);
    const announce = await screen.findByTestId('sj-jack-announce');
    expect(announce).toHaveTextContent('スラップ可能');
    fireEvent.click(screen.getByTestId('slap-button'));
    await waitFor(() => expect(announce).toHaveTextContent('スラップ不可'));
  });

  it('keeps the live region empty when no Jack is on top', async () => {
    mockExec.mockResolvedValueOnce(baseState);
    renderWithProviders(<SlapjackPage />);
    const announce = await screen.findByTestId('sj-jack-announce');
    expect(announce).toBeEmptyDOMElement();
  });

  it('announces a correct slap by the human via a polite status live region', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 0,
    });
    renderWithProviders(<SlapjackPage />);
    const announce = await screen.findByTestId('sj-slap-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    await waitFor(() => expect(announce).toHaveTextContent('スラップ成功（あなた）'));
  });

  it('shows the cards won in the correct slap burst and announcement', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 0,
      lastEventCardsWon: 4,
    });
    renderWithProviders(<SlapjackPage />);
    const burst = await screen.findByTestId('slap-burst');
    expect(burst).toHaveTextContent('あなた：ジャック！4枚獲得');
    expect(await screen.findByTestId('sj-slap-announce')).toHaveTextContent('スラップ成功（あなた、4枚獲得）');
  });

  it('uses the legacy correct-slap label when no cards-won count is available', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 0,
      lastEventCardsWon: 0,
    });
    renderWithProviders(<SlapjackPage />);
    expect(await screen.findByTestId('slap-burst')).toHaveTextContent('スラップ成功（あなた）');
  });

  it('does not show a cards-won count for a wrong slap', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_WRONG,
      lastEventPlayerIdx: 0,
      lastEventCardsWon: 0,
    });
    renderWithProviders(<SlapjackPage />);
    const burst = await screen.findByTestId('slap-burst');
    expect(burst).toHaveTextContent('あなた：ミス！');
    expect(burst).not.toHaveTextContent('枚獲得');
    expect(await screen.findByTestId('sj-slap-announce')).toHaveTextContent('お手つき（あなた）');
  });

  it('announces a false slap by the CPU naming the offender', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_WRONG,
      lastEventPlayerIdx: 1,
    });
    renderWithProviders(<SlapjackPage />);
    const burst = await screen.findByTestId('slap-burst');
    expect(burst).toHaveTextContent('CPU 1：ミス！');
    const announce = await screen.findByTestId('sj-slap-announce');
    await waitFor(() => expect(announce).toHaveTextContent('お手つき（CPU 1）'));
  });

  it('renders the game-end state with both buttons disabled when human wins', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
    expect(screen.getByTestId('slap-button')).toBeDisabled();
  });

  it('reset settings select fires reset with cpuDifficulty config', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    const select = screen.getByLabelText('CPU難易度') as HTMLSelectElement;
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('renders Enter and Space keyboard badges on step/slap buttons', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getByTestId('step-button').textContent).toContain('Enter');
    expect(screen.getByTestId('slap-button').textContent).toContain('Space');
  });

  it('Space keypress triggers slap during play', async () => {
    mockExec.mockResolvedValueOnce(jackOnTopState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: ' ' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('slap'));
  });

  it('Enter keypress triggers step during play', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('plays a card sound when the human flips via the step button', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    mockPlaySound.mockClear();
    fireEvent.click(screen.getByTestId('step-button'));
    // The central tap plays after the exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('plays a fanfare on the human’s correct slap', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 0,
    });
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('winFanfare'));
  });

  it('plays an error buzz on the human’s false slap', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_WRONG,
      lastEventPlayerIdx: 0,
    });
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz'));
  });

  it('stays silent for a CPU slap event (only the human’s slaps make sound)', async () => {
    mockExec.mockResolvedValueOnce({
      ...baseState,
      lastEventKind: SlapjackEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 1,
    });
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    expect(mockPlaySound).not.toHaveBeenCalledWith('winFanfare');
    expect(mockPlaySound).not.toHaveBeenCalledWith('errorBuzz');
  });

  // **ドメインの Tick() は pending.Kind が None なら即座に何もせず返る。**それでも
  // 100ms ごとに投げ続けており、人間が考えている間ずっと毎秒10回の無駄なリクエストが
  // 飛んでいた (#4748)。
  it('does not poll tick while no CPU action is pending', async () => {
    vi.useFakeTimers();
    try {
      mockExec.mockReset().mockResolvedValue(baseState); // pendingKind: NONE
      renderWithProviders(<SlapjackPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      mockExec.mockClear();

      await act(async () => {
        await vi.advanceTimersByTimeAsync(1000); // 従来なら 10 回飛ぶ
      });
      expect(mockExec).not.toHaveBeenCalledWith('tick');
    } finally {
      vi.useRealTimers();
    }
  });

  // **逆側。**予約があるうちは従来どおり回さないとゲームが進まない。
  // CPU のスラップは人間の手番中にも予約されるので、手番ではなく予約でゲートする。
  it('keeps polling tick while a CPU slap is pending', async () => {
    vi.useFakeTimers();
    try {
      const pendingSlapState: SlapjackResponse = { ...baseState, pendingKind: SlapjackPendingKind.SLAP };
      mockExec.mockReset().mockResolvedValue(pendingSlapState);
      renderWithProviders(<SlapjackPage />);
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      mockExec.mockClear();

      await act(async () => {
        await vi.advanceTimersByTimeAsync(300);
      });
      expect(mockExec).toHaveBeenCalledWith('tick');
    } finally {
      vi.useRealTimers();
    }
  });
});

describe('SlapjackPage keyboard shortcut announcement', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  // KbdBadge は目で見える案内でしかない。aria-keyshortcuts が無いと、
  // 支援技術はショートカットが存在すること自体を知れない (#6375)。
  it('announces the same keys useReflexShortcuts listens for', async () => {
    mockExec.mockResolvedValue(baseState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());

    // 値は useReflexShortcuts.ts の実装と一致していること:
    //   Enter / s / S -> step, Space -> slap
    expect(screen.getByTestId('step-button')).toHaveAttribute('aria-keyshortcuts', 'Enter S');
    expect(screen.getByTestId('slap-button')).toHaveAttribute('aria-keyshortcuts', 'Space');
  });
});
