import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { russianSolitaireApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, RussianSolitaireResponse } from '../types/card';
import { RussianSolitairePage } from './RussianSolitairePage';

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
    const liveRegion = screen.getByRole('status');
    expect(liveRegion).toHaveClass('sr-only');
    expect(liveRegion).toHaveTextContent('♥ 8');
  });

  it('labels the hint source with the foundation destination', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
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
    expect(mockExec).not.toHaveBeenCalled();
  });
});
