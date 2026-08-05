import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { desmocheApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, DesmochePlayer, DesmocheResponse } from '../types/card';
import { DesmochePhase } from '../types/phases';
import { DesmochePage } from './DesmochePage';

vi.mock('../api/gameApi', () => ({
  desmocheApi: { exec: vi.fn() },
  actionLogApi: { desmoche: vi.fn() },
}));

const mockExec = vi.mocked(desmocheApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<DesmochePlayer>): DesmochePlayer {
  return {
    id,
    isHuman,
    cardCount: 9,
    cards: isHuman ? [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 2)] : [],
    score: -10,
    meldedCount: 0,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<DesmocheResponse>): DesmocheResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: DesmochePhase.ACT,
    currentPlayerIdx: 0,
    stockCount: 15,
    discardTop: card('HEART', 9),
    melds: [{ owner: 0, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] }],
    roundNo: 0,
    pot: 40,
    goOutSize: 10,
    roundWinner: -1,
    roundExhausted: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

/** Select the given hand indices by clicking them. */
function selectCards(indices: number[]) {
  const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'discard');
  for (const i of indices) {
    fireEvent.click(handButtons[i]);
  }
}

describe('DesmochePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently and the pot', async () => {
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // 上がりは配札 9 枚ではなく 10 枚、そしてポーカーの役は使わない。
    expect(screen.getByText(/ちょうど10枚/)).toBeInTheDocument();
    expect(screen.getByText(/ポーカーの役は使いません/)).toBeInTheDocument();
    expect(screen.getByText(/ポット40/)).toBeInTheDocument();
  });

  it('offers the two draws only in the draw step', async () => {
    mockExec.mockResolvedValue(makeState({ phase: DesmochePhase.DRAW }));
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '捨て札を取る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('needs three selected cards before it will meld', async () => {
    // 2 枚では押せない。押せてしまうとサーバー往復が無駄になる。
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    selectCards([0, 1]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    selectCards([2]);
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, [0, 1, 2]));
  });

  it('lays off exactly one card onto a chosen meld', async () => {
    // 手札の添字と場の添字は別物なので、両方を選ばないと押せない。
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    selectCards([3]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '付ける' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(screen.getAllByTestId('desmoche-meld')[0]);
    fireEvent.click(screen.getByRole('button', { name: '付ける' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', 3, 0));
  });

  it('discards exactly one card', async () => {
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    selectCards([0, 1]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    selectCards([1]); // deselect, leaving one
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  // 組み替えは「場のカード」と「移し先のメルド」という、手札とは別の 2 つの
  // 選択を要る。片方だけでは押せない。
  it('needs a table card and a target meld before it will rearrange', async () => {
    mockExec.mockResolvedValue(
      makeState({
        melds: [
          { owner: 0, kind: 1, cards: [card('HEART', 4), card('HEART', 5), card('HEART', 6), card('HEART', 7)] },
          { owner: 0, kind: 0, cards: [card('SPADE', 7), card('CLOVER', 7), card('DIAMOND', 7)] },
        ],
      }),
    );
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '組み替え' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    // ラン ♥4-5-6-7 の 4 枚目 (♥7) を選び、セットへ移す。
    fireEvent.click(screen.getAllByTestId('desmoche-meld-card')[3]);
    fireEvent.click(screen.getAllByTestId('desmoche-meld')[1]);
    fireEvent.click(screen.getByRole('button', { name: '組み替え' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('desmoche', 3, undefined, undefined, {
        fromMeldIndex: 0,
        toMeldIndex: 1,
      }),
    );
  });

  it('will not rearrange a card back into the meld it came from', async () => {
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    fireEvent.click(screen.getAllByTestId('desmoche-meld-card')[0]);
    fireEvent.click(screen.getAllByTestId('desmoche-meld')[0]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '組み替え' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // セットとランで付けられる札が違うので、種別が見えている必要がある。
  // 1 つの it で 2 度 render すると前の DOM が残り、[0] が古い方を指すので分ける。
  it('names a set on the table', async () => {
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(screen.getAllByTestId('desmoche-meld')[0]).toHaveTextContent('セット'));
  });

  it('names a run on the table', async () => {
    mockExec.mockResolvedValue(
      makeState({ melds: [{ owner: 0, kind: 1, cards: [card('SPADE', 5), card('SPADE', 6), card('SPADE', 7)] }] }),
    );
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(screen.getAllByTestId('desmoche-meld')[0]).toHaveTextContent('ラン'));
  });

  it('names the round winner', async () => {
    mockExec.mockResolvedValue(makeState({ phase: DesmochePhase.ROUND_END, roundWinner: 2 }));
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(screen.getByTestId('desmoche-round-result')).toHaveTextContent('席2'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンドへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // 勝者なしを黙って通すと、ポットが増えた理由が画面から読めない。
  it('says when nobody won, so the carried-over pot is explained', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: DesmochePhase.ROUND_END, roundWinner: -1, roundExhausted: true, pot: 80 }),
    );
    renderWithProviders(<DesmochePage />);
    await waitFor(() => expect(screen.getByTestId('desmoche-round-result')).toHaveTextContent('80'));
    expect(screen.getByTestId('desmoche-round-result')).toHaveTextContent('持ち越し');
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'desmoche.win', '最も多く稼ぎました'],
      [2, 'desmoche.lose', '及びませんでした'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: DesmochePhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<DesmochePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **自分のメルドだけが上がり枚数に加算される。**番号を読んで暗算しないと、
  // 進捗の進まない先にレイオフして手番を捨てることになる (#4932)。
  describe('meld ownership', () => {
    const twoMelds = () =>
      makeState({
        melds: [
          { owner: 0, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] },
          { owner: 2, kind: 0, cards: [card('SPADE', 9), card('HEART', 9), card('CLOVER', 9)] },
        ],
      });

    it('marks which melds are yours', async () => {
      mockExec.mockResolvedValue(twoMelds());
      renderWithProviders(<DesmochePage />);
      await waitFor(() => expect(screen.getAllByTestId('desmoche-meld')).toHaveLength(2));

      const labels = screen.getAllByTestId('desmoche-meld');
      expect(labels[0]).toHaveTextContent('あなたのメルド');
      expect(labels[1]).not.toHaveTextContent('あなたのメルド');

      // 色でも区別が付く (テキストを読まずに済む)。
      const rows = document.querySelectorAll('[data-own-meld]');
      expect(rows[0]).toHaveAttribute('data-own-meld', 'true');
      expect(rows[1]).toHaveAttribute('data-own-meld', 'false');
    });

    it('warns when a foreign meld is picked as the lay-off target', async () => {
      mockExec.mockResolvedValue(twoMelds());
      renderWithProviders(<DesmochePage />);
      await waitFor(() => expect(screen.getAllByTestId('desmoche-meld')).toHaveLength(2));

      fireEvent.click(screen.getAllByTestId('desmoche-meld')[1]);
      expect(screen.getByTestId('desmoche-foreign-meld-warning')).toHaveTextContent('加算されません');
    });

    it('stays quiet for your own meld', async () => {
      mockExec.mockResolvedValue(twoMelds());
      renderWithProviders(<DesmochePage />);
      await waitFor(() => expect(screen.getAllByTestId('desmoche-meld')).toHaveLength(2));

      fireEvent.click(screen.getAllByTestId('desmoche-meld')[0]);
      expect(screen.queryByTestId('desmoche-foreign-meld-warning')).not.toBeInTheDocument();
    });
  });
});
