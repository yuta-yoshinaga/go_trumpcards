import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { coloradoApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ColoradoResponse } from '../types/card';
import { ColoradoPage, coloradoNextRank } from './ColoradoPage';

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
  coloradoApi: { exec: vi.fn() },
  actionLogApi: { colorado: vi.fn() },
}));

const mockExec = vi.mocked(coloradoApi.exec);
const card = (design: CardDesign, value: number): Card => ({ design, value });

const TABLEAU_CNT = 20;

/** 20 piles of one card, with pile 3 emptied so the gap paths are reachable. */
function tableau(): Card[][] {
  const piles = Array.from({ length: TABLEAU_CNT }, (_, i) => [card('SPADE', ((i % 13) + 1) as number)]);
  piles[3] = [];
  return piles;
}

function makeState(overrides: Partial<ColoradoResponse> = {}): ColoradoResponse {
  return {
    tableau: tableau(),
    // F0 spades ascending holds the Ace; F4 spades descending holds the King.
    foundation: [[card('SPADE', 1)], [], [], [], [card('SPADE', 13)], [], [], []],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    waste: [card('HEART', 6)],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    messageCode: 'colorado.playing',
    ...overrides,
  };
}

describe('coloradoNextRank', () => {
  it('opens an ascending foundation at the Ace and a descending one at the King', () => {
    expect(coloradoNextRank(undefined, 0, true)).toBe(1);
    expect(coloradoNextRank(undefined, 0, false)).toBe(13);
  });

  it('steps up or down depending on the direction', () => {
    expect(coloradoNextRank(1, 1, true)).toBe(2);
    expect(coloradoNextRank(13, 1, false)).toBe(12);
  });

  it('returns null once the pile is complete', () => {
    expect(coloradoNextRank(13, 13, true)).toBeNull();
    expect(coloradoNextRank(1, 13, false)).toBeNull();
  });
});

describe('ColoradoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('includes the top card in non-empty tableau names without repeating the suit name', async () => {
    renderWithProviders(<ColoradoPage />);
    const pile = await screen.findByTestId('co-tableau-0');
    expect(pile).toHaveAccessibleName('場札 0、1枚、一番上のカード ♠ A');
    expect(pile.getAttribute('aria-label')).not.toContain('スペード');
  });

  it('keeps empty tableau names explicit', async () => {
    renderWithProviders(<ColoradoPage />);
    const pile = await screen.findByTestId('co-tableau-3');
    expect(pile).toHaveAccessibleName('場札 3 空き山。山札か捨て札から埋められます');
  });

  it('keeps a move status region in the DOM before the board loads', () => {
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<ColoradoPage />);
    expect(screen.getAllByRole('status').length).toBeGreaterThan(0);
  });

  it.each([
    ['waste to foundation', { zone: 'waste' }, { zone: 'foundation' }, '捨て札から組札へ移動しました'],
    ['tableau to foundation', { zone: 'tableau', idx: 2 }, { zone: 'foundation' }, '場札 2から組札へ移動しました'],
    ['waste to tableau', { zone: 'waste' }, { zone: 'tableau', idx: 7 }, '捨て札から場札 7へ移動しました'],
    ['stock to tableau', { zone: 'stock' }, { zone: 'tableau', idx: 3 }, '山札から場札 3へ移動しました'],
  ] as const)('announces successful %s', async (_name, from, to, announcement) => {
    mockExec.mockResolvedValue(makeState({ moveCount: 1 }));
    renderWithProviders(<ColoradoPage />);
    await screen.findByTestId('co-tableau-0');
    const sourceButton =
      from.zone === 'waste'
        ? screen.getByTestId('co-waste-button')
        : from.zone === 'stock'
          ? screen.getByTestId('co-stock-fill-button')
          : screen.getByTestId(`co-tableau-${from.idx}`);
    fireEvent.click(sourceButton);
    const destinationButton =
      to.zone === 'foundation' ? screen.getByTestId('co-foundation-0') : screen.getByTestId(`co-tableau-${to.idx}`);
    fireEvent.click(destinationButton);
    await waitFor(() =>
      expect(screen.getAllByRole('status').some((region) => region.textContent?.includes(announcement))).toBe(true),
    );
  });

  it('replaces the live announcement node when the same move succeeds twice', async () => {
    mockExec.mockResolvedValue(makeState({ moveCount: 1 }));
    renderWithProviders(<ColoradoPage />);
    await screen.findByTestId('co-tableau-0');

    const liveRegion = screen.getByTestId('colorado-hint-live');
    const moveWasteToFoundation = () => {
      fireEvent.click(screen.getByTestId('co-waste-button'));
      fireEvent.click(screen.getByTestId('co-foundation-0'));
    };

    moveWasteToFoundation();
    await waitFor(() => expect(liveRegion.textContent).toContain('捨て札から組札へ移動しました'));
    const firstAnnouncement = liveRegion.querySelector('span');

    moveWasteToFoundation();
    await waitFor(() => expect(liveRegion.querySelector('span')).not.toBe(firstAnnouncement));

    expect(liveRegion.querySelector('span')).not.toBe(firstAnnouncement);
  });

  it('does not announce a rejected move or failed request', async () => {
    mockExec.mockResolvedValue(makeState({ message: 'illegal move', messageCode: '' }));
    renderWithProviders(<ColoradoPage />);
    await screen.findByTestId('co-tableau-0');
    fireEvent.click(screen.getByTestId('co-waste-button'));
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await flushPendingDispatch();
    expect(screen.getAllByRole('status').every((region) => !region.textContent?.includes('移動しました'))).toBe(true);
  });

  it('does not announce a move when the request fails', async () => {
    mockExec.mockResolvedValueOnce(makeState());
    mockExec.mockRejectedValueOnce(new Error('network error'));
    renderWithProviders(<ColoradoPage />);
    await screen.findByTestId('co-tableau-0');
    fireEvent.click(screen.getByTestId('co-waste-button'));
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await screen.findByRole('alert');
    expect(screen.queryByText(/移動しました/)).not.toBeInTheDocument();
  });

  it('announces a successful move in English', async () => {
    const originalLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      mockExec.mockResolvedValue(makeState({ moveCount: 1 }));
      renderWithProviders(<ColoradoPage />);
      await screen.findByTestId('co-tableau-0');
      fireEvent.click(screen.getByTestId('co-waste-button'));
      fireEvent.click(screen.getByTestId('co-foundation-0'));
      await waitFor(() =>
        expect(
          screen
            .getAllByRole('status')
            .some((region) => region.textContent?.includes('Moved from Waste to Foundation')),
        ).toBe(true),
      );
    } finally {
      await i18n.changeLanguage(originalLanguage);
    }
  });

  // Half the foundations run the other way, and the board is unreadable without
  // knowing which is which.
  it('shows the next rank for both directions', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-foundation-next-0')).toHaveTextContent('2'));
    expect(screen.getByTestId('co-foundation-next-4')).toHaveTextContent('Q');
    // The empty piles still show their opening rank, one per direction.
    expect(screen.getByTestId('co-foundation-next-1')).toHaveTextContent('A');
    expect(screen.getByTestId('co-foundation-next-5')).toHaveTextContent('K');
  });

  it('renders every tableau pile', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    expect(screen.getByTestId(`co-tableau-${(TABLEAU_CNT - 1).toString()}`)).toBeInTheDocument();
    expect(screen.getByTestId('co-tableau-0')).toHaveAccessibleName('場札 0、1枚、一番上のカード ♠ A');
    expect(screen.getByTestId('co-tableau-0').querySelectorAll('[data-testid="animated-card"]')).toHaveLength(1);
  });

  it('shows cards buried under a tableau pile', async () => {
    const piles = tableau();
    piles[7] = [card('SPADE', 2), card('HEART', 12)];
    mockExec.mockResolvedValue(makeState({ tableau: piles }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-7')).toBeInTheDocument());
    const pile = screen.getByTestId('co-tableau-7');
    expect(pile).toHaveAccessibleName('場札 7、2枚、一番上のカード ♥ Q');
    expect(pile.querySelectorAll('[data-testid="animated-card"]')).toHaveLength(2);
  });

  it('draws from the stock', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-draw-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('describes the draw button with the current stock count, including zero', async () => {
    mockExec.mockResolvedValueOnce(makeState({ stockCount: 71 })).mockResolvedValueOnce(makeState({ stockCount: 70 }));
    const { unmount } = renderWithProviders(<ColoradoPage />);
    const drawButton = await screen.findByTestId('co-draw-button');
    const stockCount = screen.getByTestId('co-stock-count');
    expect(stockCount).toHaveTextContent('71枚');
    expect(drawButton).toHaveAttribute('aria-describedby', 'co-stock-count');
    expect(drawButton).toHaveAccessibleDescription('71枚');

    fireEvent.click(drawButton);
    await waitFor(() => expect(stockCount).toHaveTextContent('70枚'));
    expect(drawButton).toHaveAccessibleDescription('70枚');

    unmount();
    mockExec.mockResolvedValue(makeState({ stockCount: 0 }));
    renderWithProviders(<ColoradoPage />);
    expect(await screen.findByTestId('co-stock-count')).toHaveTextContent('0枚');
    expect(screen.getByTestId('co-draw-button')).toHaveAccessibleDescription('0枚');
  });

  it('shows the waste count, including zero, and describes it for screen readers', async () => {
    mockExec.mockResolvedValue(makeState({ waste: [card('HEART', 6), card('CLOVER', 9), card('DIAMOND', 2)] }));
    const { unmount } = renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-count')).toHaveTextContent('3枚'));
    expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-describedby', 'co-waste-count');

    unmount();
    mockExec.mockResolvedValue(makeState({ waste: [] }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-count')).toHaveTextContent('0枚'));
    expect(screen.getByTestId('co-waste-button')).toHaveAccessibleDescription('0枚');
  });

  it('disables the draw button once the stock is empty', async () => {
    mockExec.mockResolvedValue(makeState({ stockCount: 0 }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeDisabled());
  });

  it('moves the waste card to a foundation once selected', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation' }));
  });

  // The whole game is this move: the waste goes onto ANY pile, suit and rank
  // irrelevant, burying whatever was there.
  it('buries the waste card on any tableau pile', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-tableau-7'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', idx: 7 }));
  });

  it('moves a tableau top to a foundation', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', idx: 0 }, { zone: 'foundation' }),
    );
  });

  it('fills an empty pile straight from the stock', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-stock-fill-button'));
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-tableau-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', idx: 3 }));
  });

  // The stock may only fill a gap, so clicking an occupied pile must do nothing.
  it('refuses to put the stock card onto an occupied pile', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-stock-fill-button'));
    await waitFor(() => expect(screen.getByTestId('co-stock-fill-button')).toHaveAttribute('aria-pressed', 'true'));

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('co-tableau-5'));
    // Without this await the assertion passes whether or not a move fired (#4439).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the empty pile DOES accept it.
    fireEvent.click(screen.getByTestId('co-tableau-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', idx: 3 }));
  });

  it('hides the gap-fill button when no pile is empty', async () => {
    const piles = Array.from({ length: TABLEAU_CNT }, () => [card('SPADE', 7)]);
    mockExec.mockResolvedValue(makeState({ tableau: piles }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeInTheDocument());
    expect(screen.queryByTestId('co-stock-fill-button')).not.toBeInTheDocument();
  });

  it('keeps an empty tableau target focusable and explains that a source must be selected', async () => {
    renderWithProviders(<ColoradoPage />);
    const target = await screen.findByTestId('co-tableau-3');
    expect(target).not.toBeDisabled();
    expect(target).toHaveAttribute('aria-disabled', 'true');
    const describedBy = target.getAttribute('aria-describedby');
    expect(describedBy).toBeTruthy();
    expect(document.getElementById(describedBy ?? '')).toHaveTextContent('先に移動する札を選んでください');

    mockExec.mockClear();
    fireEvent.click(target);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('does not move when nothing is selected', async () => {
    renderWithProviders(<ColoradoPage />);
    const foundation = await screen.findByTestId('co-foundation-0');
    expect(foundation).not.toBeDisabled();
    expect(foundation).toHaveAttribute('aria-disabled', 'true');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the same click DOES move once a source is selected.
    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation' }));
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 },
        messageCode: 'colorado.hintAvailable',
      }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(hintLiveRegion()).toHaveTextContent('ヒントがあります'));
  });

  // The other half of the gate: a passive hint must not surface the banner.
  it('hides the hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-draw-button')).toBeInTheDocument());
    // **領域は常設で中身が空。**読み上げは「既にある領域の変化」でしか起きない
    // ので、領域ごと消してはいけない (#5955)。文言が出ていないことを見る。
    expect(hintLiveRegion()).toBeInTheDocument();
    expect(hintLiveRegion()).toHaveTextContent('');
    expect(screen.queryByText(/ヒントがあります/)).not.toBeInTheDocument();
  });

  it('names the stock in the hint banner when the hint is to draw', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'colorado.hintAvailable',
      }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(hintLiveRegion()).toHaveTextContent('山札'));
  });

  it('requests a hint', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('undoes only when there is history', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());

    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '元に戻す' })[1]).toBeEnabled());
    fireEvent.click(screen.getAllByRole('button', { name: '元に戻す' })[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // "Bury the waste somewhere" is always legal, so gating auto-complete on any
  // hint at all would leave the button lit for the whole game.
  it('enables auto-complete only for a move to a foundation', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'tableau', toIdx: 2 } }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeDisabled());

    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getAllByTestId('autocomplete-button')[1]).toBeEnabled());
    fireEvent.click(screen.getAllByTestId('autocomplete-button')[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('deselects via the cancel button', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-waste-button'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.getByTestId('co-waste-button')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('clicking the selected pile again deselects it', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('hides the playing controls once the game clears', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.queryByTestId('co-draw-button')).not.toBeInTheDocument());
  });

  it('gives up through the confirm dialog', async () => {
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('hides the playing controls after a give-up', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.queryByTestId('co-draw-button')).not.toBeInTheDocument());
  });

  it('shows an error with a retry', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<ColoradoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再試行|retry/i })).toBeInTheDocument());
  });

  it('renders stalemate escape button when stuck and dispatches undo_n with the escape count', async () => {
    mockExec.mockResolvedValue(makeState({ isStalemate: true, undoToEscape: 3, canUndo: true }));
    renderWithProviders(<ColoradoPage />);
    const escapeBtn = await screen.findByTestId('stalemate-escape-button');
    expect(escapeBtn).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(escapeBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('does not render stalemate escape button when isStalemate is false', async () => {
    // undoToEscape is set to a positive value to verify gating on isStalemate
    mockExec.mockResolvedValue(makeState({ isStalemate: false, undoToEscape: 3 }));
    renderWithProviders(<ColoradoPage />);
    // **描画が終わるまで待つ。**reset の呼び出しを待つだけだと、まだ骨組みしか
    // 出ていない時点で「無い」と読めてしまい、条件を外しても通る。
    expect(await screen.findByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('does not render stalemate escape button when undoToEscape is 0 even if isStalemate is true', async () => {
    mockExec.mockResolvedValue(makeState({ isStalemate: true, undoToEscape: 0 }));
    renderWithProviders(<ColoradoPage />);
    // **描画が終わるまで待つ。**reset の呼び出しを待つだけだと、まだ骨組みしか
    // 出ていない時点で「無い」と読めてしまい、条件を外しても通る。
    expect(await screen.findByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });
});
