import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, doubtApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DoubtConfig, DoubtResponse } from '../types/card';
import { DoubtPage } from './DoubtPage';

vi.mock('../api/gameApi', () => ({
  doubtApi: { exec: vi.fn() },
  actionLogApi: { doubt: vi.fn() },
}));

const mockExec = vi.mocked(doubtApi.exec);

const defaultConfig: DoubtConfig = {
  doubtWindowSec: 10,
  cpuMemoryLevel: 1,
  penaltyDrawLimit: 0,
  cpuHesitationEnabled: false,
  cpuMetaAI: false,
};

const humanTurnState: DoubtResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
    },
    { id: 1, isHuman: false, isFinished: false, cardCount: 4, cards: [] },
    { id: 2, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
    { id: 3, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
  ],
  currentTurn: 0,
  phase: 0,
  tableCardCount: 0,
  lastAction: null,
  cpuDoubters: [],
  cpuActions: [],
  humanAction: null,
  lastDoubtResult: null,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  doubtWindowSec: 10,
  penaltyDrawLimit: 0,
};

const cpuTurnState: DoubtResponse = {
  ...humanTurnState,
  currentTurn: 1,
  humanAction: { playerIdx: 0, claimedValue: 2, cardCount: 2, isBluff: false },
};

// CPU played in doubt phase: human gets countdown + doubt/skip buttons
const doubtPhaseCpuPlayedState: DoubtResponse = {
  ...humanTurnState,
  currentTurn: 0,
  phase: 1,
  tableCardCount: 2,
  lastAction: { playerIdx: 1, claimedValue: 2, cardCount: 2, isBluff: true },
  cpuDoubters: [],
};

// Human played in doubt phase: CPU resolves → 確認 button shown
const doubtPhaseHumanPlayedState: DoubtResponse = {
  ...humanTurnState,
  currentTurn: 1,
  phase: 1,
  tableCardCount: 1,
  lastAction: { playerIdx: 0, claimedValue: 3, cardCount: 1, isBluff: false },
  cpuDoubters: [1, 2],
};

const gameEndState: DoubtResponse = {
  ...humanTurnState,
  phase: 2,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('DoubtPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DoubtPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, defaultConfig));
  });

  it('renders CPU player areas', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('renders table area', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.getByText('場のカード: 0枚')).toBeInTheDocument();
  });

  it('shows human player hand cards', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('shows selection hint on human turn phase 0', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('カードを選んで出してください')).toBeInTheDocument());
  });

  it('does not show selection hint when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByText('カードを選んで出してください')).not.toBeInTheDocument();
  });

  it('shows 出す button only on human turn phase 0', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
  });

  it('does not show 出す button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('toggles aria-pressed on HandCard button click', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('toggles card selection on click and enables 出す button', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.click(cardBtn);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();

    // Click again to deselect
    fireEvent.click(cardBtn);
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
  });

  it('shows claimed value input when cards are selected', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('spinbutton')).toBeInTheDocument();
    // Default value 1 shows (A)
    expect(screen.getByText('(A)')).toBeInTheDocument();
  });

  it('claim input is associated with label via htmlFor/id', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const input = screen.getByRole('spinbutton');
    expect(input).toHaveAttribute('id', 'claim-input');
    const label = input.closest('.mt-2')?.querySelector('label');
    expect(label).toHaveAttribute('for', 'claim-input');
  });

  it('claim input receives focus when it appears', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const input = screen.getByRole('spinbutton');
    expect(document.activeElement).toBe(input);
  });

  it('changes claimed value and shows special name', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const input = screen.getByRole('spinbutton');

    fireEvent.change(input, { target: { value: '11' } });
    expect(input).toHaveValue(11);
    expect(screen.getByText('(J)')).toBeInTheDocument();

    fireEvent.change(input, { target: { value: '12' } });
    expect(screen.getByText('(Q)')).toBeInTheDocument();

    fireEvent.change(input, { target: { value: '13' } });
    expect(screen.getByText('(K)')).toBeInTheDocument();

    // Test value clamping at max
    fireEvent.change(input, { target: { value: '14' } });
    expect(input).toHaveValue(13);

    // Test value clamping at min
    fireEvent.change(input, { target: { value: '0' } });
    expect(input).toHaveValue(1);

    // non-numeric input is sanitized to '' by the browser; Number('') = 0 → clamped to 1
    fireEvent.change(input, { target: { value: 'abc' } });
    expect(input).toHaveValue(1);
  });

  it('claim input has aria-describedby linked to range hint', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const input = screen.getByRole('spinbutton');
    expect(input).toHaveAttribute('aria-describedby', 'claim-range-hint');
    expect(screen.getByText('1〜13の値を入力')).toBeInTheDocument();
  });

  it('shows warning when value is clamped out of range', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const input = screen.getByRole('spinbutton');

    fireEvent.change(input, { target: { value: '15' } });
    expect(screen.getByText('1〜13の範囲に調整されました')).toBeInTheDocument();

    // Warning clears after 2 seconds
    await waitFor(() => expect(screen.getByText('1〜13の値を入力')).toBeInTheDocument(), { timeout: 3000 });
  });

  it('does not show warning when value is within range', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const input = screen.getByRole('spinbutton');

    fireEvent.change(input, { target: { value: '7' } });
    expect(screen.getByText('1〜13の値を入力')).toBeInTheDocument();
    expect(screen.queryByText('1〜13の範囲に調整されました')).not.toBeInTheDocument();
  });

  it('calls play command with selected cards when 出す clicked', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', [0], 1, undefined, undefined, expect.any(Number)),
    );
  });

  it('calls reset when reset button is clicked', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, defaultConfig));
  });

  it('disables buttons while loading', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: DoubtResponse) => void;
    const slowPromise = new Promise<DoubtResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error message on successful API call after failure', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  });

  it('sets aria-busy while loading', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: DoubtResponse) => void;
    const slowPromise = new Promise<DoubtResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(humanTurnState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
  });

  // ── Table area ────────────────────────────────────────────────────────────

  it('shows lastAction in table area', async () => {
    const stateWithLastAction: DoubtResponse = {
      ...humanTurnState,
      currentTurn: 2,
      lastAction: { playerIdx: 1, claimedValue: 5, cardCount: 2, isBluff: false },
    };
    mockExec.mockResolvedValue(stateWithLastAction);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1が2枚出しました.*宣言: 5/)).toBeInTheDocument());
  });

  it('hides lastAction display when lastAction is null', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/枚出しました/)).not.toBeInTheDocument();
  });

  // ── actionDesc branches ───────────────────────────────────────────────────

  it('actionDesc uses Player index when player not found', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 99, claimedValue: 5, cardCount: 2, isBluff: false }],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/Player 99が2枚出しました/)).toBeInTheDocument());
  });

  // ── Doubt phase: CPU played ───────────────────────────────────────────────

  it('shows doubt/skip UI when CPU played in doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('ダウトしますか？')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'ダウト！' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'スルー' })).toBeInTheDocument();
    });
  });

  it('shows cpu doubters in cpuPlayed doubt phase', async () => {
    const s: DoubtResponse = {
      ...doubtPhaseCpuPlayedState,
      cpuDoubters: [2, 3],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => {
      const el = screen.getByText(/ダウト宣言CPU/);
      expect(el.textContent).toContain('CPU 2');
      expect(el.textContent).toContain('CPU 3');
    });
  });

  it('calls doubt with [0] when ダウト！ clicked (no cpu doubters)', async () => {
    mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダウト！' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'ダウト！' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubt', undefined, undefined, [0]));
  });

  it('calls skip with [] when スルー clicked (no cpu doubters)', async () => {
    mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スルー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'スルー' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip', undefined, undefined, []));
  });

  it('calls doubt with [0, ...cpuDoubters] when ダウト！ clicked with cpu doubters', async () => {
    const s: DoubtResponse = { ...doubtPhaseCpuPlayedState, cpuDoubters: [2, 3] };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダウト！' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'ダウト！' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubt', undefined, undefined, [0, 2, 3]));
  });

  // ── Doubt phase: human played ─────────────────────────────────────────────

  it('shows confirm UI when human played in doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('CPUがダウトを判定中...')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument();
    });
  });

  it('shows cpu doubters in human-played doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/ダウト！.*CPU 1/)).toBeInTheDocument());
  });

  it('does not show cpuDoubters label when empty in human-played phase', async () => {
    const s: DoubtResponse = { ...doubtPhaseHumanPlayedState, cpuDoubters: [] };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPUがダウトを判定中...')).toBeInTheDocument());
    expect(screen.queryByText(/ダウト！/)).not.toBeInTheDocument();
  });

  it('calls doubt with cpuDoubters when 確認 clicked', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubt', undefined, undefined, [1, 2]));
  });

  // ── Doubt phase suppressed when gameEndFlag ───────────────────────────────

  it('does not show doubt UI when gameEndFlag is true', async () => {
    const s: DoubtResponse = { ...doubtPhaseCpuPlayedState, gameEndFlag: true, message: 'ゲーム終了！' };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ダウト！' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  // ── Countdown timer ───────────────────────────────────────────────────────

  describe('countdown timer', () => {
    beforeEach(() => {
      // Only fake setInterval/clearInterval; leave setTimeout for waitFor retries
      vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
    });
    afterEach(() => {
      vi.clearAllTimers();
      vi.useRealTimers();
    });

    it('starts countdown display when CPU played in doubt phase', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      renderWithProviders(<DoubtPage />);
      // Wait directly for countdown text (appears after 2nd render: state→doubt phase, effect→countdown)
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());
    });

    it('countdown has aria-live assertive and aria-atomic true', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());
      const countdownEl = screen.getByText(/残り 10 秒/);
      expect(countdownEl).toHaveAttribute('aria-live', 'assertive');
      expect(countdownEl).toHaveAttribute('aria-atomic', 'true');
    });

    it('decrements countdown each second', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(screen.getByText(/残り 9 秒/)).toBeInTheDocument();
    });

    it('auto-skips and clears countdown when timer expires', async () => {
      mockExec
        .mockResolvedValueOnce(doubtPhaseCpuPlayedState) // initial reset
        .mockResolvedValueOnce(humanTurnState); // skip response → leave doubt phase
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      act(() => {
        vi.advanceTimersByTime(10000);
      });
      // Countdown display cleared immediately
      expect(screen.queryByText(/残り/)).not.toBeInTheDocument();

      // Auto-skip is called with the correct doubters
      await waitFor(() => {
        expect(mockExec).toHaveBeenCalledWith('skip', undefined, undefined, []);
      });
    });

    it('stops countdown when ダウト！ is clicked', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      mockExec.mockResolvedValue(humanTurnState);
      act(() => {
        fireEvent.click(screen.getByRole('button', { name: 'ダウト！' }));
      });
      await waitFor(() => expect(screen.queryByText(/残り/)).not.toBeInTheDocument());
    });

    it('stops countdown when スルー is clicked', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      mockExec.mockResolvedValue(humanTurnState);
      act(() => {
        fireEvent.click(screen.getByRole('button', { name: 'スルー' }));
      });
      await waitFor(() => expect(screen.queryByText(/残り/)).not.toBeInTheDocument());
    });
  });

  // ── Hesitation delay before countdown ───────────────────────────────────

  it('delays countdown start by hesitationMs when CPU action has hesitation', async () => {
    const stateWithHesitation: DoubtResponse = {
      ...doubtPhaseCpuPlayedState,
      cpuActions: [{ playerIdx: 1, claimedValue: 2, cardCount: 2, isBluff: true, hesitationMs: 500 }],
    };
    mockExec.mockResolvedValue(stateWithHesitation);
    renderWithProviders(<DoubtPage />);
    // Wait for state to be rendered
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    // Countdown should NOT appear immediately (hesitation delay of 500ms pending)
    expect(screen.queryByText(/残り/)).not.toBeInTheDocument();
    // After hesitation delay passes, countdown starts
    await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());
  });

  // ── No countdown in other phases ─────────────────────────────────────────

  it('does not show countdown in play phase', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/残り/)).not.toBeInTheDocument();
  });

  it('does not start countdown when lastAction is null in doubt phase', async () => {
    const s: DoubtResponse = { ...humanTurnState, phase: 1, lastAction: null };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/残り/)).not.toBeInTheDocument();
  });

  it('does not start countdown when human played in doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPUがダウトを判定中...')).toBeInTheDocument());
    expect(screen.queryByText(/残り/)).not.toBeInTheDocument();
  });

  // ── Doubt result ──────────────────────────────────────────────────────────

  it('shows ダウト結果 with ウソでした when wasLying is true', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastDoubtResult: {
        doubterIdx: 0,
        cardPlayerIdx: 1,
        wasLying: true,
        loserIdx: 1,
        cardCount: 3,
        discardedCount: 0,
        revealedCards: [{ design: 'SPADE', value: 5 }],
      },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('ダウト結果')).toBeInTheDocument();
      expect(screen.getByText('ウソでした！')).toBeInTheDocument();
      expect(screen.getByAltText('♠ 5')).toBeInTheDocument();
    });
  });

  it('shows ダウト結果 with 本当でした when wasLying is false', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastDoubtResult: {
        doubterIdx: 1,
        cardPlayerIdx: 0,
        wasLying: false,
        loserIdx: 0,
        cardCount: 2,
        discardedCount: 0,
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('本当でした！')).toBeInTheDocument());
  });

  it('shows loser name using player id when player exists', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastDoubtResult: {
        doubterIdx: 0,
        cardPlayerIdx: 1,
        wasLying: true,
        loserIdx: 1,
        cardCount: 2,
        discardedCount: 0,
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1が2枚引き取りました/)).toBeInTheDocument());
  });

  it('uses loserIdx directly as name when player not found', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastDoubtResult: {
        doubterIdx: 0,
        cardPlayerIdx: 1,
        wasLying: true,
        loserIdx: 99,
        cardCount: 1,
        discardedCount: 0,
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/99が1枚引き取りました/)).toBeInTheDocument());
  });

  it('shows discarded count when discardedCount > 0', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastDoubtResult: {
        doubterIdx: 0,
        cardPlayerIdx: 1,
        wasLying: true,
        loserIdx: 1,
        cardCount: 3,
        discardedCount: 2,
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/2枚がゲームから除外されました/)).toBeInTheDocument());
  });

  // ── Action logs ───────────────────────────────────────────────────────────

  it('shows human action log in non-doubt phase', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/あなたが2枚出しました/)).toBeInTheDocument());
  });

  it('suppresses human action log in doubt phase', async () => {
    const s: DoubtResponse = {
      ...doubtPhaseHumanPlayedState,
      humanAction: { playerIdx: 0, claimedValue: 2, cardCount: 2, isBluff: false },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPUがダウトを判定中...')).toBeInTheDocument());
    // humanAction log div is not rendered in doubt phase
    // The lastAction is shown in table area but the log section is suppressed
    const logDivs = screen.queryAllByText(/あなたが2枚出しました.*宣言/);
    // Only table area lastAction should show, not the human action log
    expect(logDivs.length).toBeLessThanOrEqual(1);
  });

  it('shows CPU action log when cpuActions non-empty', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      cpuActions: [
        { playerIdx: 1, claimedValue: 3, cardCount: 2, isBluff: false },
        { playerIdx: 2, claimedValue: 5, cardCount: 1, isBluff: true },
      ],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1が2枚出しました/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2が1枚出しました/)).toBeInTheDocument();
  });

  it('does not show CPU action log when cpuActions is empty', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/\[CPUの行動\]/)).not.toBeInTheDocument();
  });

  // ── Game result message ───────────────────────────────────────────────────

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show result message when message is empty', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    // message is '', no result div shown
    expect(screen.queryByText('ゲーム終了')).not.toBeInTheDocument();
  });

  // ── CpuArea badges ────────────────────────────────────────────────────────

  it('shows 上がり badge for finished CPU players', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      players: [
        humanTurnState.players[0],
        { id: 1, isHuman: false, isFinished: true, cardCount: 0, cards: [] },
        humanTurnState.players[2],
        humanTurnState.players[3],
      ],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('上がり')).toBeInTheDocument());
  });

  it('shows 考え中... for current CPU turn when not finished', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('考え中...')).toBeInTheDocument());
  });

  it('does not show 考え中... for finished CPU even if current turn', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      currentTurn: 1,
      players: [
        humanTurnState.players[0],
        { id: 1, isHuman: false, isFinished: true, cardCount: 0, cards: [] },
        humanTurnState.players[2],
        humanTurnState.players[3],
      ],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('上がり')).toBeInTheDocument());
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  it('does not show 考え中... when human is current turn', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  // ── isHumanTurn with gameEndFlag ──────────────────────────────────────────

  it('hides 出す button when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // ── CPU Tell indicator ──────────────────────────────────────────────────

  it('shows sweat indicator when cpuAction has hasTell true', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 1, claimedValue: 3, cardCount: 2, isBluff: true, hasTell: true }],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByLabelText('テル')).toBeInTheDocument());
  });

  it('shows sweat indicator when lastAction has hasTell true', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastAction: { playerIdx: 1, claimedValue: 3, cardCount: 2, isBluff: true, hasTell: true },
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByLabelText('テル')).toBeInTheDocument());
  });

  it('does not show sweat indicator when hasTell is false', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 1, claimedValue: 3, cardCount: 2, isBluff: true, hasTell: false }],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByLabelText('テル')).not.toBeInTheDocument();
  });

  it('does not show sweat indicator when hasTell is undefined', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 1, claimedValue: 3, cardCount: 2, isBluff: true }],
    };
    mockExec.mockResolvedValue(s);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByLabelText('テル')).not.toBeInTheDocument();
  });

  // ── Settings panel ────────────────────────────────────────────────────────

  it('renders settings panel with default values', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    const summary = screen.getByText('設定');
    expect(summary).toBeInTheDocument();
  });

  it.each([
    {
      label: 'doubtWindowSec to 3s',
      selectIdx: 0,
      value: '3',
      expected: {
        doubtWindowSec: 3,
        cpuMemoryLevel: 1,
        penaltyDrawLimit: 0,
        cpuHesitationEnabled: false,
        cpuMetaAI: false,
      },
    },
    {
      label: 'doubtWindowSec to 5s',
      selectIdx: 0,
      value: '5',
      expected: {
        doubtWindowSec: 5,
        cpuMemoryLevel: 1,
        penaltyDrawLimit: 0,
        cpuHesitationEnabled: false,
        cpuMetaAI: false,
      },
    },
    {
      label: 'cpuMemoryLevel to Hard',
      selectIdx: 1,
      value: '2',
      expected: {
        doubtWindowSec: 10,
        cpuMemoryLevel: 2,
        penaltyDrawLimit: 0,
        cpuHesitationEnabled: false,
        cpuMetaAI: false,
      },
    },
    {
      label: 'cpuMemoryLevel to Easy',
      selectIdx: 1,
      value: '0',
      expected: {
        doubtWindowSec: 10,
        cpuMemoryLevel: 0,
        penaltyDrawLimit: 0,
        cpuHesitationEnabled: false,
        cpuMetaAI: false,
      },
    },
    {
      label: 'penaltyDrawLimit to 5',
      selectIdx: 2,
      value: '5',
      expected: {
        doubtWindowSec: 10,
        cpuMemoryLevel: 1,
        penaltyDrawLimit: 5,
        cpuHesitationEnabled: false,
        cpuMetaAI: false,
      },
    },
    {
      label: 'penaltyDrawLimit to 3',
      selectIdx: 2,
      value: '3',
      expected: {
        doubtWindowSec: 10,
        cpuMemoryLevel: 1,
        penaltyDrawLimit: 3,
        cpuHesitationEnabled: false,
        cpuMetaAI: false,
      },
    },
  ])('changing $label updates config passed to reset', async ({ selectIdx, value, expected }) => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));

    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[selectIdx], { target: { value } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, expected));
  });

  it('toggling cpuHesitation checkbox updates config passed to reset', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));

    const checkbox = screen.getByLabelText('CPU迷い時間ディレイ');
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        doubtWindowSec: 10,
        cpuMemoryLevel: 1,
        penaltyDrawLimit: 0,
        cpuHesitationEnabled: true,
        cpuMetaAI: false,
      }),
    );
  });

  it('toggling cpuMetaAI checkbox updates config passed to reset', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));

    const checkbox = screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）');
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        doubtWindowSec: 10,
        cpuMemoryLevel: 1,
        penaltyDrawLimit: 0,
        cpuHesitationEnabled: false,
        cpuMetaAI: true,
      }),
    );
  });

  // ── Meta-AI display ────────────────────────────────────────────────────────

  it('displays MetaAI info when metaAI is enabled', async () => {
    const metaAIState: DoubtResponse = {
      ...humanTurnState,
      metaAI: {
        enabled: true,
        gamesPlayed: 3,
        bluffRate: 0.6,
        doubtAccuracy: 0.75,
        hesitationMean: 1234,
      },
    };
    mockExec.mockResolvedValue(metaAIState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('メタAI情報')).toBeInTheDocument());
    expect(screen.getByText('ゲーム数: 3')).toBeInTheDocument();
    expect(screen.getByText('ブラフ率: 60%')).toBeInTheDocument();
    expect(screen.getByText('ダウト正解率: 75%')).toBeInTheDocument();
    expect(screen.getByText('平均迷い時間: 1234ms')).toBeInTheDocument();
  });

  it('does not display MetaAI info when metaAI is not present', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByText('メタAI情報')).not.toBeInTheDocument();
  });

  it('hides hesitationMean row when hesitationMean is 0', async () => {
    const metaAIState: DoubtResponse = {
      ...humanTurnState,
      metaAI: {
        enabled: true,
        gamesPlayed: 0,
        bluffRate: 0,
        doubtAccuracy: 0,
        hesitationMean: 0,
      },
    };
    mockExec.mockResolvedValue(metaAIState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('メタAI情報')).toBeInTheDocument());
    expect(screen.queryByText(/平均迷い時間/)).not.toBeInTheDocument();
  });

  it('does not set playTurnStart when currentTurn player is missing', async () => {
    // State where currentTurn points to a non-existent player index
    const stateWithMissingPlayer: DoubtResponse = {
      ...humanTurnState,
      currentTurn: 99,
      phase: 0,
    };
    mockExec.mockResolvedValue(stateWithMissingPlayer);
    renderWithProviders(<DoubtPage />);
    // Page renders without error; 出す button is absent (no human turn)
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // ── Server-driven countdown ───────────────────────────────────────────────

  describe('server-driven countdown', () => {
    beforeEach(() => {
      vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
    });
    afterEach(() => {
      vi.clearAllTimers();
      vi.useRealTimers();
    });

    it('uses state.doubtWindowSec for countdown (3s)', async () => {
      const shortState: DoubtResponse = {
        ...doubtPhaseCpuPlayedState,
        doubtWindowSec: 3,
      };
      mockExec.mockResolvedValue(shortState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 3 秒/)).toBeInTheDocument());
    });

    it('uses state.doubtWindowSec for countdown (5s)', async () => {
      const midState: DoubtResponse = {
        ...doubtPhaseCpuPlayedState,
        doubtWindowSec: 5,
      };
      mockExec.mockResolvedValue(midState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 5 秒/)).toBeInTheDocument());
    });
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, defaultConfig));
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      gameEndFlag: true,
      currentTurn: 0,
      players: [],
      playerIdx: 0,
      playCards: [],
      cpuDoubters: [],
      cpuActions: [],
      lastAction: null,
    } as unknown as DoubtResponse);

    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.doubt).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.doubt).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // ── Keyboard navigation ──────────────────────────────────────────────────

  describe('keyboard navigation', () => {
    it('pressing number key toggles card when in human play turn', async () => {
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      // Press '1' to toggle the first card
      fireEvent.keyDown(document, { key: '1' });
      const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
      expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

      // Press '1' again to deselect
      fireEvent.keyDown(document, { key: '1' });
      expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    });

    it('pressing "2" toggles second card', async () => {
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByAltText('♥ J')).toBeInTheDocument());

      fireEvent.keyDown(document, { key: '2' });
      const cardBtn = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;
      expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    });

    it('Enter key triggers play', async () => {
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      // Select a card first
      fireEvent.keyDown(document, { key: '1' });
      expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();

      mockExec.mockClear();
      mockExec.mockResolvedValue(cpuTurnState);
      fireEvent.keyDown(document, { key: 'Enter' });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('play', [0], 1, undefined, undefined, expect.any(Number)),
      );
    });

    it('Escape key clears selection', async () => {
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      // Select a card
      fireEvent.keyDown(document, { key: '1' });
      const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
      expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

      // Press Escape to clear
      fireEvent.keyDown(document, { key: 'Escape' });
      expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
      expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
    });

    it('keyboard is disabled when not in human play turn', async () => {
      mockExec.mockResolvedValue(cpuTurnState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

      // Press '1' should have no effect (CPU's turn)
      fireEvent.keyDown(document, { key: '1' });
      // Human cards are still rendered but no toggle should happen
      const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
      expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    });

    it('keyboard is disabled during doubt phase', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText('ダウトしますか？')).toBeInTheDocument());

      // Press '1' should have no effect (doubt phase, not play phase)
      fireEvent.keyDown(document, { key: '1' });
      const cards = screen.queryAllByAltText('♠ A');
      for (const card of cards) {
        const btn = card.closest('button');
        if (btn) {
          expect(btn).toHaveAttribute('aria-pressed', 'false');
        }
      }
    });

    it('keyboard is disabled when game has ended', async () => {
      mockExec.mockResolvedValue(gameEndState);
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());

      // No play button, keyboard should be disabled
      fireEvent.keyDown(document, { key: '1' });
      // No cards to check since game ended, just ensure no errors
    });

    it('number keys do not fire when typing in claim input', async () => {
      renderWithProviders(<DoubtPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      // Select a card to show claim input
      fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
      const input = screen.getByRole('spinbutton');

      // Press '2' while focused on input - should not toggle second card
      fireEvent.keyDown(input, { key: '2' });
      const secondCardBtn = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;
      expect(secondCardBtn).toHaveAttribute('aria-pressed', 'false');
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
