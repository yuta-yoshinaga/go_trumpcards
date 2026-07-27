import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { yukonApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, YukonResponse } from '../types/card';
import { YukonPage } from './YukonPage';

vi.mock('../api/gameApi', () => ({
  yukonApi: { exec: vi.fn() },
  actionLogApi: { yukon: vi.fn() },
}));

const mockExec = vi.mocked(yukonApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: YukonResponse = {
  tableau: [
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 8), faceUp: true },
    ],
    [
      { card: null, faceUp: false },
      { card: null, faceUp: false },
      { card: card('CLOVER', 5), faceUp: true },
    ],
    [{ card: card('DIAMOND', 10), faceUp: true }],
    [{ card: card('SPADE', 3), faceUp: true }],
    [{ card: card('HEART', 7), faceUp: true }],
    [{ card: card('CLOVER', 2), faceUp: true }],
  ],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'yukon.playing',
};

const gameClearState: YukonResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'yukon.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: YukonResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'yukon.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('YukonPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('embeds the hint in card aria-labels instead of a text panel', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 1, cardIndex: 1, toZone: 'tableau', toCol: 4 },
    });
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The source (♥8) and target (♠3) cards carry the hint in their aria-labels.
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ヒント: このカードを場札 4へ移動/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /ヒント: 移動先/ })).toBeInTheDocument();
    // The hint live region survives for screen readers but is visually hidden,
    // so it no longer squeezes the footer on mobile. It names the card since
    // "this card" is ambiguous when announced without focus context.
    const liveRegion = screen.getByRole('status');
    expect(liveRegion).toHaveClass('sr-only');
    expect(liveRegion).toHaveTextContent('♥ 8');
  });

  it('labels the hint source with the foundation destination', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    });
    renderWithProviders(<YukonPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ヒント: このカードを組札へ移動/ })).toBeInTheDocument(),
    );
  });

  it('autocomplete button triggers autocomplete command when all face-up', async () => {
    const readyState: YukonResponse = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [{ card: card('HEART', 8), faceUp: true }],
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('autocomplete button is disabled while face-down cards exist', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    expect(undoBtn).toBeDisabled();
  });

  it('renders empty tableau column placeholder', async () => {
    const stateWithEmpty = {
      ...playingState,
      tableau: [[], ...playingState.tableau.slice(1)],
    };
    mockExec.mockResolvedValue(stateWithEmpty);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('renders foundation suit labels', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('announces foundations with localized suit names, not bare glyphs', async () => {
    // Fill the spade foundation so the filled-foundation label path is exercised too.
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 1)], [], [], []],
    });
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Filled spade foundation reads "スペード 組札 1枚"; empty ones read "空の組札 (<name>)".
    expect(screen.getByRole('button', { name: 'スペード 組札 1枚' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (クラブ)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (ハート)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (ダイヤ)' })).toBeInTheDocument();
  });

  describe('block move preview', () => {
    // Column 0 has three face-up cards stacked (♠K on top, then ♥Q, then ♣J).
    // Selecting a mid-column card should preview the whole block that lifts with it.
    const blockState: YukonResponse = {
      ...playingState,
      tableau: [
        [
          { card: card('SPADE', 13), faceUp: true },
          { card: card('HEART', 12), faceUp: true },
          { card: card('CLOVER', 11), faceUp: true },
        ],
        [{ card: card('DIAMOND', 9), faceUp: true }],
        [],
        [],
        [],
        [],
        [],
      ],
    };

    it('highlights the selected card and every card below it as a block', async () => {
      mockExec.mockResolvedValue(blockState);
      renderWithProviders(<YukonPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const middle = screen.getByRole('button', { name: '♥ Q' });
      fireEvent.click(middle);

      // Selected card carries the selection ring; the card below is a block member.
      expect(middle).toHaveAttribute('data-selected-block');
      expect(screen.getByRole('button', { name: '♣ J' })).toHaveAttribute('data-selected-block');
      // The card above the selection is NOT part of the block.
      expect(screen.getByRole('button', { name: '♠ K' })).not.toHaveAttribute('data-selected-block');
    });

    it('clears the block highlight when the selection is toggled off', async () => {
      mockExec.mockResolvedValue(blockState);
      renderWithProviders(<YukonPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      const middle = screen.getByRole('button', { name: '♥ Q' });
      fireEvent.click(middle); // select
      expect(screen.getByRole('button', { name: '♣ J' })).toHaveAttribute('data-selected-block');

      fireEvent.click(middle); // deselect (same card toggles off)
      expect(screen.getByRole('button', { name: '♣ J' })).not.toHaveAttribute('data-selected-block');
      expect(screen.getByRole('button', { name: '♥ Q' })).not.toHaveAttribute('data-selected-block');
    });
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('YukonPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<YukonPage />);
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
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

describe('YukonPage deselect routing (#4439)', () => {
  // Clicking a selected card that is ALSO the last card in its column must
  // deselect it. It used to be routed to handleSelectTarget instead, dispatching a
  // move onto its own column, which the server rejects — so a player trying to
  // deselect got a rejection message. Scorpion had this same bug and had a test
  // for it, but the assertion ran before react-query's microtask could deliver the
  // call, so it passed regardless. These two pages had no such test at all.
  it('clicking the same selected card deselects it instead of moving onto itself', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const heart8 = screen.getByRole('button', { name: /♥ 8/ });
    fireEvent.click(heart8);
    await waitFor(() => expect(heart8.className).toMatch(/ring-/));
    fireEvent.click(heart8);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
