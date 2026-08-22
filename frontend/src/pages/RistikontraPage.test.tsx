import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ristikontraApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RistikontraPlayer, RistikontraResponse } from '../types/card';
import { RistikontraPage } from './RistikontraPage';

vi.mock('../api/gameApi', () => ({
  ristikontraApi: { exec: vi.fn() },
  actionLogApi: { ristikontra: vi.fn() },
}));

const mockExec = vi.mocked(ristikontraApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<RistikontraPlayer> = {}): RistikontraPlayer {
  return {
    id: 1,
    isHuman: false,
    cardCount: 4,
    cards: [],
    capturedCount: 0,
    finalScore: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<RistikontraResponse> = {}): RistikontraResponse {
  return {
    players: [
      makePlayer({
        id: 0,
        isHuman: true,
        cards: [card('SPADE', 5), card('HEART', 11), card('DIAMOND', 1), card('CLOVER', 9)],
      }),
      makePlayer({ id: 1 }),
      makePlayer({ id: 2 }),
      makePlayer({ id: 3 }),
    ],
    currentTurn: 0,
    pile: [card('CLOVER', 7)],
    pileTop: card('CLOVER', 7),
    pileCount: 1,
    lastCaptureIdx: -1,
    counterRank: 0,
    gameEndFlag: false,
    phase: 'play',
    remainingDeck: 36,
    winners: [],
    finalScores: [],
    config: { playerCnt: 4, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

const playState = makeState();
const emptyPileState = makeState({ pile: [], pileTop: null, pileCount: 0 });
const gameEndState = makeState({
  phase: 'gameEnd',
  gameEndFlag: true,
  currentTurn: -1,
  winners: [0],
  finalScores: [11, 7, 4, 8],
  players: [
    makePlayer({ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 28, finalScore: 11 }),
    makePlayer({ id: 1, cardCount: 0, finalScore: 7 }),
    makePlayer({ id: 2, cardCount: 0, finalScore: 4 }),
    makePlayer({ id: 3, cardCount: 0, finalScore: 8 }),
  ],
});

beforeEach(() => {
  // Clear persisted CLI-mode state so an earlier CLI-toggle test does not leave the
  // terminal enabled and hide the hand in later tests.
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('RistikontraPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<RistikontraPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the remaining stock', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText(/山札: 36枚/)).toBeInTheDocument());
  });

  it('renders the players list with captured counts', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.getAllByText(/捕獲 0枚/).length).toBe(4);
  });

  it('renders the human hand cards', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-3')).toBeInTheDocument();
  });

  it('names each hand card in its aria-label and announces the turn', async () => {
    renderWithProviders(<RistikontraPage />);
    // hand[0] is ♠5 → "♠ 5 を出す".
    expect(await screen.findByRole('button', { name: '♠ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ A を出す' })).toBeInTheDocument();
    // The turn notice is a live region so turn arrival is announced.
    expect(screen.getByTestId('ristikontra-turn-notice')).toHaveAttribute('role', 'status');
  });

  it('plays a hand card on the human turn', async () => {
    renderWithProviders(<RistikontraPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
  });

  // 打ち返し (steal) の合図は**獲得枚数が減ること**。札が席から離れるのは
  // 奪われたときだけなので、減少は打ち返しの確実な signal。
  // クローン元 (ピシュティ) は Pişti ボーナスの上昇を祝っていたが、
  // このゲームにボーナスは無く、そのバッジは一度も出せない。
  it('celebrates a counter when a captured pile is stolen', async () => {
    // **奪われる前の状態から始める。** 最初の応答が基準になるので、
    // 席 1 が束を持っているところから始めないと「減った」が観測できない。
    const human = playState.players[0];
    mockExec.mockResolvedValue(
      makeState({
        players: [
          { ...human, capturedCount: 1 },
          makePlayer({ id: 1, capturedCount: 6 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    renderWithProviders(<RistikontraPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    expect(screen.queryByTestId('ristikontra-counter-celebration')).not.toBeInTheDocument();

    // 席 1 が持っていた 6 枚を、席 0 が打ち返しで奪った状態。
    mockExec.mockResolvedValueOnce(
      makeState({
        players: [
          { ...human, capturedCount: 7 },
          makePlayer({ id: 1, capturedCount: 0 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    fireEvent.click(cardBtn);
    expect(await screen.findByTestId('ristikontra-counter-celebration')).toBeInTheDocument();
  });

  it('stays quiet on an ordinary capture, where nobody loses cards', async () => {
    renderWithProviders(<RistikontraPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockResolvedValueOnce(
      makeState({
        players: [
          { ...playState.players[0], capturedCount: 3 },
          makePlayer({ id: 1 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('ristikontra-counter-celebration')).not.toBeInTheDocument();
  });

  it('does not mistake a new deal for a counter', async () => {
    // **配り直しは全員の獲得枚数が同時に 0 に戻る。** 個々の減少だけを見ると
    // これを打ち返しと読んでしまうので、合計が減った場合は祝わない。
    renderWithProviders(<RistikontraPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockResolvedValueOnce(
      makeState({
        players: [
          { ...playState.players[0], capturedCount: 9 },
          makePlayer({ id: 1, capturedCount: 5 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockResolvedValueOnce(
      makeState({
        players: [
          { ...playState.players[0], capturedCount: 0 },
          makePlayer({ id: 1, capturedCount: 0 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    fireEvent.change(screen.getByLabelText('CPU難易度'), { target: { value: '2' } });
    await waitFor(() => expect(screen.queryByTestId('ristikontra-counter-celebration')).not.toBeInTheDocument());
  });

  it('shows the empty-pile label when the pile is empty', async () => {
    mockExec.mockResolvedValue(emptyPileState);
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText('場札なし')).toBeInTheDocument());
  });

  it('does not dispatch play when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2 }));
    renderWithProviders(<RistikontraPage />);
    const cardBtn = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('shows the win message and a next-game button when the human wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝利です！')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '次のゲームへ' })).toBeInTheDocument();
  });

  it('dispatches next when next-game is clicked', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<RistikontraPage />);
    const btn = await screen.findByRole('button', { name: '次のゲームへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows final scores on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText(/11点/)).toBeInTheDocument());
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2, playerCnt: 4 } }));
  });

  // **席数を選ばせない。** リスティコントラは常に 2 対 2 の固定
  // パートナーシップで、クローン元のピシュティにあった 2/3/4 の選択肢を
  // 残すとチームを組めない卓を頼めてしまう (バックエンドが弾くので、
  // 押しても何も起きない死んだコントロールになる)。
  it('offers no player-count selector, and always resets to a 4-player table', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.queryByLabelText('プレイヤー数')).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.change(screen.getByLabelText('CPU難易度'), { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2, playerCnt: 4 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });

  it('highlights capturing hand cards: Jack in accent, pile-top match in success', async () => {
    // Pile top is rank 5, so the SPADE 5 (index 0) captures (success); HEART J (index 1)
    // always captures (accent); DIAMOND A (2) and CLOVER 9 (3) do not.
    mockExec.mockResolvedValue(makeState({ pile: [card('CLOVER', 5)], pileTop: card('CLOVER', 5) }));
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-0').className).toContain('ring-ds-success');
    expect(screen.getByTestId('hand-card-1').className).toContain('ring-ds-accent');
    expect(screen.getByTestId('hand-card-2').className).not.toContain('ring-ds-');
    expect(screen.getByTestId('hand-card-3').className).not.toContain('ring-ds-');
  });

  it('highlights only the Jack when the pile is empty', async () => {
    mockExec.mockResolvedValue(makeState({ pile: [], pileTop: null, pileCount: 0 }));
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-1')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1').className).toContain('ring-ds-accent');
    expect(screen.getByTestId('hand-card-0').className).not.toContain('ring-ds-');
  });

  it('does not highlight capturing cards on a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 1, pile: [card('CLOVER', 5)], pileTop: card('CLOVER', 5) }));
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-1')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1').className).not.toContain('ring-ds-');
    expect(screen.getByTestId('hand-card-0').className).not.toContain('ring-ds-');
  });

  // 途中経過は**チームの獲得枚数**。クローン元は席ごとの近似スコア
  // (確定ボーナス + 最多捕獲 +3) を出していたが、このゲームには
  // ボーナスもカード点も無いので、枚数がそのまま結果になる。
  it('shows the team total during play, not the seat total', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          { ...playState.players[0], capturedCount: 5 },
          makePlayer({ id: 1, capturedCount: 3 }),
          makePlayer({ id: 2, capturedCount: 4 }),
          makePlayer({ id: 3, capturedCount: 1 }),
        ],
      }),
    );
    renderWithProviders(<RistikontraPage />);
    // 席 0 と席 2 はチーム 0 = 5+4 = 9。席 1 と席 3 はチーム 1 = 3+1 = 4。
    await waitFor(() => expect(screen.getByTestId('ristikontra-provisional-0')).toHaveTextContent('9'));
    expect(screen.getByTestId('ristikontra-provisional-2')).toHaveTextContent('9');
    expect(screen.getByTestId('ristikontra-provisional-1')).toHaveTextContent('4');
    expect(screen.getByTestId('ristikontra-provisional-3')).toHaveTextContent('4');
  });

  it('marks the leading team, and neither team when they are level', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          { ...playState.players[0], capturedCount: 4 },
          makePlayer({ id: 1, capturedCount: 4 }),
          makePlayer({ id: 2, capturedCount: 4 }),
          makePlayer({ id: 3, capturedCount: 4 }),
        ],
      }),
    );
    renderWithProviders(<RistikontraPage />);
    // 8 対 8 なので、どちらにも星は付かない。
    await waitFor(() => expect(screen.getByTestId('ristikontra-provisional-0')).toHaveTextContent('8'));
    for (const seat of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`ristikontra-provisional-${seat}`)).not.toHaveTextContent('★');
    }
  });

  it('hides the provisional readout and shows the final score on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<RistikontraPage />);
    await waitFor(() => expect(screen.getByText(/11点/)).toBeInTheDocument());
    expect(screen.queryByTestId('ristikontra-provisional-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ristikontra-provisional-note')).not.toBeInTheDocument();
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<RistikontraPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
