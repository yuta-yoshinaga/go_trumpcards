import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { napApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeNapState } from '../test/stateFactories';
import { NapPage } from './NapPage';

vi.mock('../api/gameApi', () => ({
  napApi: { exec: vi.fn() },
  actionLogApi: { nap: vi.fn() },
}));

const mockExec = vi.mocked(napApi.exec);

// Default fixture: a human bid turn (bid phase).
const bidPhaseState = makeNapState();
// A human play turn with a started trick (so the play control is shown).
const playPhaseState = makeNapState({
  phase: 1,
  declarerIdx: 0,
  contract: 3,
  trumpSuit: 3,
  isHumanBidTurn: false,
  isHumanTurn: true,
  playableIndices: [0, 1, 2],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
});
const cpuTurnState = makeNapState({
  phase: 1,
  declarerIdx: 1,
  isHumanBidTurn: false,
  isHumanTurn: false,
  currentPlayerIdx: 1,
});
const trickEndState = makeNapState({ phase: 2, isHumanBidTurn: false });
const roundEndState = makeNapState({ phase: 3, isHumanBidTurn: false, roundTricks: [3, 1, 1, 0] });
const gameEndState = makeNapState({
  phase: 4,
  isHumanBidTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidPhaseState);
});

describe('NapPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<NapPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<NapPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 20 },
      }),
    );
  });

  it('shows bid buttons on a human bid turn', async () => {
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(screen.getByTestId('bid-0')).toBeInTheDocument());
    expect(screen.getByTestId('bid-2')).toBeInTheDocument();
    expect(screen.getByTestId('bid-3')).toBeInTheDocument();
    expect(screen.getByTestId('bid-4')).toBeInTheDocument();
    expect(screen.getByTestId('bid-5')).toBeInTheDocument();
  });

  // #5651: Nap 契約だけ賭け金が非対称 (達成 +10 / 失敗は相手が各 +5)。他は契約数
  // と同じ数が動く。この差が宣言するかどうかの判断そのものなのに、ボタンには
  // 契約名しか出ていなかった。
  it('puts each contract stake on its bid button', async () => {
    renderWithProviders(<NapPage />);

    expect(await screen.findByTestId('bid-2')).toHaveTextContent('+2/+2');
    expect(screen.getByTestId('bid-3')).toHaveTextContent('+3/+3');
    expect(screen.getByTestId('bid-4')).toHaveTextContent('+4/+4');
    // ナップだけ非対称。
    expect(screen.getByTestId('bid-5')).toHaveTextContent('+10/+5');
    // パスは賭けないので数字を出さない。
    expect(screen.getByTestId('bid-0')).not.toHaveTextContent('+');
  });

  it('reports the chips the round actually moved', async () => {
    mockExec.mockResolvedValue(
      makeNapState({ ...roundEndState, declarerIdx: 0, contract: 5, roundTricks: [5, 0, 0, 0] }),
    );
    renderWithProviders(<NapPage />);

    const panel = await screen.findByTestId('nap-round-payout');
    // 5トリック全取り = ナップ達成。宣言者が +10。
    expect(panel).toHaveTextContent('10');
  });

  // **失敗時に減るのは宣言者ではなく、増えるのが相手。**「-5」と書くと嘘になる。
  it('says the opponents gain when the contract fails', async () => {
    mockExec.mockResolvedValue(
      makeNapState({ ...roundEndState, declarerIdx: 0, contract: 5, roundTricks: [3, 1, 1, 0] }),
    );
    renderWithProviders(<NapPage />);

    const panel = await screen.findByTestId('nap-round-payout');
    expect(panel).toHaveTextContent('5');
  });

  it('dispatches a bid when a bid button is clicked', async () => {
    renderWithProviders(<NapPage />);
    const bidTwo = await screen.findByTestId('bid-2');
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bidTwo);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 2 }));
  });

  it('disables bids that do not beat the current highest bid', async () => {
    mockExec.mockResolvedValue(makeNapState({ bids: [3, 0, 0, 0] }));
    renderWithProviders(<NapPage />);
    // Highest bid is 3 (Three): Two (2) and Three (3) are disabled, Four (4) and Nap (5) and Pass (0) enabled.
    await waitFor(() => expect(screen.getByTestId('bid-2')).toBeDisabled());
    expect(screen.getByTestId('bid-3')).toBeDisabled();
    expect(screen.getByTestId('bid-4')).toBeEnabled();
    expect(screen.getByTestId('bid-5')).toBeEnabled();
    expect(screen.getByTestId('bid-0')).toBeEnabled();
  });

  it('shows the current highest bid and a too-low tooltip on disabled bids', async () => {
    mockExec.mockResolvedValue(makeNapState({ bids: [3, 0, 0, 0] }));
    renderWithProviders(<NapPage />);
    const info = await screen.findByTestId('nap-highest-bid');
    // A real highest bid is shown (not the "no bids yet" placeholder).
    expect(info).not.toHaveTextContent('まだ入札なし');
    // Too-low bids carry a tooltip on the wrapping span (browsers suppress tooltips on
    // disabled buttons); valid bids do not.
    const tooLowWrap = screen.getByTestId('bid-wrap-2');
    expect(tooLowWrap).toHaveAttribute('title');
    expect(tooLowWrap.getAttribute('title')).not.toBe('');
    expect(screen.getByTestId('bid-wrap-4')).not.toHaveAttribute('title');
    // The reason is also exposed to screen readers via aria-label on the disabled button.
    expect(screen.getByTestId('bid-2')).toHaveAttribute('aria-label', expect.stringContaining(tooLowWrap.title));
    expect(screen.getByTestId('bid-4')).not.toHaveAttribute('aria-label');
  });

  it('shows "no bids yet" before anyone has bid', async () => {
    mockExec.mockResolvedValue(makeNapState({ bids: [0, 0, 0, 0] }));
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(screen.getByTestId('nap-highest-bid')).toHaveTextContent('まだ入札なし'));
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<NapPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('宣言者')).toBeInTheDocument();
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  it('shows the declarer trick progress toward the contract during play', async () => {
    // contract 3, declarer won 0, all 5 tricks still to play.
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<NapPage />);
    const progress = await screen.findByTestId('nap-declarer-progress');
    expect(progress).toHaveTextContent('宣言者: 0 / 3 トリック（残り5トリック）');
    expect(progress).toHaveAttribute('role', 'status');
    expect(progress).not.toHaveTextContent('達成不可');
    expect(progress.className).not.toContain('text-ds-error');
  });

  it('marks the contract as unreachable once too few tricks remain', async () => {
    // contract 4, declarer has 0 tricks, opponents have taken 2 → only 3 remain, so 4 is impossible.
    const unreachableState = makeNapState({
      phase: 1,
      declarerIdx: 0,
      contract: 4,
      isHumanBidTurn: false,
      isHumanTurn: true,
      players: [
        { id: 0, isHuman: true, cardCount: 3, cards: [], trickCount: 0, score: 0, isDeclarer: true },
        { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 1, score: 0, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 3, cards: [], trickCount: 1, score: 0, isDeclarer: false },
        { id: 3, isHuman: false, cardCount: 3, cards: [], trickCount: 0, score: 0, isDeclarer: false },
      ],
    });
    mockExec.mockResolvedValue(unreachableState);
    renderWithProviders(<NapPage />);
    const progress = await screen.findByTestId('nap-declarer-progress');
    expect(progress).toHaveTextContent('達成不可');
    expect(progress.className).toContain('text-ds-error');
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<NapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...bidPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'nap.hintRequested',
    });
    renderWithProviders(<NapPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
