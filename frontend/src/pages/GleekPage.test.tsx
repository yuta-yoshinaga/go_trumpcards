import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gleekApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGleekState } from '../test/stateFactories';
import { GleekPhase } from '../types/phases';
import { GleekPage } from './GleekPage';

vi.mock('../api/gameApi', () => ({
  gleekApi: { exec: vi.fn() },
  actionLogApi: { gleek: vi.fn() },
}));

const mockExec = vi.mocked(gleekApi.exec);

const playPhaseState = makeGleekState();
const bidPhaseState = makeGleekState({
  phase: GleekPhase.BID,
  currentBidderIdx: 0,
  isHumanTurn: false,
  isHumanBidTurn: true,
  buyerIdx: -1,
  ruffWinnerIdx: -1,
  winningBid: 0,
  highestBid: 12,
  nextBidAmount: 14,
});
const discardPhaseState = makeGleekState({
  phase: GleekPhase.DISCARD,
  isHumanTurn: false,
  isHumanDiscardTurn: true,
  ruffWinnerIdx: -1,
  discardCount: 2,
  playableIndices: [],
});
const trickEndState = makeGleekState({
  phase: GleekPhase.TRICK_END,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeGleekState({ phase: GleekPhase.ROUND_END, dealPoints: 78, par: 26 });
const gameEndState = makeGleekState({
  phase: GleekPhase.GAME_END,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeGleekState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('GleekPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GleekPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<GleekPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the buyer badge', async () => {
    renderWithProviders(<GleekPage />);
    await waitFor(() => expect(screen.getByAltText('♥ A')).toBeInTheDocument());
    expect(screen.getByText('落札者')).toBeInTheDocument();
  });

  // **段階の点は出さないと見えない。** ラフとメルドで動いた点が画面に無いと、
  // 累積点だけが理由なく動いているように見える。
  it('shows the stock, the ruff and both meld kinds', async () => {
    mockExec.mockResolvedValue(
      makeGleekState({
        melds: [
          { playerIdx: 0, rank: 13, count: 3, value: 3 },
          { playerIdx: 1, rank: 11, count: 4, value: 2 },
        ],
      }),
    );
    renderWithProviders(<GleekPage />);
    const stage = await screen.findByTestId('gleek-stage-line');
    expect(stage).toHaveTextContent('14 で落札');
    expect(screen.getByTestId('gleek-ruff-line')).toHaveTextContent('ハート');
    const melds = screen.getAllByTestId('gleek-meld-line');
    expect(melds).toHaveLength(2);
    expect(melds[0]).toHaveTextContent('グリーク');
    expect(melds[1]).toHaveTextContent('マーニヴァル');
  });

  it('omits the ruff line before the ruff is scored', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GleekPage />);
    await screen.findByTestId('gleek-bid-controls');
    expect(screen.queryByTestId('gleek-ruff-line')).not.toBeInTheDocument();
  });

  it('renders the bid phase and dispatches a raise of exactly the next amount', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GleekPage />);
    const raise = await screen.findByRole('button', { name: '14 まで競り上げる' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(raise);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 14 }));
  });

  it('dropping out dispatches bid 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GleekPage />);
    const drop = await screen.findByRole('button', { name: '降りる' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(drop);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0 }));
  });

  // **サーバが弾く選択肢は出さない。** 上限に達した卓では 0 が返るので、
  // 押しても必ず失敗するボタンを出してはいけない。
  it('hides the raise button once the auction has hit its ceiling', async () => {
    mockExec.mockResolvedValue(makeGleekState({ ...bidPhaseState, nextBidAmount: 0 }));
    renderWithProviders(<GleekPage />);
    await screen.findByTestId('gleek-bid-controls');
    expect(screen.queryByRole('button', { name: /競り上げる/ })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument();
  });

  // **捨て札はちょうど指定枚数。** 足りないまま送れると、サーバに弾かれるだけの
  // ボタンになる。
  it('enables the discard button only once exactly the required cards are selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GleekPage />);
    const confirm = await screen.findByTestId('gleek-discard-confirm');
    expect(confirm).toBeDisabled();

    fireEvent.click(screen.getByAltText('♥ A'));
    expect(screen.getByTestId('gleek-discard-confirm')).toBeDisabled();

    fireEvent.click(screen.getByAltText('♥ J'));
    expect(screen.getByTestId('gleek-discard-confirm')).toBeEnabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByTestId('gleek-discard-confirm'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [0, 1] }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<GleekPage />);
    const card = await screen.findByAltText('♥ A');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<GleekPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  // **基準点はそのディールから数える。** 上限を出すと、名札が場外に落ちた
  // ディールで説明が合わなくなる。
  it('renders the deal total and the par it is settled against', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GleekPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    const par = screen.getByTestId('gleek-par-line');
    expect(par).toHaveTextContent('78');
    expect(par).toHaveTextContent('26');
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GleekPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GleekPage />);
    await waitFor(() => expect(screen.getByAltText('♥ A')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回ヒントを
  // 載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'lead_high' } });
    renderWithProviders(<GleekPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'lead_high' },
      messageCode: 'gleek.hintRequested',
    });
    renderWithProviders(<GleekPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
