import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, spiderApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpiderResponse, SpiderTableauCard } from '../types/card';
import { SpiderPage } from './SpiderPage';

vi.mock('../api/gameApi', () => ({
  spiderApi: { exec: vi.fn() },
  actionLogApi: { spider: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const playSoundMock = vi.fn();
vi.mock('../providers/SoundProvider', async () => {
  const actual = await vi.importActual<typeof import('../providers/SoundProvider')>('../providers/SoundProvider');
  return {
    ...actual,
    useSound: () => ({
      playSound: playSoundMock,
      muted: false,
      toggleMute: vi.fn(),
      claimExecSound: vi.fn(),
      consumeExecClaim: () => false,
    }),
    // GamePageShell's central taps read useOptionalSound; route it to the
    // same spy so the fanfare it now owns is observable.
    useOptionalSound: () => ({
      playSound: playSoundMock,
      muted: false,
      toggleMute: vi.fn(),
      claimExecSound: vi.fn(),
      consumeExecClaim: () => false,
    }),
  };
});

const mockExec = vi.mocked(spiderApi.exec);

function makeTableau(cols: SpiderTableauCard[][]): SpiderTableauCard[][] {
  const result: SpiderTableauCard[][] = [];
  for (let i = 0; i < 10; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SpiderResponse = {
  tableau: makeTableau([
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 5), faceUp: true },
    ],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
    [],
  ]),
  stockCount: 50,
  completedSuits: 0,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  score: 500,
  difficulty: 1,
  message: '',
};

const gameClearState: SpiderResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'spider.gameClear',
  messageParams: { moveCount: '42', score: '500' },
};

const gameOverState: SpiderResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'spider.gameOver',
};

beforeEach(() => {
  localStorage.clear();
  // Re-suppress the first-visit tutorial dialog: the global setup sets this in its
  // own beforeEach, which runs before ours, so our clear() would otherwise wipe it.
  localStorage.setItem('tutorial_no_suggest', 'true');
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  playSoundMock.mockClear();
});

describe('SpiderPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpiderPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(50\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders score', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/スコア: 500/));
  });

  it('renders completed suits count', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/完成: 0\/8/));
  });

  it('renders tableau columns without index headers', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // Column index headers should not be rendered
  });

  it('clicking deal button dispatches deal when no empty columns exist', async () => {
    const filledTableauState: SpiderResponse = {
      ...playingState,
      tableau: makeTableau(Array.from({ length: 10 }, () => [{ card: card('SPADE', 13), faceUp: true }])),
    };
    mockExec.mockResolvedValue(filledTableauState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '配る' }).length).toBeGreaterThanOrEqual(1));

    mockExec.mockClear();
    mockExec.mockResolvedValue(filledTableauState);
    // The footer button is the last one
    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('announces and flashes when a new suit is completed', async () => {
    const filled: SpiderResponse = {
      ...playingState,
      tableau: makeTableau(Array.from({ length: 10 }, () => [{ card: card('SPADE', 13), faceUp: true }])),
    };
    mockExec.mockResolvedValue(filled); // completedSuits 0
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '配る' }).length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByTestId('spd-suit-complete')).not.toBeInTheDocument();

    mockExec.mockResolvedValue({ ...filled, completedSuits: 1 });
    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);
    await waitFor(() => expect(screen.getByTestId('spd-suit-complete')).toBeInTheDocument());
  });

  it('clicking deal with empty columns triggers shake on empty placeholders and skips API', async () => {
    // playingState has empty columns 2..9
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('spd-empty-col-2')).toBeInTheDocument());

    // Before click: no shake class
    expect(screen.getByTestId('spd-empty-col-2').className).not.toContain('animate-shake');

    mockExec.mockClear();
    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);

    // After click: shake class applied to every empty column placeholder
    await waitFor(() => {
      expect(screen.getByTestId('spd-empty-col-2').className).toContain('animate-shake');
    });
    expect(screen.getByTestId('spd-empty-col-9').className).toContain('animate-shake');

    // API was NOT called with deal
    expect(mockExec).not.toHaveBeenCalledWith('deal');
  });

  it('deal button exposes empty-column reason via title when blocked', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '配る' }).length).toBeGreaterThanOrEqual(1));

    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    const footerDeal = dealButtons[dealButtons.length - 1];
    expect(footerDeal).toHaveAttribute('title', '空の列をすべて埋めないと配れません');
  });

  it('clicking undo button dispatches undo', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('clicking auto complete button dispatches autocomplete when ready', async () => {
    const readyState: SpiderResponse = {
      ...playingState,
      stockCount: 0,
      tableau: makeTableau([[{ card: card('SPADE', 13), faceUp: true }], [], [], [], [], [], [], [], [], []]),
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '自動完成' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('auto complete button is disabled while stock remains or face-down cards exist', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('game clear shows WinCelebration', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームクリア'));
  });

  it('game clear plays winFanfare sound via onCelebrate', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(playSoundMock).toHaveBeenCalledWith('winFanfare'));
  });

  it('game over shows phase text', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('changing difficulty mid-game asks for confirmation before resetting', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByLabelText('難易度')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, difficulty: 2 });
    fireEvent.change(screen.getByLabelText('難易度'), { target: { value: '2' } });
    // Mid-game: no reset until the dialog is confirmed (#2188).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { difficulty: 2 }));
  });

  it('cancelling the difficulty change does not reset', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByLabelText('難易度')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.change(screen.getByLabelText('難易度'), { target: { value: '4' } });
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('changing difficulty after game end resets without confirmation', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByLabelText('難易度')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, difficulty: 2 });
    fireEvent.change(screen.getByLabelText('難易度'), { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { difficulty: 2 }));
  });

  it('renders difficulty selector', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByLabelText('難易度')).toBeInTheDocument());
    expect(screen.getByText('1スート')).toBeInTheDocument();
    expect(screen.getByText('2スート')).toBeInTheDocument();
    expect(screen.getByText('4スート')).toBeInTheDocument();
  });

  it('displays error message', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.spider);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→tableau' }],
    });

    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('clicking tableau face-up card selects as source', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const emptyButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '空');
    for (const btn of emptyButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('pressing d triggers deal in PLAYING phase when no empty columns', async () => {
    const filledTableauState: SpiderResponse = {
      ...playingState,
      tableau: makeTableau(Array.from({ length: 10 }, () => [{ card: card('SPADE', 13), faceUp: true }])),
    };
    mockExec.mockResolvedValue(filledTableauState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(filledTableauState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('pressing h triggers hint in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: { fromCol: 0, cardIndex: 0, toCol: 3 } });
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  // #5955: ヒントは無言で現れていた。**空のまま先にマウントしてある**領域の中身が
  // 変わることが読み上げの条件なので、hint がある間だけ現れる内側の div ではなく、
  // 常設のラッパーがライブ領域でなければならない。
  it('announces the hint through a region that was already mounted', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    const region = screen.getByTestId('spider-hint-live');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('');

    mockExec.mockResolvedValue({ ...playingState, hint: { fromCol: 0, cardIndex: 0, toCol: 3 } });
    fireEvent.keyDown(document, { key: 'h' });
    // **同じ要素**の中身が変わる (別の要素が現れるのではない)。
    await waitFor(() => expect(region).toHaveTextContent(/ヒントがあります/));
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 'h' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameClearState);
    vi.mocked(actionLogApi.spider).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→tableau' }],
    });

    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  it('shows hint text after clicking hint', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 0, toCol: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'プレイ中',
      messageCode: 'spider.playing',
    });
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getAllByText('プレイ中').length).toBeGreaterThanOrEqual(1));
  });

  it('stock card back is clickable during playing phase', async () => {
    const filledTableauState: SpiderResponse = {
      ...playingState,
      tableau: makeTableau(Array.from({ length: 10 }, () => [{ card: card('SPADE', 13), faceUp: true }])),
    };
    mockExec.mockResolvedValue(filledTableauState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(filledTableauState);
    const dealLabels = screen.getAllByLabelText('配る');
    fireEvent.click(dealLabels[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('empty stock shows empty text', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with flex-1 min-w-0 tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="spd-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with responsive tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="spd-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('clicking face-up card in different column after source selected triggers move', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select ♠K in col 0 as source
    const sourceCard = screen.getByAltText('♠ K');
    const sourceButton = sourceCard.closest('button') as HTMLButtonElement;
    fireEvent.click(sourceButton);
    await waitFor(() => expect(sourceButton).toHaveAttribute('aria-pressed', 'true'));

    // Click ♥5 in col 1 (different column) — triggers handleSelectTarget → calls move
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const targetCard = screen.getByAltText('♥ 5');
    const targetButton = targetCard.closest('button') as HTMLButtonElement;
    fireEvent.click(targetButton);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.anything(), expect.anything()));
  });

  it('clicking face-up card in same column after source selected re-selects source', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select ♠K in col 0 as source
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));

    // Click ♠K again (same col 0) — triggers handleSelectSource which deselects
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'false'));
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'deal', reason: 'frontendHint.dealFromStock', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 7, canUndo: true });
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  describe('movable-run hover highlight (#3061)', () => {
    // Col 0 is a valid same-suit descending run (♠K ♠Q ♠J); col 1 breaks at the first
    // card (♣7 then ♥6 — suit mismatch) so ♣7 is not movable while ♥6 rings itself.
    const runState: SpiderResponse = {
      ...playingState,
      tableau: makeTableau([
        [
          { card: card('SPADE', 13), faceUp: true },
          { card: card('SPADE', 12), faceUp: true },
          { card: card('SPADE', 11), faceUp: true },
        ],
        [
          { card: card('CLOVER', 7), faceUp: true },
          { card: card('HEART', 6), faceUp: true },
        ],
        [],
        [],
        [],
        [],
        [],
        [],
        [],
        [],
      ]),
    };

    const btnFor = (alt: string) => screen.getByAltText(alt).closest('button') as HTMLButtonElement;

    it('rings the whole same-suit descending run on hover and clears on leave', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      const top = btnFor('♠ K');
      fireEvent.mouseEnter(top);
      await waitFor(() => expect(btnFor('♠ K')).toHaveAttribute('data-movable-run', 'true'));
      expect(btnFor('♠ Q')).toHaveAttribute('data-movable-run', 'true');
      expect(btnFor('♠ J')).toHaveAttribute('data-movable-run', 'true');
      expect(btnFor('♠ K').className).toContain('ring-ds-success');

      fireEvent.mouseLeave(top);
      await waitFor(() => expect(btnFor('♠ K')).not.toHaveAttribute('data-movable-run'));
    });

    it('rings only the valid suffix when hovering a mid-run card', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ Q')).toBeInTheDocument());

      fireEvent.mouseEnter(btnFor('♠ Q'));
      await waitFor(() => expect(btnFor('♠ Q')).toHaveAttribute('data-movable-run', 'true'));
      expect(btnFor('♠ J')).toHaveAttribute('data-movable-run', 'true');
      // ♠K is above the hovered card, so it is not part of the run.
      expect(btnFor('♠ K')).not.toHaveAttribute('data-movable-run');
    });

    it('does not ring a card whose tail is a broken sequence', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♣ 7')).toBeInTheDocument());

      // Hovering ♣7 would drag ♥6 with it — a broken sequence — so no ring appears.
      fireEvent.mouseEnter(btnFor('♣ 7'));
      await waitFor(() => expect(btnFor('♥ 6')).toBeInTheDocument());
      expect(btnFor('♣ 7')).not.toHaveAttribute('data-movable-run');
      expect(btnFor('♥ 6')).not.toHaveAttribute('data-movable-run');

      // The lone bottom card still rings itself.
      fireEvent.mouseEnter(btnFor('♥ 6'));
      await waitFor(() => expect(btnFor('♥ 6')).toHaveAttribute('data-movable-run', 'true'));
    });

    // **タップにはホバーもフォーカスも無い (#4780)。**クリックで選んだだけの
    // 状態で、一緒に動く連番が見えなければならない。
    it('rings the whole run after a plain tap, with no hover involved', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      fireEvent.click(btnFor('♠ K'));
      await waitFor(() => expect(btnFor('♠ Q')).toHaveAttribute('data-selected-block', 'true'));
      expect(btnFor('♠ J')).toHaveAttribute('data-selected-block', 'true');
      expect(btnFor('♠ Q').className).toContain('ring-ds-info');
      // 選んだ札自身は従来どおり選択リング。
      expect(btnFor('♠ K').className).toContain('ring-ds-warning');
      // 別の列の札は巻き込まない。
      expect(btnFor('♥ 6')).not.toHaveAttribute('data-selected-block');
    });

    // **同じ列でも連番の外は光らせない。**列で絞るだけの実装だと、選んだ札より
    // 上に積まれた札まで「一緒に動く」ように見えてしまう。
    it('leaves cards above the tapped one out of the block', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ Q')).toBeInTheDocument());

      fireEvent.click(btnFor('♠ Q'));
      await waitFor(() => expect(btnFor('♠ J')).toHaveAttribute('data-selected-block', 'true'));
      expect(btnFor('♠ K')).not.toHaveAttribute('data-selected-block');
    });

    // 連番が途切れていれば、タップしても何も光らない (動かせないため)。
    it('marks nothing when the tapped card cannot lift its tail', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♣ 7')).toBeInTheDocument());

      fireEvent.click(btnFor('♣ 7'));
      await waitFor(() => expect(btnFor('♣ 7')).toHaveAttribute('aria-pressed', 'true'));
      expect(btnFor('♣ 7')).not.toHaveAttribute('data-selected-block');
      expect(btnFor('♥ 6')).not.toHaveAttribute('data-selected-block');
    });

    it('marks nothing as a selected block before anything is tapped', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
      expect(document.querySelectorAll('[data-selected-block]')).toHaveLength(0);
    });

    it('highlights the run on keyboard focus too', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      fireEvent.focus(btnFor('♠ K'));
      await waitFor(() => expect(btnFor('♠ K')).toHaveAttribute('data-movable-run', 'true'));
      expect(btnFor('♠ J')).toHaveAttribute('data-movable-run', 'true');

      fireEvent.blur(btnFor('♠ K'));
      await waitFor(() => expect(btnFor('♠ K')).not.toHaveAttribute('data-movable-run'));
    });

    it('keeps the selection ring (warning) when a hovered card is the selected source', async () => {
      mockExec.mockResolvedValue(runState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      fireEvent.click(btnFor('♠ K'));
      await waitFor(() => expect(btnFor('♠ K')).toHaveAttribute('aria-pressed', 'true'));
      fireEvent.mouseEnter(btnFor('♠ K'));
      // Selection ring takes priority over the hover ring on the selected card.
      expect(btnFor('♠ K').className).toContain('ring-ds-warning');
      expect(btnFor('♠ K').className).not.toContain('ring-ds-success');
    });
  });

  describe('drag and drop', () => {
    function buildDataTransfer() {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    }

    it('tableau face-up card is draggable when playing', async () => {
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
      const cardButton = screen.getByAltText('♠ K').closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging a tableau card to another column dispatches move', async () => {
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      const sourceButton = screen.getByAltText('♠ K').closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();
      fireEvent.dragStart(sourceButton, { dataTransfer });

      // Find an empty column via the empty placeholder
      const emptyButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'なし');
      if (emptyButtons.length === 0) {
        // Skip if no empty column in this layout
        return;
      }
      const dropZone = emptyButtons[0].closest('div');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'tableau' }),
          expect.objectContaining({ zone: 'tableau' }),
        ),
      );
    });
  });

  describe('per-difficulty stats (#3062)', () => {
    it('renders the stats panel with a zeroed win rate initially', async () => {
      renderWithProviders(<SpiderPage />);
      const panel = await screen.findByTestId('spd-stats-panel');
      expect(panel).toHaveTextContent('勝率 0% (0/0)');
    });

    it('records a win to localStorage and shows the best badge', async () => {
      mockExec.mockResolvedValue(gameClearState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => {
        const stored = JSON.parse(localStorage.getItem('trumpcards-spider-stats') ?? '{}');
        expect(stored['1']).toEqual({ plays: 1, wins: 1, bestScore: 500, fewestMoves: 5 });
      });
      // Win rate reflects the recorded game and the personal-best badge appears.
      expect(await screen.findByTestId('spd-stats-panel')).toHaveTextContent('勝率 100% (1/1)');
      expect(screen.getByTestId('spd-best-badge')).toBeInTheDocument();
    });

    it('records a loss (plays only) and shows no best badge', async () => {
      mockExec.mockResolvedValue(gameOverState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => {
        const stored = JSON.parse(localStorage.getItem('trumpcards-spider-stats') ?? '{}');
        expect(stored['1']).toEqual({ plays: 1, wins: 0, bestScore: null, fewestMoves: null });
      });
      expect(await screen.findByTestId('spd-stats-panel')).toHaveTextContent('勝率 0% (0/1)');
      expect(screen.queryByTestId('spd-best-badge')).not.toBeInTheDocument();
    });

    it('reads back previously stored stats for the current difficulty', async () => {
      localStorage.setItem(
        'trumpcards-spider-stats',
        JSON.stringify({ '1': { plays: 4, wins: 3, bestScore: 620, fewestMoves: 33 } }),
      );
      renderWithProviders(<SpiderPage />);
      const panel = await screen.findByTestId('spd-stats-panel');
      expect(panel).toHaveTextContent('勝率 75% (3/4)');
      expect(panel).toHaveTextContent('ベスト 620');
      expect(panel).toHaveTextContent('最少 33手');
    });

    it('keeps difficulties separate (difficulty 2 stats hidden while on difficulty 1)', async () => {
      localStorage.setItem(
        'trumpcards-spider-stats',
        JSON.stringify({ '2': { plays: 5, wins: 5, bestScore: 900, fewestMoves: 10 } }),
      );
      renderWithProviders(<SpiderPage />);
      const panel = await screen.findByTestId('spd-stats-panel');
      // Playing state is difficulty 1, so difficulty 2's best must not leak in.
      expect(panel).toHaveTextContent('勝率 0% (0/0)');
      expect(panel).not.toHaveTextContent('ベスト 900');
    });
  });

  it('announces why the deal was refused, not just a shake', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('spd-empty-col-2')).toBeInTheDocument());
    // Nothing to announce before the player tries.
    expect(screen.queryByTestId('spd-deal-refusal')).not.toBeInTheDocument();

    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);

    // The shake and the title attribute reach neither a screen reader nor a
    // keyboard user; the button is not even disabled, so without this the
    // refusal is silent.
    const live = await screen.findByTestId('spd-deal-refusal');
    expect(live).toHaveTextContent('空の列');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveClass('sr-only');
  });
});
