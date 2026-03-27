import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, spiderApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpiderResponse, SpiderTableauCard } from '../types/card';
import { SpiderPage } from './SpiderPage';

vi.mock('../api/gameApi', () => ({
  spiderApi: { exec: vi.fn() },
  actionLogApi: { spider: vi.fn() },
}));

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
  mockExec.mockResolvedValue(playingState);
});

describe('SpiderPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpiderPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
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

  it('renders tableau columns', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());
    // Column labels 0-9
    for (let i = 0; i < 10; i++) {
      expect(screen.getByText(i.toString())).toBeInTheDocument();
    }
  });

  it('clicking deal button dispatches deal', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '配る' }).length).toBeGreaterThanOrEqual(1));

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // The footer button is the last one
    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
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

  it('clicking auto complete button dispatches autocomplete', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '自動完成' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('clicking give up button dispatches giveup', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('game clear shows WinCelebration', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームクリア'));
  });

  it('game over shows phase text', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
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

  it('changing difficulty resets game with config', async () => {
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
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    const emptyButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '空');
    for (const btn of emptyButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('pressing d triggers deal in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
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

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 'h' });
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
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

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
    renderWithProviders(<SpiderPage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
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

  it('renders mobile viewport with flex-shrink-0 and fixed-width tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<SpiderPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="spd-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-shrink-0');
      expect(firstCol?.className).toContain('sm:flex-1');
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
      expect(firstCol?.className).toContain('flex-shrink-0');
      expect(firstCol?.className).toContain('sm:flex-1');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });
});
