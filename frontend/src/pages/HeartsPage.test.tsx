import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, heartsApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeHeartsState } from '../test/stateFactories';
import type { HeartsResponse } from '../types/card';
import { HeartsPage } from './HeartsPage';

vi.mock('../api/gameApi', () => ({
  heartsApi: { exec: vi.fn() },
  actionLogApi: { hearts: vi.fn() },
}));

const mockExec = vi.mocked(heartsApi.exec);

const playPhaseState = makeHeartsState();

const passPhaseState = makeHeartsState({ phase: 0, passDirection: 0 });

const trickEndState = makeHeartsState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
});

const roundEndState = makeHeartsState({ phase: 3 });

const gameEndState = makeHeartsState({
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
});

const gameEndByFlagState = makeHeartsState({
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
});

const heartsBrokenState = makeHeartsState({ heartsBroken: true });

const cpuTurnState = makeHeartsState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('HeartsPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<HeartsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        omnibusJD: false,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('shows shoot-the-moon alert when one CPU monopolises ≥13 round points', async () => {
    const moonState = makeHeartsState({
      players: [
        { id: 0, isHuman: true, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
        { id: 1, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
        { id: 2, isHuman: false, cardCount: 13, cards: [], roundScore: 14, cumulativeScore: 14, trickCount: 5 },
        { id: 3, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      ],
    });
    mockExec.mockResolvedValue(moonState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getAllByTestId('hearts-shoot-the-moon-badge').length).toBeGreaterThan(0));
  });

  it('does not show shoot-the-moon alert when points are split between players', async () => {
    const splitState = makeHeartsState({
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
        { id: 1, isHuman: false, cardCount: 13, cards: [], roundScore: 8, cumulativeScore: 8, trickCount: 3 },
        { id: 2, isHuman: false, cardCount: 13, cards: [], roundScore: 8, cumulativeScore: 8, trickCount: 3 },
        { id: 3, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      ],
    });
    mockExec.mockResolvedValue(splitState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByTestId('hearts-shoot-the-moon-badge')).not.toBeInTheDocument();
  });

  it('renders pass phase with pass button and the recipient name', async () => {
    mockExec.mockResolvedValue(passPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
      // Left pass from seat 0 goes to seat 1 (CPU 1).
      expect(screen.getByText('パス方向: 左 → CPU 1 へ')).toBeInTheDocument();
    });
  });

  it('shows the pass-progress badge from zero cards selected', async () => {
    mockExec.mockResolvedValue(passPhaseState);
    renderWithProviders(<HeartsPage />);
    const badge = await screen.findByTestId('hearts-pass-progress');
    expect(badge).toHaveTextContent('0/3');
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

  it('shows a N/3 selection-progress pill during the pass phase', async () => {
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

    // Pill shows 0/3 before any selection, then updates as cards are picked.
    expect(screen.getByTestId('hearts-pass-progress')).toHaveTextContent('0/3');

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('hearts-pass-progress')).toHaveTextContent('1/3');

    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('hearts-pass-progress')).toHaveTextContent('2/3');
  });

  it('shows a directional arrow glyph for the pass direction', async () => {
    mockExec.mockResolvedValue(passPhaseState); // left
    renderWithProviders(<HeartsPage />);
    const arrow = await screen.findByTestId('hearts-pass-arrow');
    expect(arrow).toHaveTextContent('←');
  });

  it('shows the remaining-card count until 3 are selected', async () => {
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

    const badge = screen.getByTestId('hearts-pass-progress');
    expect(badge).toHaveTextContent('あと3枚');

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);
    // All three selected: remaining hint disappears.
    expect(badge).not.toHaveTextContent('あと');
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
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 100,
        omnibusJD: false,
      }),
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
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
        omnibusJD: false,
      }),
    );
  });

  it('settings panel toggles omnibusJD', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const checkbox = screen.getByRole('checkbox', { name: 'オムニバス (J♦ = -10点)' });
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        omnibusJD: true,
      }),
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
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        omnibusJD: false,
      }),
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

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
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
    await waitFor(() => expect(screen.getByText('パス方向: 右 → CPU 3 へ')).toBeInTheDocument());
  });

  it('shows pass direction across', async () => {
    mockExec.mockResolvedValue({ ...passPhaseState, passDirection: 2 });
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('パス方向: 対面 → CPU 2 へ')).toBeInTheDocument());
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
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        omnibusJD: false,
      }),
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

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn during pass phase', async () => {
    mockExec.mockResolvedValue(passPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn when human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // ── Keyboard navigation ────────────────────────────────────────────────────

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('pressing number key toggles card in pass phase', async () => {
    mockExec.mockResolvedValue(passPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('Enter key triggers pass in pass phase', async () => {
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

    fireEvent.keyDown(document, { key: '1' });
    fireEvent.keyDown(document, { key: '2' });
    fireEvent.keyDown(document, { key: '3' });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', [0, 1, 2]));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.hearts).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<HeartsPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with 2-row hand grid', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="ht-player-hand"]');
      expect(hand).toBeInTheDocument();
      const rows = hand?.querySelectorAll('[data-testid="hand-row"]');
      expect(rows?.length).toBeGreaterThanOrEqual(1);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with wrapping hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="ht-player-hand"]');
      expect(hand?.className).toContain('flex-wrap');
      expect(hand?.querySelectorAll('[data-testid="hand-row"]')).toHaveLength(0);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      await waitFor(() => {
        const allDetails = container.querySelectorAll('details');
        const cpuDetails = Array.from(allDetails).find((d) =>
          d.querySelector('summary')?.textContent?.includes('CPU対戦相手'),
        );
        expect(cpuDetails).toBeTruthy();
      });
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders score table as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="ht-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('スコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  describe('legal-move highlight', () => {
    const followSuitState = makeHeartsState({
      trickNumber: 3,
      heartsBroken: true,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: { design: 'DIAMOND', value: 3 } }],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [
            { design: 'DIAMOND', value: 8 },
            { design: 'SPADE', value: 9 },
          ],
          roundScore: 0,
          cumulativeScore: 0,
          trickCount: 0,
        },
        { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
        { id: 2, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
        { id: 3, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      ],
    });

    it('rings only follow-suit cards and leaves the off-suit card clickable (no hard block)', async () => {
      mockExec.mockResolvedValue(followSuitState);
      renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByAltText('♦ 8')).toBeInTheDocument());

      const diamond = screen.getByAltText('♦ 8').closest('button') as HTMLButtonElement;
      const spade = screen.getByAltText('♠ 9').closest('button') as HTMLButtonElement;

      // Legal follow-suit card is ringed (data-legal).
      expect(diamond).toHaveAttribute('data-legal', 'true');
      expect(diamond.className).toContain('ring-ds-success');

      // Illegal off-suit card gets no ring, but stays clickable — the server remains
      // authoritative (the highlight is a visual aid, not a hard block).
      expect(spade).not.toHaveAttribute('data-legal');
      expect(spade.className).not.toContain('ring-ds-success');
      expect(spade).not.toHaveAttribute('aria-disabled');
      expect(spade.className).not.toContain('cursor-not-allowed');
    });

    it('rings only non-heart leads before hearts are broken (heart lead left un-ringed)', async () => {
      const leadState = makeHeartsState({
        trickNumber: 3,
        heartsBroken: false,
        currentPlayerIdx: 0,
        currentTrick: [],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [
              { design: 'HEART', value: 5 },
              { design: 'SPADE', value: 9 },
            ],
            roundScore: 0,
            cumulativeScore: 0,
            trickCount: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
          { id: 2, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
          { id: 3, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
        ],
      });
      mockExec.mockResolvedValue(leadState);
      renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByAltText('♥ 5')).toBeInTheDocument());

      const heart = screen.getByAltText('♥ 5').closest('button') as HTMLButtonElement;
      const spade = screen.getByAltText('♠ 9').closest('button') as HTMLButtonElement;

      expect(spade).toHaveAttribute('data-legal', 'true');
      expect(spade.className).toContain('ring-ds-success');
      expect(heart).not.toHaveAttribute('data-legal');
      expect(heart.className).not.toContain('ring-ds-success');
    });

    it('does not ring or dim cards when it is not the human turn', async () => {
      mockExec.mockResolvedValue(cpuTurnState);
      renderWithProviders(<HeartsPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      const spadeA = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
      expect(spadeA).not.toHaveAttribute('data-legal');
      expect(spadeA.className).not.toContain('opacity-50');
    });
  });
});
