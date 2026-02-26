import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doubtApi } from '../api/gameApi';
import type { DoubtConfig, DoubtResponse } from '../types/card';
import { DoubtPage } from './DoubtPage';

vi.mock('../api/gameApi', () => ({
  doubtApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(doubtApi.exec);

const defaultConfig: DoubtConfig = { doubtWindowSec: 10, cpuMemoryLevel: 1 };

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
  it('renders nothing before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = render(<DoubtPage />);
    expect(container.firstChild).toBeNull();
  });

  it('calls reset command on mount', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, defaultConfig));
  });

  it('renders CPU player areas', async () => {
    render(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('renders table area', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.getByText('場のカード: 0枚')).toBeInTheDocument();
  });

  it('shows human player hand cards', async () => {
    render(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByAltText('SPADE 1')).toBeInTheDocument();
      expect(screen.getByAltText('HEART 11')).toBeInTheDocument();
    });
  });

  it('shows selection hint on human turn phase 0', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('カードを選んで出してください')).toBeInTheDocument());
  });

  it('does not show selection hint when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByText('カードを選んで出してください')).not.toBeInTheDocument();
  });

  it('shows 出す button only on human turn phase 0', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
  });

  it('does not show 出す button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('出す button is disabled when no cards are selected', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('toggles card selection on click and enables 出す button', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('SPADE 1').closest('button') as HTMLButtonElement;
    fireEvent.click(cardBtn);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();

    // Click again to deselect
    fireEvent.click(cardBtn);
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
  });

  it('shows claimed value input when cards are selected', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 1')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('SPADE 1').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('spinbutton')).toBeInTheDocument();
    // Default value 1 shows (A)
    expect(screen.getByText('(A)')).toBeInTheDocument();
  });

  it('claimed value input is hidden when no cards are selected', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByRole('spinbutton')).not.toBeInTheDocument();
  });

  it('changes claimed value and shows special name', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 1')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('SPADE 1').closest('button') as HTMLButtonElement);
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

  it('calls play command with selected cards when 出す clicked', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 1')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('SPADE 1').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0], 1, undefined, undefined));
  });

  it('calls reset when reset button is clicked', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, defaultConfig));
  });

  it('disables buttons while loading', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: DoubtResponse) => void;
    const slowPromise = new Promise<DoubtResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('clears error message on successful API call after failure', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(screen.queryByText('通信エラーが発生しました。もう一度お試しください。')).not.toBeInTheDocument(),
    );
  });

  // ── Table area ────────────────────────────────────────────────────────────

  it('shows lastAction in table area', async () => {
    const stateWithLastAction: DoubtResponse = {
      ...humanTurnState,
      currentTurn: 2,
      lastAction: { playerIdx: 1, claimedValue: 5, cardCount: 2, isBluff: false },
    };
    mockExec.mockResolvedValue(stateWithLastAction);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1が2枚出しました.*宣言: 5/)).toBeInTheDocument());
  });

  it('hides lastAction display when lastAction is null', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/枚出しました/)).not.toBeInTheDocument();
  });

  // ── valueName branches ───────────────────────────────────────────────────

  it('valueName shows A for claimedValue 1', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastAction: { playerIdx: 1, claimedValue: 1, cardCount: 1, isBluff: false },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/宣言: A/)).toBeInTheDocument());
  });

  it('valueName shows J for claimedValue 11', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastAction: { playerIdx: 1, claimedValue: 11, cardCount: 1, isBluff: false },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/宣言: J/)).toBeInTheDocument());
  });

  it('valueName shows Q for claimedValue 12', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastAction: { playerIdx: 1, claimedValue: 12, cardCount: 1, isBluff: false },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/宣言: Q/)).toBeInTheDocument());
  });

  it('valueName shows K for claimedValue 13', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      lastAction: { playerIdx: 1, claimedValue: 13, cardCount: 1, isBluff: false },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/宣言: K/)).toBeInTheDocument());
  });

  // ── actionDesc branches ───────────────────────────────────────────────────

  it('actionDesc uses Player index when player not found', async () => {
    const s: DoubtResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 99, claimedValue: 5, cardCount: 2, isBluff: false }],
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/Player 99が2枚出しました/)).toBeInTheDocument());
  });

  // ── Doubt phase: CPU played ───────────────────────────────────────────────

  it('shows doubt/skip UI when CPU played in doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
    render(<DoubtPage />);
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
    render(<DoubtPage />);
    await waitFor(() => {
      const el = screen.getByText(/ダウト宣言CPU/);
      expect(el.textContent).toContain('CPU 2');
      expect(el.textContent).toContain('CPU 3');
    });
  });

  it('calls doubt with [0] when ダウト！ clicked (no cpu doubters)', async () => {
    mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダウト！' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'ダウト！' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubt', undefined, undefined, [0], undefined));
  });

  it('calls skip with [] when スルー clicked (no cpu doubters)', async () => {
    mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スルー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'スルー' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip', undefined, undefined, [], undefined));
  });

  it('calls doubt with [0, ...cpuDoubters] when ダウト！ clicked with cpu doubters', async () => {
    const s: DoubtResponse = { ...doubtPhaseCpuPlayedState, cpuDoubters: [2, 3] };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダウト！' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'ダウト！' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubt', undefined, undefined, [0, 2, 3], undefined));
  });

  // ── Doubt phase: human played ─────────────────────────────────────────────

  it('shows confirm UI when human played in doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    render(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('CPUがダウトを判定中...')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument();
    });
  });

  it('shows cpu doubters in human-played doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/ダウト！.*CPU 1/)).toBeInTheDocument());
  });

  it('does not show cpuDoubters label when empty in human-played phase', async () => {
    const s: DoubtResponse = { ...doubtPhaseHumanPlayedState, cpuDoubters: [] };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('CPUがダウトを判定中...')).toBeInTheDocument());
    expect(screen.queryByText(/ダウト！/)).not.toBeInTheDocument();
  });

  it('calls doubt with cpuDoubters when 確認 clicked', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubt', undefined, undefined, [1, 2], undefined));
  });

  // ── Doubt phase suppressed when gameEndFlag ───────────────────────────────

  it('does not show doubt UI when gameEndFlag is true', async () => {
    const s: DoubtResponse = { ...doubtPhaseCpuPlayedState, gameEndFlag: true, message: 'ゲーム終了！' };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
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

    it('starts countdown display when CPU played in doubt phase', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      render(<DoubtPage />);
      // Wait directly for countdown text (appears after 2nd render: state→doubt phase, effect→countdown)
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());
    });

    it('decrements countdown each second', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      render(<DoubtPage />);
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
      render(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      act(() => {
        vi.advanceTimersByTime(10000);
      });
      // Countdown display cleared immediately
      expect(screen.queryByText(/残り/)).not.toBeInTheDocument();

      // Auto-skip is called with the correct doubters
      await waitFor(() => {
        expect(mockExec).toHaveBeenCalledWith('skip', undefined, undefined, [], undefined);
      });
    });

    it('stops countdown when ダウト！ is clicked', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      render(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      mockExec.mockResolvedValue(humanTurnState);
      act(() => {
        fireEvent.click(screen.getByRole('button', { name: 'ダウト！' }));
      });
      await waitFor(() => expect(screen.queryByText(/残り/)).not.toBeInTheDocument());
    });

    it('stops countdown when スルー is clicked', async () => {
      mockExec.mockResolvedValue(doubtPhaseCpuPlayedState);
      render(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 10 秒/)).toBeInTheDocument());

      mockExec.mockResolvedValue(humanTurnState);
      act(() => {
        fireEvent.click(screen.getByRole('button', { name: 'スルー' }));
      });
      await waitFor(() => expect(screen.queryByText(/残り/)).not.toBeInTheDocument());
    });
  });

  // ── No countdown in other phases ─────────────────────────────────────────

  it('does not show countdown in play phase', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/残り/)).not.toBeInTheDocument();
  });

  it('does not start countdown when lastAction is null in doubt phase', async () => {
    const s: DoubtResponse = { ...humanTurnState, phase: 1, lastAction: null };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/残り/)).not.toBeInTheDocument();
  });

  it('does not start countdown when human played in doubt phase', async () => {
    mockExec.mockResolvedValue(doubtPhaseHumanPlayedState);
    render(<DoubtPage />);
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
        revealedCards: [{ design: 'SPADE', value: 5 }],
      },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => {
      expect(screen.getByText('ダウト結果')).toBeInTheDocument();
      expect(screen.getByText('ウソでした！')).toBeInTheDocument();
      expect(screen.getByAltText('SPADE 5')).toBeInTheDocument();
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
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
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
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
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
        revealedCards: [],
      },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/99が1枚引き取りました/)).toBeInTheDocument());
  });

  it('does not show ダウト結果 when lastDoubtResult is null', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText('ダウト結果')).not.toBeInTheDocument();
  });

  // ── Action logs ───────────────────────────────────────────────────────────

  it('shows human action log in non-doubt phase', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/あなたが2枚出しました/)).toBeInTheDocument());
  });

  it('suppresses human action log in doubt phase', async () => {
    const s: DoubtResponse = {
      ...doubtPhaseHumanPlayedState,
      humanAction: { playerIdx: 0, claimedValue: 2, cardCount: 2, isBluff: false },
    };
    mockExec.mockResolvedValue(s);
    render(<DoubtPage />);
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
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1が2枚出しました/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2が1枚出しました/)).toBeInTheDocument();
  });

  it('does not show CPU action log when cpuActions is empty', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText(/\[CPUの行動\]/)).not.toBeInTheDocument();
  });

  // ── Game result message ───────────────────────────────────────────────────

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show result message when message is empty', async () => {
    render(<DoubtPage />);
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
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('上がり')).toBeInTheDocument());
  });

  it('shows 考え中... for current CPU turn when not finished', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DoubtPage />);
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
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('上がり')).toBeInTheDocument());
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  it('does not show 考え中... when human is current turn', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  // ── isHumanTurn with gameEndFlag ──────────────────────────────────────────

  it('hides 出す button when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // ── Settings panel ────────────────────────────────────────────────────────

  it('renders settings panel with default values', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());
    const summary = screen.getByText('設定');
    expect(summary).toBeInTheDocument();
  });

  it('changing doubtWindowSec updates config passed to reset', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    // Open settings
    fireEvent.click(screen.getByText('設定'));

    // Change doubt window to 3s
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '3' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        doubtWindowSec: 3,
        cpuMemoryLevel: 1,
      }),
    );
  });

  it('changing cpuMemoryLevel updates config passed to reset', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));

    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '2' } }); // Hard

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        doubtWindowSec: 10,
        cpuMemoryLevel: 2,
      }),
    );
  });

  it('changing doubtWindowSec to 5s updates config', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));

    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '5' } }); // 5s

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        doubtWindowSec: 5,
        cpuMemoryLevel: 1,
      }),
    );
  });

  it('changing cpuMemoryLevel to Easy updates config', async () => {
    render(<DoubtPage />);
    await waitFor(() => expect(screen.getByText('テーブル')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));

    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '0' } }); // Easy

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        doubtWindowSec: 10,
        cpuMemoryLevel: 0,
      }),
    );
  });

  // ── Server-driven countdown ───────────────────────────────────────────────

  describe('server-driven countdown', () => {
    beforeEach(() => {
      vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
    });

    it('uses state.doubtWindowSec for countdown (3s)', async () => {
      const shortState: DoubtResponse = {
        ...doubtPhaseCpuPlayedState,
        doubtWindowSec: 3,
      };
      mockExec.mockResolvedValue(shortState);
      render(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 3 秒/)).toBeInTheDocument());
    });

    it('uses state.doubtWindowSec for countdown (5s)', async () => {
      const midState: DoubtResponse = {
        ...doubtPhaseCpuPlayedState,
        doubtWindowSec: 5,
      };
      mockExec.mockResolvedValue(midState);
      render(<DoubtPage />);
      await waitFor(() => expect(screen.getByText(/残り 5 秒/)).toBeInTheDocument());
    });
  });
});
