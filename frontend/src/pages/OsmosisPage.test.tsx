import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { osmosisApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, OsmosisResponse } from '../types/card';
import { OsmosisPage } from './OsmosisPage';

vi.mock('../api/gameApi', () => ({
  osmosisApi: { exec: vi.fn() },
  actionLogApi: { osmosis: vi.fn() },
}));

const mockExec = vi.mocked(osmosisApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: OsmosisResponse = {
  isStalemate: false,
  reserve: [[card('SPADE', 2)], [card('HEART', 9)], [card('CLOVER', 4)], [card('DIAMOND', 10)]],
  stockCount: 34,
  waste: [card('HEART', 4)],
  foundation: [[card('SPADE', 5)], [], [], []],
  baseRank: 5,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  message: '',
  messageCode: 'osmosis.playing',
};

const gameClearState: OsmosisResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'osmosis.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: OsmosisResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'osmosis.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('OsmosisPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // **手詰まりでもサーバは Playing を返す (#4808)。**「プレイ中」と出たままだと、
  // もう動かない盤面をめくり続けることになる。出る側と出ない側の両方を踏む。
  it('announces the dead end once no card can be placed', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, messageCode: 'osmosis.stalemate' });
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByText(/手詰まりです/)).toBeInTheDocument());
  });

  it('says nothing about a dead end while a move remains', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getAllByText('プレイ中').length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByText(/手詰まりです/)).not.toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows base rank', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByText(/ベースランク/)).toBeInTheDocument());
  });

  it('shows the allowed-rank guide per foundation row', async () => {
    // foundation [[♠5],[],[],[]], baseRank 5.
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByTestId('os-allowed-0')).toBeInTheDocument());
    // Row 0 is the base row (★) with a fixed suit → any rank.
    expect(screen.getByTestId('os-allowed-0')).toHaveTextContent('★');
    expect(screen.getByTestId('os-allowed-0')).toHaveTextContent('任意');
    // Row 1 (empty, row 0 non-empty) accepts the base rank 5.
    expect(screen.getByTestId('os-allowed-1')).toHaveTextContent('5');
    // Row 2 (empty, row 1 empty) accepts nothing yet.
    expect(screen.getByTestId('os-allowed-2')).toHaveTextContent('—');
  });

  it('clicks stock to fire draw command', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: /山札/ }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('selects waste then moves it to a foundation row', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ウェイスト' }).click();
    // Foundation row 0 becomes enabled once a source is selected.
    await waitFor(() => expect(screen.getByRole('button', { name: '組札 0' })).toBeEnabled());
    screen.getByRole('button', { name: '組札 0' }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', col: 0 }),
    );
  });

  it('selects a reserve column then moves it to a foundation row', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'リザーブ 1' }).click();
    await waitFor(() => expect(screen.getByRole('button', { name: '組札 2' })).toBeEnabled());
    screen.getByRole('button', { name: '組札 2' }).click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve', col: 1 }, { zone: 'foundation', col: 2 }),
    );
  });

  it('flags foundation rows the selected card cannot be placed on', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // ♥9 (reserve col 1): wrong suit for the ♠ base row and not the base rank
    // for the empty rows → cannot be placed anywhere.
    screen.getByRole('button', { name: 'リザーブ 1' }).click();
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '組札 0' })).toHaveAttribute('title', 'この段には置けません'),
    );
    expect(screen.getByRole('button', { name: '組札 0' }).className).toContain('border-ds-error');
  });

  it('foundation rows are disabled until a source is selected', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '組札 0' })).toBeDisabled();
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('autocomplete button triggers autocomplete command', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '自動完成' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button fires undo when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('advertises the keyboard shortcuts on the action buttons', async () => {
    renderWithProviders(<OsmosisPage />);
    const draw = await screen.findByRole('button', { name: '引く' });
    expect(draw).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(draw.querySelector('kbd')?.textContent).toBe('D');
    // KbdBadge text is aria-hidden, so button accessible names stay clean.
    expect(screen.getByRole('button', { name: 'ヒント' })).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(screen.getByRole('button', { name: '自動完成' })).toHaveAttribute('aria-keyshortcuts', 'a');
    expect(screen.getByRole('button', { name: '元に戻す' })).toHaveAttribute('aria-keyshortcuts', 'z');
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toHaveAttribute('aria-keyshortcuts', 'g');
  });

  it('fires draw when the d key is pressed while playing', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('fires hint when the h key is pressed', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('fires autocomplete when the a key is pressed', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('fires undo when the z key is pressed', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('routes the g key through the give-up confirm dialog', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'g' });
    // The key must not dispatch giveup directly — it opens the confirm dialog first.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('does not fire shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    fireEvent.keyDown(document.body, { key: 'h' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('draw');
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('hint');
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase and hides action buttons', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders empty waste placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [] });
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ウェイスト' })).not.toBeInTheDocument();
  });

  describe('drag and drop', () => {
    // Minimal stand-in for the browser DataTransfer object (jsdom lacks one).
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

    it('reserve and waste top cards are draggable while playing', async () => {
      renderWithProviders(<OsmosisPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      expect(screen.getByRole('button', { name: 'ウェイスト' })).toHaveAttribute('draggable', 'true');
      expect(screen.getByRole('button', { name: 'リザーブ 0' })).toHaveAttribute('draggable', 'true');
    });

    it('dragging the waste top onto a foundation row dispatches move', async () => {
      mockExec.mockResolvedValue({ ...playingState, waste: [card('SPADE', 3)] });
      renderWithProviders(<OsmosisPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      const wasteBtn = screen.getByRole('button', { name: 'ウェイスト' });
      const dt = buildDataTransfer();
      fireEvent.dragStart(wasteBtn, { dataTransfer: dt });
      const dropZone = screen.getByRole('button', { name: '組札 0' }).parentElement as HTMLElement;
      mockExec.mockClear();
      fireEvent.dragOver(dropZone, { dataTransfer: dt });
      fireEvent.drop(dropZone, { dataTransfer: dt });
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', col: 0 }),
      );
    });

    it('dragging a reserve top onto a foundation row dispatches move', async () => {
      renderWithProviders(<OsmosisPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      const reserveBtn = screen.getByRole('button', { name: 'リザーブ 0' });
      const dt = buildDataTransfer();
      fireEvent.dragStart(reserveBtn, { dataTransfer: dt });
      const dropZone = screen.getByRole('button', { name: '組札 0' }).parentElement as HTMLElement;
      mockExec.mockClear();
      fireEvent.dragOver(dropZone, { dataTransfer: dt });
      fireEvent.drop(dropZone, { dataTransfer: dt });
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve', col: 0 }, { zone: 'foundation', col: 0 }),
      );
    });

    it('a drop with no active drag does not dispatch a move', async () => {
      renderWithProviders(<OsmosisPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      const dropZone = screen.getByRole('button', { name: '組札 0' }).parentElement as HTMLElement;
      mockExec.mockClear();
      // No dragStart ran, so the dataTransfer carries no source payload.
      fireEvent.drop(dropZone, { dataTransfer: buildDataTransfer() });
      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalled();
    });

    it('marks a foundation row that cannot accept the dragged card with an error border', async () => {
      renderWithProviders(<OsmosisPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
      // Reserve col 1 top is ♥9 — wrong suit for the ♠ base row and not the base rank.
      const reserveBtn = screen.getByRole('button', { name: 'リザーブ 1' });
      const dt = buildDataTransfer();
      fireEvent.dragStart(reserveBtn, { dataTransfer: dt });
      const foundation0 = screen.getByRole('button', { name: '組札 0' });
      fireEvent.dragOver(foundation0.parentElement as HTMLElement, { dataTransfer: dt });
      await waitFor(() => expect(foundation0.className).toContain('border-ds-error'));
      expect(foundation0).toHaveAttribute('title', 'この段には置けません');
    });
  });
});

// #5625: 置けない段は赤枠と `title` だけで示していた。`title` は支援技術が
// 読み上げる保証が無いので、キーボード/スクリーンリーダー利用者には**理由が
// どこにも無い**状態だった。
describe('OsmosisPage blocked foundation rows', () => {
  it('puts the reason in the accessible name of the rows that reject the card', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByTestId('os-allowed-0')).toBeInTheDocument());

    // ♥4 を選ぶ。ベース段 (♠) にも、まだ何も入っていない下の段にも置けない。
    fireEvent.click(screen.getByRole('button', { name: 'ウェイスト' }));

    // 名前は状態で変えず、理由は説明として結びつける。
    const row0 = await screen.findByRole('button', { name: '組札 0' });
    await waitFor(() => expect(row0).toHaveAttribute('aria-describedby'));
    const id = row0.getAttribute('aria-describedby') as string;
    expect(document.getElementById(id)).toHaveTextContent('この段には置けません');
    // 視覚 (赤枠) と支援技術が同じ段を指す。
    expect(row0.className).toContain('border-ds-error');
  });

  it('leaves the accessible name alone when nothing is selected', async () => {
    renderWithProviders(<OsmosisPage />);
    await waitFor(() => expect(screen.getByTestId('os-allowed-0')).toBeInTheDocument());

    expect(screen.getByRole('button', { name: '組札 0' })).not.toHaveAttribute('aria-describedby');
    expect(screen.queryByText('この段には置けません')).not.toBeInTheDocument();
  });
});
