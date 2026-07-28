import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, crescentApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CrescentResponse, CrescentTableauCard } from '../types/card';
import { CrescentPage } from './CrescentPage';

vi.mock('../api/gameApi', () => ({
  crescentApi: { exec: vi.fn() },
  actionLogApi: { crescent: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(crescentApi.exec);

function makeTableau(cols: CrescentTableauCard[][]): CrescentTableauCard[][] {
  const result: CrescentTableauCard[][] = [];
  for (let i = 0; i < 16; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: CrescentResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 5), faceUp: true },
      { card: card('SPADE', 4), faceUp: true },
    ],
    [{ card: card('HEART', 6), faceUp: true }],
  ]),
  foundation: [
    [card('SPADE', 1)],
    [card('CLOVER', 1)],
    [card('HEART', 1)],
    [card('DIAMOND', 1)],
    [card('SPADE', 13)],
    [card('CLOVER', 13)],
    [card('HEART', 13)],
    [card('DIAMOND', 13)],
  ],
  redealsRemaining: 3,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: CrescentResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'crescent.gameClear',
};

const gameOverState: CrescentResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'crescent.gameOver',
};

const noRedealsState: CrescentResponse = {
  ...playingState,
  redealsRemaining: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('CrescentPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CrescentPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders redeals remaining in header', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByText(/残り再配り回数: 3/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders ascending and descending foundation suit headers', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getAllByText(/♠ ↑/).length).toBeGreaterThanOrEqual(1));
    expect(screen.getAllByText(/♣ ↑/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/♥ ↑/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/♦ ↑/).length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText(/♠ ↓/).length).toBeGreaterThanOrEqual(1);
  });

  it('color-codes the foundation direction badges and names the top card', async () => {
    renderWithProviders(<CrescentPage />);
    const asc = await screen.findByTestId('foundation-dir-0'); // row 0 = ascending
    expect(asc.className).toContain('text-ds-success');
    const desc = screen.getByTestId('foundation-dir-4'); // row 1 = descending
    expect(desc.className).toContain('text-ds-warning');
    // Ascending ♠ pile tops out at A → aria-label is localized and names the top card.
    expect(screen.getByLabelText(/昇順ファンデーション ♠ 残り1枚 トップ ♠ A/)).toBeInTheDocument();
  });

  it('redeal button shows remaining count', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再配り/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /再配り \(3\)/ })).toBeInTheDocument();
  });

  it('exposes the redeal count via an aria-live status region', async () => {
    renderWithProviders(<CrescentPage />);
    const status = await screen.findByText(/残り再配り回数/);
    expect(status).toHaveAttribute('aria-live', 'polite');
  });

  it('announces a stalemate via a role=alert region', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2 });
    renderWithProviders(<CrescentPage />);
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toMatch(/手詰まり/);
    expect(alert.textContent).toMatch(/2/);
  });

  it('defaults the escape count to 0 when undoToEscape is absent', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: undefined });
    renderWithProviders(<CrescentPage />);
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toMatch(/手詰まり/);
    expect(alert.textContent).toMatch(/0/);
  });

  it('clicking redeal dispatches redeal', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再配り \(3\)/ })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: /再配り \(3\)/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('redeal button disabled when no redeals remain', async () => {
    mockExec.mockResolvedValue(noRedealsState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再配り \(0\)/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /再配り \(0\)/ })).toBeDisabled();
  });

  it('clicking hint dispatches hint', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<CrescentPage />);
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

  it('clicking tableau top card selects it as source', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const cardImg = screen.getByAltText('♠ 4');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('selecting source then foundation dispatches move', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select top card of col 0 (♠ 4)
    const cardImg = screen.getByAltText('♠ 4');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click a foundation (♠ A ascending)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const foundationImg = screen.getByAltText('♠ A');
    const foundationButton = foundationImg.closest('button') as HTMLButtonElement;
    fireEvent.click(foundationButton);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        expect.objectContaining({ zone: 'tableau', col: 0 }),
        expect.objectContaining({ zone: 'foundation', col: 0 }),
      ),
    );
  });

  it('rings valid destinations when a source is selected (no hover needed)', async () => {
    // col0 top ♠4 can go onto foundation ♠3 (asc) and onto tableau ♠5; ♥6 is invalid.
    const highlightState: CrescentResponse = {
      ...playingState,
      tableau: makeTableau([
        [{ card: card('SPADE', 4), faceUp: true }],
        [{ card: card('SPADE', 5), faceUp: true }],
        [{ card: card('HEART', 6), faceUp: true }],
      ]),
      foundation: [
        [card('SPADE', 3)],
        [card('CLOVER', 1)],
        [card('HEART', 1)],
        [card('DIAMOND', 1)],
        [card('SPADE', 13)],
        [card('CLOVER', 13)],
        [card('HEART', 13)],
        [card('DIAMOND', 13)],
      ],
    };
    mockExec.mockResolvedValue(highlightState);

    const { container } = renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const foundations = container.querySelector('[data-tutorial="crescent-foundations"]') as HTMLElement;
    const tableau = container.querySelector('[data-tutorial="crescent-tableau"]') as HTMLElement;
    // No selection yet -> nothing highlighted.
    expect(foundations.querySelectorAll('.ring-ds-success')).toHaveLength(0);
    expect(tableau.querySelectorAll('.ring-ds-success')).toHaveLength(0);

    // Select ♠4 as the source.
    const cardButton = screen.getByAltText('♠ 4').closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);

    await waitFor(() => expect(foundations.querySelectorAll('.ring-ds-success').length).toBe(1));
    // Exactly the ♠5 column rings; the ♥6 column and the source do not.
    expect(tableau.querySelectorAll('.ring-ds-success')).toHaveLength(1);
    const invalidTop = screen.getByAltText('♥ 6').closest('button') as HTMLButtonElement;
    expect(invalidTop.className).toContain('opacity-40');
  });

  it('clicking reset dispatches reset (after confirm)', async () => {
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over hides play buttons', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('renders empty foundation placeholder', async () => {
    const emptyFndState: CrescentResponse = {
      ...playingState,
      foundation: [[], [], [], [], [], [], [], []],
    };
    mockExec.mockResolvedValue(emptyFndState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getAllByText(/♠ ↑/).length).toBeGreaterThanOrEqual(1));
    // Empty asc foundations show "A"; descending show "K".
    expect(screen.getAllByText('A').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('K').length).toBeGreaterThanOrEqual(1);
  });

  it('action log fetches and shows log entries', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.crescent);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'redeal', detail: 'shuffle' }],
    });

    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('assigns the crescent arc offsets symmetrically across the 16 tableau columns (desktop)', async () => {
    // Force a desktop viewport so the arc is not zeroed out by the mobile guard.
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<CrescentPage />);
      await waitFor(() => expect(document.querySelectorAll('[data-crescent-arc]').length).toBeGreaterThanOrEqual(16));
      const arcs = document.querySelectorAll('[data-crescent-arc]');
      // Top half curves up (negative), bottom half mirrors it down (positive).
      expect(arcs[0]).toHaveAttribute('data-crescent-arc', '-21'); // col 0
      expect(arcs[3]).toHaveAttribute('data-crescent-arc', '-3'); // col 3 (closest to center)
      expect(arcs[7]).toHaveAttribute('data-crescent-arc', '-21'); // col 7 (symmetric to col 0)
      expect(arcs[8]).toHaveAttribute('data-crescent-arc', '21'); // col 8, mirrored
      expect(arcs[15]).toHaveAttribute('data-crescent-arc', '21'); // col 15
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('renders a [0]..[15] column-number badge above each tableau pile', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('crescent-col-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('crescent-col-badge-0')).toHaveTextContent('[0]');
    expect(screen.getByTestId('crescent-col-badge-15')).toHaveTextContent('[15]');
    // The badge is decorative — it must not add to the SR card label noise.
    expect(screen.getByTestId('crescent-col-badge-7')).toHaveAttribute('aria-hidden', 'true');
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('CrescentPage keyboard shortcuts', () => {
  it.each([
    ['d', 'redeal'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    // give-up is irreversible, so the key must route through the dialog (#2099)
    // instead of dispatching straight away.
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CrescentPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
