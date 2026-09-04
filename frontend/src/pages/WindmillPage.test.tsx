import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { windmillApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, WindmillResponse } from '../types/card';
import { WindmillPage } from './WindmillPage';

vi.mock('../api/gameApi', () => ({
  windmillApi: { exec: vi.fn() },
  actionLogApi: { windmill: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(windmillApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeSails(cards: (Card | null)[]): (Card | null)[] {
  return Array.from({ length: 8 }, (_, i) => cards[i] ?? null);
}

const playingState: WindmillResponse = {
  sails: makeSails([card('SPADE', 9), card('HEART', 4)]),
  center: [card('CLOVER', 1)],
  corners: [[card('DIAMOND', 13)], [], [], []],
  stockCount: 100,
  waste: [],
  transferBlocked: false,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: WindmillResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'windmill.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: WindmillResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'windmill.gameOver',
};

describe('WindmillPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByText(/ウィンドミル/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight sails and four corners', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    // Six sails are empty in this fixture, and three corners are unopened.
    await waitFor(() => expect(screen.getAllByLabelText(/帆 \d は空です/).length).toBe(6));
    expect(screen.getAllByLabelText(/空の四隅基礎札\d/).length).toBe(3);
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  // The centre runs A-K four times, so its progress readout is 52, not 13.
  it('shows the centre progress against 52 cards', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByText('1/52')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '中央基礎札 1/52枚' })).toBeInTheDocument();
  });

  it('labels an unopened corner as Kings-only', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空の四隅基礎札1 (K のみ置けます)' })).toBeInTheDocument(),
    );
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り100枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('sends a sail card to the centre', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    const sail = await screen.findByRole('button', { name: /: ♠ 9$/ });
    fireEvent.click(sail);
    await waitFor(() => expect(sail).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '中央基礎札 1/52枚' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'sail', col: 0 }, { zone: 'center' }));
  });

  it('sends a sail card to an unopened corner', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    const sail = await screen.findByRole('button', { name: /: ♥ 4$/ });
    fireEvent.click(sail);
    await waitFor(() => expect(sail).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '空の四隅基礎札2 (K のみ置けます)' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'sail', col: 1 }, { zone: 'corner', col: 2 }),
    );
  });

  // An occupied corner is both a target for a descending card and the source of
  // the pull-back, so clicking it does different things by selection state.
  it('pulls a corner card back onto the centre', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    const corner = await screen.findByRole('button', { name: '四隅基礎札0 1/13枚' });
    fireEvent.click(corner);
    await waitFor(() => expect(corner).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '中央基礎札 1/52枚' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'corner', col: 0 }, { zone: 'center' }));
  });

  it('selects the waste top and moves it', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('DIAMOND', 2)] });
    renderWithProviders(<WindmillPage />);
    const wasteTop = await screen.findByRole('button', { name: '♦ 2' });
    fireEvent.click(wasteTop);
    await waitFor(() => expect(wasteTop).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '中央基礎札 1/52枚' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'center' }));
  });

  it('shows an empty waste slot when nothing has been turned', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  // The restriction has no visual counterpart on the board, so it is spelled out.
  it('announces the transfer block', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    expect(screen.queryByText(/次に中央へ置く札は/)).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue({ ...playingState, transferBlocked: true });
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByText(/次に中央へ置く札は/)).toBeInTheDocument());
  });

  // **移動先としては塞がない。**四隅は置き場と引き戻し元を兼ねていて、
  // transferBlocked が禁じるのは引き戻しだけ。ボタンごと無効にすると、
  // 影響を受けないはずの「四隅へ置く」手まで潰れる。
  describe('the corner while the transfer is blocked', () => {
    const cornerName = /四隅基礎札0/;

    it('cannot be picked as a source', async () => {
      mockExec.mockResolvedValue({ ...playingState, transferBlocked: true });
      renderWithProviders(<WindmillPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: cornerName })).toBeInTheDocument());

      const corner = screen.getByRole('button', { name: cornerName });
      expect(corner).toBeDisabled();
      expect(corner).toHaveAttribute('title', expect.stringContaining('引き戻した直後'));
    });

    it('is still usable as a target once something is selected', async () => {
      mockExec.mockResolvedValue({ ...playingState, transferBlocked: true });
      renderWithProviders(<WindmillPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: cornerName })).toBeInTheDocument());

      // 帆の札を選ぶと、四隅は「置き先」になるので再び押せる。
      // 帆のボタンの aria-label は sailAriaLabel ("帆 N: <札>") なので、札名は末尾に来る。
      fireEvent.click(screen.getByRole('button', { name: /: ♠ 9$/ }));

      await waitFor(() => expect(screen.getByRole('button', { name: cornerName })).toBeEnabled());
    });

    it('stays selectable as a source when nothing is blocked', async () => {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<WindmillPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: cornerName })).toBeEnabled());
      expect(screen.getByRole('button', { name: cornerName })).not.toHaveAttribute('title');
    });
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);

    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  // Two decks: the summary counts the centre and all four corners against 104.
  it('counts the summary across the centre and the corners', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WindmillPage />);
    const summary = await screen.findByTestId('wm-gameover-summary');
    expect(summary).toHaveTextContent('2/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('wm-gameover-summary')).not.toBeInTheDocument();
  });

  // Auto-complete only moves sail and waste cards, so an empty cross with an
  // empty waste leaves it nothing to do.
  it('disables auto-complete when neither a sail nor the waste holds a card', async () => {
    mockExec.mockResolvedValue({ ...playingState, sails: makeSails([]) });
    const { unmount } = renderWithProviders(<WindmillPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({ ...playingState, sails: makeSails([]), waste: [card('SPADE', 2)] });
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['center', { fromZone: 'sail', fromIdx: 1, toZone: 'center', toIdx: -1 }, '中央基礎札'],
    ['corner', { fromZone: 'waste', fromIdx: -1, toZone: 'corner', toIdx: 2 }, '四隅基礎札2'],
    ['pull-back', { fromZone: 'corner', fromIdx: 0, toZone: 'center', toIdx: -1 }, '四隅基礎札0'],
    ['draw', { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });

  it('names each sail card with its position for screen readers', async () => {
    // Earlier tests in this file queue one-shot resolutions and can leave CLI
    // mode persisted in localStorage; reset both so the board actually renders.
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/帆 \d+: /).length).toBeGreaterThan(0));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('WindmillPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WindmillPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

describe('WindmillPage progress tracking', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('shows foundation progress during gameplay and tracks identically to game over', async () => {
    // 2/104 = 1.92% => Math.round = 2%
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<WindmillPage />);
    const progress = await screen.findByTestId('wm-foundation-progress');
    expect(progress).toHaveTextContent('収納: 2/104 (2%)');
    unmount();

    // 否定コントロール: 0枚の場合
    const zeroState = { ...playingState, center: [], corners: [[], [], [], []], stockCount: 102 };
    mockExec.mockResolvedValue(zeroState);
    renderWithProviders(<WindmillPage />);
    const progressZero = await screen.findByTestId('wm-foundation-progress');
    expect(progressZero).toHaveTextContent('収納: 0/104 (0%)');
  });
});
