import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ultiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeUltiState } from '../test/stateFactories';
import { UltiPage } from './UltiPage';

vi.mock('../api/gameApi', () => ({
  ultiApi: { exec: vi.fn() },
  actionLogApi: { ulti: vi.fn() },
}));

const mockExec = vi.mocked(ultiApi.exec);

const playPhaseState = makeUltiState();
const bidPhaseState = makeUltiState({
  phase: 0,
  isHumanTurn: true,
  isHumanBidTurn: true,
  contract: 0,
  trumpSuit: -1,
  talonTaken: false,
  playableIndices: [],
});
const discardPhaseState = makeUltiState({
  phase: 1,
  isHumanTurn: true,
  isHumanBidTurn: false,
  contract: 1,
  trumpSuit: 1,
  playableIndices: [],
});
const trickEndState = makeUltiState({
  phase: 3,
  isHumanTurn: false,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeUltiState({
  phase: 4,
  isHumanTurn: false,
  outcome: 1,
});
const gameEndState = makeUltiState({
  phase: 5,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeUltiState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('UltiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<UltiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<UltiPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    renderWithProviders(<UltiPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the declarer — the badge renders.
    expect(screen.getAllByText('デクレアラー').length).toBeGreaterThan(0);
  });

  it('renders the bid phase with Party, Ulti, Betli and Durchmarsch buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パルティ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ウルティ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ベトリ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ドゥルマルス' })).toBeInTheDocument();
  });

  it('Ulti is disabled until a trump suit is picked, then dispatches bid with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    const ultiBtn = await screen.findByRole('button', { name: 'ウルティ' });
    expect(ultiBtn).toBeDisabled();
    // Pick hearts (♥) as trump.
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    expect(screen.getByRole('button', { name: 'ウルティ' })).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ウルティ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'ulti', trumpSuit: 3 }));
  });

  it('Party is disabled until a trump suit is picked, then dispatches bid with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    const partyBtn = await screen.findByRole('button', { name: 'パルティ' });
    expect(partyBtn).toBeDisabled();
    // Pick spades (♠) as trump.
    fireEvent.click(screen.getByRole('button', { name: 'スペード' }));
    expect(screen.getByRole('button', { name: 'パルティ' })).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'パルティ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'party', trumpSuit: 1 }));
  });

  it('declaring Betli dispatches bid with no trump requirement', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    const betliBtn = await screen.findByRole('button', { name: 'ベトリ' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(betliBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'betli', trumpSuit: undefined }));
  });

  it('does not show bid controls on a CPU/non-bid turn', async () => {
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'パルティ' })).not.toBeInTheDocument();
  });

  it('renders the discard phase and dispatches discard for the two selected cards', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<UltiPage />);
    const first = await screen.findByAltText('♥ Q');
    fireEvent.click(first);
    fireEvent.click(screen.getByAltText('♥ K'));
    const discardBtn = screen.getByRole('button', { name: '捨てる' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 1] }));
  });

  it('shows the discard selection count that updates as cards are picked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<UltiPage />);
    const progress = await screen.findByTestId('ulti-discard-progress');
    // Starts at 0 of the required 2.
    expect(progress).toHaveTextContent('0/2');
    fireEvent.click(screen.getByAltText('♥ Q'));
    expect(screen.getByTestId('ulti-discard-progress')).toHaveTextContent('1/2');
    fireEvent.click(screen.getByAltText('♥ K'));
    expect(screen.getByTestId('ulti-discard-progress')).toHaveTextContent('2/2');
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<UltiPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('shows signed coin deltas and announces the round result on settlement', async () => {
    // 精算額はサーバが lastDealCoins で返す。累積 coins は前ディールぶんを含むので、
    // 「累積 -1 だが今回は +2 動いた」形にしてどちらを読んでいるかを確かめる。
    const settledState = makeUltiState({
      phase: 4,
      isHumanTurn: false,
      outcome: 1,
      lastDealCoins: [2, -1, -1],
      players: [
        { id: 0, isHuman: true, cardCount: 0, cards: [], trickCount: 0, cardPoints: 0, coins: 9, isDeclarer: true },
        { id: 1, isHuman: false, cardCount: 0, cards: [], trickCount: 0, cardPoints: 0, coins: -4, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 0, cards: [], trickCount: 0, cardPoints: 0, coins: -5, isDeclarer: false },
      ],
    });
    mockExec.mockResolvedValueOnce(trickEndState); // mount → trick end
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(settledState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));

    await waitFor(() => expect(screen.getByTestId('ulti-coin-delta-0')).toHaveTextContent('+2'));
    expect(screen.getByTestId('ulti-coin-delta-1')).toHaveTextContent('-1');

    const panel = screen.getByTestId('ulti-round-result');
    expect(panel).toHaveAttribute('role', 'status');
    expect(panel).toHaveAttribute('aria-live', 'polite');
    expect(panel).toHaveTextContent('あなた: +2コイン');
  });

  it('shows coin deltas on the final round even though the backend jumps to GAME_END', async () => {
    // The match-deciding round settles straight into GAME_END (no ROUND_END).
    const finalSettle = makeUltiState({
      phase: 5,
      lastDealCoins: [3, -3, 0],
      isHumanTurn: false,
      gameEndFlag: true,
      outcome: 1,
      winnerPlayer: 0,
      players: [
        { id: 0, isHuman: true, cardCount: 0, cards: [], trickCount: 0, cardPoints: 0, coins: 12, isDeclarer: true },
        { id: 1, isHuman: false, cardCount: 0, cards: [], trickCount: 0, cardPoints: 0, coins: -6, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 0, cards: [], trickCount: 0, cardPoints: 0, coins: 0, isDeclarer: false },
      ],
    });
    mockExec.mockResolvedValueOnce(trickEndState); // mount → trick end (coins 0)
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(finalSettle);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));

    await waitFor(() => expect(screen.getByTestId('ulti-coin-delta-0')).toHaveTextContent('+3'));
    expect(screen.getByTestId('ulti-coin-delta-1')).toHaveTextContent('-3');
  });

  it('does not show coin deltas outside the settlement phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByTestId('ulti-coin-delta-0')).not.toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<UltiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'ulti.hintRequested',
    });
    renderWithProviders(<UltiPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('states each contract win condition, including Ulti', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<UltiPage />);
    const betli = await screen.findByRole('button', { name: 'ベトリ' });
    expect(betli).toHaveAttribute('title', expect.stringContaining('1トリックも取らなければ'));
    // The contract the game is named after had no explanation on the web at all.
    expect(document.getElementById('ulti-bid-desc-ulti')?.textContent).toMatch(/切り札の7/);
  });
});
