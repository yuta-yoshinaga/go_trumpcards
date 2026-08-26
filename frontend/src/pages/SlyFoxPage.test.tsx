import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { slyFoxApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SlyFoxResponse } from '../types/card';
import { SlyFoxPage, slyfoxNextRank } from './SlyFoxPage';

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
  slyFoxApi: { exec: vi.fn() },
  actionLogApi: { slyfox: vi.fn() },
}));

const mockExec = vi.mocked(slyFoxApi.exec);
const card = (design: CardDesign, value: number): Card => ({ design, value });

const TABLEAU_CNT = 20;

/** 20 piles of one card, with pile 3 emptied so the gap paths are reachable. */
function tableau(): Card[][] {
  const piles = Array.from({ length: TABLEAU_CNT }, (_, i) => [card('SPADE', ((i % 13) + 1) as number)]);
  piles[3] = [];
  return piles;
}

function makeState(overrides: Partial<SlyFoxResponse> = {}): SlyFoxResponse {
  return {
    tableau: tableau(),
    // F0 spades ascending holds the Ace; F4 spades descending holds the King.
    foundation: [[card('SPADE', 1)], [], [], [], [card('SPADE', 13)], [], [], []],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    dealtThisCycle: 20,
    dealCycle: 20,
    reserveLocked: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    messageCode: 'slyfox.playing',
    ...overrides,
  };
}

describe('slyfoxNextRank', () => {
  it('opens an ascending foundation at the Ace and a descending one at the King', () => {
    expect(slyfoxNextRank(undefined, 0, true)).toBe(1);
    expect(slyfoxNextRank(undefined, 0, false)).toBe(13);
  });

  it('steps up or down depending on the direction', () => {
    expect(slyfoxNextRank(1, 1, true)).toBe(2);
    expect(slyfoxNextRank(13, 1, false)).toBe(12);
  });

  it('returns null once the pile is complete', () => {
    expect(slyfoxNextRank(13, 13, true)).toBeNull();
    expect(slyfoxNextRank(1, 13, false)).toBeNull();
  });
});

describe('SlyFoxPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // Half the foundations run the other way, and the board is unreadable without
  // knowing which is which.
  it('shows the next rank for both directions', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-foundation-next-0')).toHaveTextContent('2'));
    expect(screen.getByTestId('co-foundation-next-4')).toHaveTextContent('Q');
    // The empty piles still show their opening rank, one per direction.
    expect(screen.getByTestId('co-foundation-next-1')).toHaveTextContent('A');
    expect(screen.getByTestId('co-foundation-next-5')).toHaveTextContent('K');
  });

  it('renders every tableau pile', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    expect(screen.getByTestId(`co-tableau-${(TABLEAU_CNT - 1).toString()}`)).toBeInTheDocument();
  });

  it('opens deal mode from the stock', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-deal-button'));
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toHaveAttribute('aria-pressed', 'true'));
  });

  it('disables the deal button once the stock is empty', async () => {
    mockExec.mockResolvedValue(makeState({ stockCount: 0 }));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toBeDisabled());
  });

  // **配り先を選ぶまで札は動かない。**捨て札が無いので、めくった札の置き場所を
  // 決める 1 手がそのまま 20 枚のうちの 1 枚になる。
  it('deals onto the reserve slot you pick', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-deal-button'));
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-tableau-7'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal', undefined, { zone: 'tableau', idx: 7 }));
  });

  // 組札へ直接送る手は 20 枚に数えないので、別のリクエストとして出す。
  it('deals straight to a foundation', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-deal-button'));
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal', undefined, { zone: 'foundation', idx: 0 }));
  });

  it('moves a reserve top to a foundation', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));

    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', idx: 0 }, { zone: 'foundation' }),
    );
  });

  // **周を配り切るまでリザーブは選べない。**選べてしまうと、組札を押した瞬間に
  // サーバが拒む。
  it('locks the reserve until the round is dealt out', async () => {
    mockExec.mockResolvedValue(makeState({ reserveLocked: true, dealtThisCycle: 13 }));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeDisabled());
    // なぜ押せないのかを画面に出すこと。
    expect(screen.getByTestId('co-cycle-status')).toHaveTextContent('13');
  });

  // 負のコントロール: 開いていれば選べる。閉じっぱなしの実装でも上は通る。
  it('unlocks the reserve once the round is dealt out', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).not.toBeDisabled());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));
  });

  it('does not move when nothing is selected', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-foundation-0')).toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: the same click DOES move once a source is selected.
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('co-foundation-0'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', idx: 0 }, { zone: 'foundation' }),
    );
  });

  it('hides the hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toBeInTheDocument());
    // **領域は常設で中身が空。**読み上げは「既にある領域の変化」でしか起きない
    // ので、領域ごと消してはいけない (#5955)。文言が出ていないことを見る。
    expect(hintLiveRegion()).toBeInTheDocument();
    expect(hintLiveRegion()).toHaveTextContent('');
    expect(screen.queryByText(/ヒントがあります/)).not.toBeInTheDocument();
  });

  it('names the stock and the slot in the hint banner when the hint is to deal', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 6 },
        messageCode: 'slyfox.hintAvailable',
      }),
    );
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(hintLiveRegion()).toHaveTextContent('山札'));
    // 行き先の枠まで言う。捨て札が無いので「めくる」だけでは手にならない。
    expect(hintLiveRegion()).toHaveTextContent('6');
  });

  it('requests a hint', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('undoes only when there is history', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());

    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '元に戻す' })[1]).toBeEnabled());
    fireEvent.click(screen.getAllByRole('button', { name: '元に戻す' })[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // 配るヒントは常に出るので、それで有効にすると全編ボタンが光る。
  it('enables auto-complete only for a reserve move to a foundation', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 2 } }));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeDisabled());

    mockExec.mockResolvedValue(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 4, toZone: 'foundation', toIdx: 0 } }),
    );
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getAllByTestId('autocomplete-button')[1]).toBeEnabled());
    fireEvent.click(screen.getAllByTestId('autocomplete-button')[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('deselects via the cancel button', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-deal-button'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.getByTestId('co-deal-button')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('clicking the selected pile again deselects it', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('co-tableau-0'));
    await waitFor(() => expect(screen.getByTestId('co-tableau-0')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('hides the playing controls once the game clears', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.queryByTestId('co-deal-button')).not.toBeInTheDocument());
  });

  it('gives up through the confirm dialog', async () => {
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('hides the playing controls after a give-up', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.queryByTestId('co-deal-button')).not.toBeInTheDocument());
  });

  it('shows an error with a retry', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<SlyFoxPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再試行|retry/i })).toBeInTheDocument());
  });
});
