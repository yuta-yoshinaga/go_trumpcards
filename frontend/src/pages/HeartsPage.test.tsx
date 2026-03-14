import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, heartsApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { HeartsResponse } from '../types/card';
import { HeartsPage } from './HeartsPage';

vi.mock('../api/gameApi', () => ({
  heartsApi: { exec: vi.fn() },
  actionLogApi: { hearts: vi.fn() },
}));

const mockExec = vi.mocked(heartsApi.exec);

const playPhaseState: HeartsResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], roundScore: 3, cumulativeScore: 10, trickCount: 1 },
    { id: 2, isHuman: false, cardCount: 13, cards: [], roundScore: 5, cumulativeScore: 20, trickCount: 2 },
    { id: 3, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 5, trickCount: 0 },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  heartsBroken: false,
  passDirection: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100 },
};

const passPhaseState: HeartsResponse = {
  ...playPhaseState,
  phase: 0,
  passDirection: 0,
};

const trickEndState: HeartsResponse = {
  ...playPhaseState,
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: HeartsResponse = {
  ...playPhaseState,
  phase: 3,
};

const gameEndState: HeartsResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
};

const gameEndByFlagState: HeartsResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
};

const heartsBrokenState: HeartsResponse = {
  ...playPhaseState,
  heartsBroken: true,
};

const cpuTurnState: HeartsResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('HeartsPage', () => {
  it('renders null when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = renderWithProviders(<HeartsPage />);
    expect(container.firstChild).toBeNull();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 100 }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('renders pass phase with pass button', async () => {
    mockExec.mockResolvedValue(passPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
      expect(screen.getByText('パス方向: 左')).toBeInTheDocument();
    });
  });

  it('pass button disabled when not 3 cards selected', async () => {
    mockExec.mockResolvedValue(passPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button enabled when 3 cards selected', async () => {
    mockExec.mockResolvedValue({
      ...passPhaseState,
      players: [
        {
          ...passPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 11 },
            { design: 'CLOVER', value: 5 },
          ],
        },
        ...passPhaseState.players.slice(1),
      ],
    });
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);

    expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled();
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByText('ゲーム終了！')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByText('ゲーム終了！')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 100 }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '200' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '♠ A');

    const cardBtn2 = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '♥ J');
  });

  it('reset button calls exec', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 100 }),
    );
  });

  it('score table shows all players', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('shows hearts broken text', async () => {
    mockExec.mockResolvedValue(heartsBrokenState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('ハートブレイク済')).toBeInTheDocument());
  });

  it('shows hearts not broken text', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('ハート未ブレイク')).toBeInTheDocument());
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByText('現在のトリック')).toBeInTheDocument();
      expect(screen.getByAltText('♦ 3')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
    });
  });

  it('does not show current trick when empty', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('現在のトリック')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*13枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*13枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*13枚/)).toBeInTheDocument();
    });
  });

  it('shows loading state', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: HeartsResponse) => void;
    const slow = new Promise<HeartsResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();
    expect(screen.getByText('処理中...')).toBeInTheDocument();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('calls play command when 出す is clicked', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('calls pass command when パス is clicked', async () => {
    mockExec.mockResolvedValue({
      ...passPhaseState,
      players: [
        {
          ...passPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 11 },
            { design: 'CLOVER', value: 5 },
          ],
        },
        ...passPhaseState.players.slice(1),
      ],
    });
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', [0, 1, 2]));
  });

  it('calls next when 次のトリック is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('calls nextround when 次のラウンド is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows pass direction right', async () => {
    mockExec.mockResolvedValue({ ...passPhaseState, passDirection: 1 });
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('パス方向: 右')).toBeInTheDocument());
  });

  it('shows pass direction across', async () => {
    mockExec.mockResolvedValue({ ...passPhaseState, passDirection: 2 });
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('パス方向: 対面')).toBeInTheDocument());
  });

  it('shows pass direction none', async () => {
    mockExec.mockResolvedValue({ ...passPhaseState, passDirection: 3 });
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('パスなし')).toBeInTheDocument());
  });

  it('does not show pass direction in non-pass phase', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText(/パス方向/)).not.toBeInTheDocument();
    expect(screen.queryByText('パスなし')).not.toBeInTheDocument();
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('トリック 1')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('ゲーム終了')).not.toBeInTheDocument();
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 100 }),
    );
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.hearts).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.hearts).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  it('does not show pass button in play phase', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    // The "パス" button for hearts pass phase should NOT appear in play phase
    // Only "出す" and "リセット" should be present
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
  });

  it('disables buttons while loading', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: HeartsResponse) => void;
    const slow = new Promise<HeartsResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('trick card shows player name with fallback', async () => {
    const stateWithBadIdx: HeartsResponse = {
      ...trickEndState,
      currentTrick: [{ playerIdx: 99, card: { design: 'SPADE', value: 1 } }],
    };
    mockExec.mockResolvedValue(stateWithBadIdx);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('CPU 99')).toBeInTheDocument());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: HeartsResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('♠ A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
