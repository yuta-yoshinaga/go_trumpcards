import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { russianSolitaireApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, RussianSolitaireResponse } from '../types/card';
import { RussianSolitairePage } from './RussianSolitairePage';

/**
 * This page's own hint region.
 *
 * **`GameMessageBox` is also `role="status"`**, and it now renders on every
 * phase because this game's messageCodes are translated (#5291). Querying the
 * role alone therefore matches two elements; the message box is the one built
 * from `glass-panel`, so the hint region is the other one.
 */
const hintLiveRegion = () =>
  screen.queryAllByRole('status').find((el) => !el.classList.contains('glass-panel')) ?? null;

vi.mock('../api/gameApi', () => ({
  russianSolitaireApi: { exec: vi.fn() },
  actionLogApi: { russiansolitaire: vi.fn() },
}));

const mockExec = vi.mocked(russianSolitaireApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: RussianSolitaireResponse = {
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
  messageCode: 'russiansolitaire.playing',
};

const gameClearState: RussianSolitaireResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'russiansolitaire.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: RussianSolitaireResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'russiansolitaire.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  // Clear the persisted rule-dismissed flag so it never leaks across tests
  // (a dismissed flag left in localStorage would hide the banner in later tests).
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('RussianSolitairePage', () => {
  it('renders heading', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
  });

  it('shows the face-down rule note and gives face-down cards concise positional labels', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('rs-facedown-rule')).toBeInTheDocument());
    // **Yukon との唯一の違いを伝える (#4789)。**裏向きは動かせないという汎用ルール
    // だけでは、交互色で積もうとして詰まるプレイヤーを救えない。
    expect(screen.getByTestId('rs-facedown-rule')).toHaveTextContent('同じスートで1つ下');
    expect(screen.getByTestId('rs-facedown-rule')).toHaveTextContent('交互色では繋げません');
    // The rule text lives only in the note, not repeated on every card label.
    expect(screen.queryByLabelText(/移動できません/)).not.toBeInTheDocument();
    // Each face-down card exposes a short label with its column and position.
    // Columns use the same 0-based index shown in the visible column header:
    // column 1 has one face-down card at the top; column 2 has two.
    expect(screen.getByLabelText('列1、上から1枚目、裏向き')).toBeInTheDocument();
    expect(screen.getByLabelText('列2、上から1枚目、裏向き')).toBeInTheDocument();
    expect(screen.getByLabelText('列2、上から2枚目、裏向き')).toBeInTheDocument();
  });

  it('dismisses the face-down rule note, persists it, and keeps it hidden on remount', async () => {
    const { unmount } = renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('rs-facedown-rule')).toBeInTheDocument());

    // Clicking the close control hides the note and persists the dismissal.
    fireEvent.click(screen.getByRole('button', { name: 'ルール説明を閉じる' }));
    expect(screen.queryByTestId('rs-facedown-rule')).not.toBeInTheDocument();
    expect(localStorage.getItem('russiansolitaire-rules-dismissed')).toBe('true');

    // On a fresh remount the persisted flag keeps the banner hidden.
    unmount();
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('rs-facedown-rule')).not.toBeInTheDocument();
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('embeds the hint in card aria-labels instead of a text panel', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 1, cardIndex: 1, toZone: 'tableau', toCol: 4 },
      messageCode: 'russiansolitaire.hintAvailable',
    });
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The source (♥8) and target (♠3) cards carry the hint in their aria-labels.
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ヒント: このカードを場札 4へ移動/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /ヒント: 移動先/ })).toBeInTheDocument();
    // The hint live region survives for screen readers but is visually hidden,
    // so it no longer squeezes the footer on mobile. It names the card since
    // "this card" is ambiguous when announced without focus context.
    const liveRegion = hintLiveRegion();
    expect(liveRegion).toHaveClass('sr-only');
    expect(liveRegion).toHaveTextContent('♥ 8');
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('hides the hint when it was not requested', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 1, cardIndex: 1, toZone: 'tableau', toCol: 4 },
    });
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The source (♥8) and target (♠3) cards carry the hint in their aria-labels.
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: /ヒント: このカードを場札 4へ移動/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /ヒント: 移動先/ })).not.toBeInTheDocument();
    // The hint live region survives for screen readers but is visually hidden,
    // so it no longer squeezes the footer on mobile. It names the card since
    // "this card" is ambiguous when announced without focus context.
    // 頼んでいないので live region ごと出ない。
    expect(hintLiveRegion()).not.toBeInTheDocument();
  });

  it('labels the hint source with the foundation destination', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
      messageCode: 'russiansolitaire.hintAvailable',
    });
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ヒント: このカードを組札へ移動/ })).toBeInTheDocument(),
    );
  });

  it('autocomplete button triggers autocomplete command when all face-up', async () => {
    const readyState: RussianSolitaireResponse = {
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
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('autocomplete button is disabled while face-down cards exist', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<RussianSolitairePage />);
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
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('renders foundation suit labels', async () => {
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('RussianSolitairePage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RussianSolitairePage />);
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
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

describe('RussianSolitairePage deselect routing (#4439)', () => {
  // Clicking a selected card that is ALSO the last card in its column must
  // deselect it. It used to be routed to handleSelectTarget instead, dispatching a
  // move onto its own column, which the server rejects — so a player trying to
  // deselect got a rejection message. Scorpion had this same bug and had a test
  // for it, but the assertion ran before react-query's microtask could deliver the
  // call, so it passed regardless. These two pages had no such test at all.
  // The other branch of the same condition: a DIFFERENT column's last card, while
  // something is selected, is a move target. Neither of these two pages had any
  // test covering a move dispatch at all before this, so the fix above changed a
  // line that nothing exercised.
  it("clicking another column's last card while a card is selected dispatches the move", async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RussianSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /\u2665 8/ }));
    fireEvent.click(screen.getByRole('button', { name: /\u2663 5/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 1, cardIndex: 1 },
        { zone: 'tableau', col: 2 },
      ),
    );
  });

  it('clicking the same selected card deselects it instead of moving onto itself', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<RussianSolitairePage />);
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
