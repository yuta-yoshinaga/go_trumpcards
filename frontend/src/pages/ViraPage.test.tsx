import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { viraApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeViraState } from '../test/stateFactories';
import { ViraPage } from './ViraPage';

vi.mock('../api/gameApi', () => ({
  viraApi: { exec: vi.fn() },
  actionLogApi: { vira: vi.fn() },
}));

const mockExec = vi.mocked(viraApi.exec);

// Default fixture: a human bid turn (bid phase).
const bidPhaseState = makeViraState();
// A human play turn with the human as declarer (so the play control is shown).
const playPhaseState = makeViraState({
  phase: 1,
  declarerIdx: 0,
  contract: 1,
  trumpSuit: 3,
  isHumanBidTurn: false,
  isHumanTurn: true,
  playableIndices: [0, 1, 2],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
});
const cpuTurnState = makeViraState({
  phase: 1,
  declarerIdx: 1,
  isHumanBidTurn: false,
  isHumanTurn: false,
  currentPlayerIdx: 1,
});
const trickEndState = makeViraState({ phase: 2, isHumanBidTurn: false });
const roundEndState = makeViraState({ phase: 3, isHumanBidTurn: false, roundTricks: [6, 2, 2] });
const gameEndState = makeViraState({
  phase: 4,
  isHumanBidTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidPhaseState);
});

describe('ViraPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ViraPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<ViraPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 30 },
      }),
    );
  });

  it('shows bid buttons on a human bid turn', async () => {
    renderWithProviders(<ViraPage />);
    await waitFor(() => expect(screen.getByTestId('bid-0')).toBeInTheDocument());
    expect(screen.getByTestId('bid-1')).toBeInTheDocument();
    expect(screen.getByTestId('bid-2')).toBeInTheDocument();
    expect(screen.getByTestId('bid-3')).toBeInTheDocument();
    expect(screen.getByTestId('bid-4')).toBeInTheDocument();
  });

  it('dispatches a bid when a bid button is clicked', async () => {
    renderWithProviders(<ViraPage />);
    const bidSix = await screen.findByTestId('bid-1');
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bidSix);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1 }));
  });

  it('disables bids that do not beat the current highest bid', async () => {
    mockExec.mockResolvedValue(makeViraState({ bids: [2, 0, 0] }));
    renderWithProviders(<ViraPage />);
    // Highest bid is 2 (Misère): Six (1) and Misère (2) are disabled; Seven (3), Eight (4), Pass (0) enabled.
    await waitFor(() => expect(screen.getByTestId('bid-1')).toBeDisabled());
    expect(screen.getByTestId('bid-2')).toBeDisabled();
    expect(screen.getByTestId('bid-3')).toBeEnabled();
    expect(screen.getByTestId('bid-4')).toBeEnabled();
    expect(screen.getByTestId('bid-0')).toBeEnabled();
  });

  it('shows the current highest bid (none when no one has bid)', async () => {
    renderWithProviders(<ViraPage />);
    const banner = await screen.findByTestId('vira-highest-bid');
    expect(banner).toHaveTextContent('なし');
  });

  it('names the current highest bid once someone has bid', async () => {
    mockExec.mockResolvedValue(makeViraState({ bids: [2, 0, 0] }));
    renderWithProviders(<ViraPage />);
    const banner = await screen.findByTestId('vira-highest-bid');
    expect(banner).toHaveTextContent('ミゼール');
  });

  it('explains via tooltip and aria-label why a too-low bid is disabled', async () => {
    // 席 0 が Solo(2) を出している。Gask(1) はそれより下なので押せない。
    mockExec.mockResolvedValue(makeViraState({ bids: [2, 0, 0] }));
    renderWithProviders(<ViraPage />);
    const gask = await screen.findByTestId('bid-1');
    expect(gask.closest('span')).toHaveAttribute('title', '現在の最高入札を上回る必要があります');
    // The button name itself carries the reason for SR users.
    expect(gask).toHaveAttribute('aria-label', 'ガスク (7) — 現在の最高入札を上回る必要があります');
  });

  it('marks the Misère bid as a special contract with a described risk', async () => {
    renderWithProviders(<ViraPage />);
    const misere = await screen.findByTestId('bid-3');
    expect(misere).toHaveTextContent('特殊');
    // The risk explanation is available to screen readers via aria-describedby.
    expect(misere).toHaveAttribute('aria-describedby', 'vira-misere-desc');
    expect(document.getElementById('vira-misere-desc')).toHaveTextContent(/ミゼール/);
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<ViraPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('宣言者')).toBeInTheDocument();
  });

  it('does not show the play button or bid buttons on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ViraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<ViraPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ViraPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ViraPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  // A play-phase state where CPU 1 is the declarer with the given contract and trick progress.
  const declarerProgressState = (contract: number, tricksWon: number, cardsLeft: number) =>
    makeViraState({
      phase: 1,
      declarerIdx: 1,
      contract,
      trumpSuit: 3,
      isHumanBidTurn: false,
      isHumanTurn: false,
      currentPlayerIdx: 0,
      players: [
        { id: 0, isHuman: true, cardCount: cardsLeft, cards: [], trickCount: 0, score: 0, isDeclarer: false },
        {
          id: 1,
          isHuman: false,
          cardCount: cardsLeft,
          cards: [],
          trickCount: tricksWon,
          score: 0,
          isDeclarer: true,
        },
        { id: 2, isHuman: false, cardCount: cardsLeft, cards: [], trickCount: 0, score: 0, isDeclarer: false },
      ],
    });

  it('shows in-progress contract progress with the target', async () => {
    // Six (needs 6), 4 won with 3 tricks left → reachable, still in progress.
    mockExec.mockResolvedValue(declarerProgressState(1, 4, 3));
    renderWithProviders(<ViraPage />);
    const readout = await screen.findByTestId('vira-contract-progress');
    expect(readout).toHaveTextContent('4 / 6');
    expect(readout).toHaveClass('text-ds-warning');
    expect(readout).not.toHaveTextContent('達成');
    expect(readout).not.toHaveTextContent('失敗確定');
  });

  it('shows the contract as made once the target is reached', async () => {
    // Six (needs 6), 6 won with 0 left → made.
    mockExec.mockResolvedValue(declarerProgressState(1, 6, 0));
    renderWithProviders(<ViraPage />);
    const readout = await screen.findByTestId('vira-contract-progress');
    expect(readout).toHaveTextContent('6 / 6');
    expect(readout).toHaveTextContent('達成');
    expect(readout).toHaveClass('text-ds-success');
  });

  it('shows failure once the target becomes unreachable', async () => {
    // Eight (needs 8), 2 won with 3 left → 2+3 < 8, failure certain.
    mockExec.mockResolvedValue(declarerProgressState(4, 2, 3));
    renderWithProviders(<ViraPage />);
    const readout = await screen.findByTestId('vira-contract-progress');
    expect(readout).toHaveTextContent('2 / 8');
    expect(readout).toHaveTextContent('失敗確定');
    expect(readout).toHaveClass('text-ds-error');
  });

  it('flags Misère failure the instant the declarer wins a trick', async () => {
    // Misère (target 0), 1 trick won → immediate failure.
    // **Vira では Misère は 3。**Préférence の 2 から位置が変わっている。
    mockExec.mockResolvedValue(declarerProgressState(3, 1, 5));
    renderWithProviders(<ViraPage />);
    const readout = await screen.findByTestId('vira-contract-progress');
    expect(readout).toHaveTextContent('ミゼール');
    expect(readout).toHaveTextContent('失敗確定');
    expect(readout).toHaveClass('text-ds-error');
  });

  it('does not show contract progress before a declarer is decided', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<ViraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('vira-contract-progress')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<ViraPage />);
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
      messageCode: 'vira.hintRequested',
    });
    renderWithProviders(<ViraPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **ハンドラは押して初めて走る。**描画だけを見るテストでは `useViraGame` の
  // exec 呼び出しが一度も通らず、カバレッジに穴が残る（codecov が #4660 で
  // 指摘したのがこれ）。
  it('plays the selected card', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<ViraPage />);
    // 手札は `PlayerHandSection` が `cardAlt` を aria-label にして描く。
    await screen.findByRole('button', { name: '出す' });
    const hand = screen.getAllByRole('button').filter((b) => /^[♠♣♥♦] /.test(b.getAttribute('aria-label') ?? ''));
    expect(hand.length).toBeGreaterThan(0);
    fireEvent.click(hand[0]);

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', expect.anything()));
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<ViraPage />);
    const next = await screen.findByRole('button', { name: '次のトリック' });
    mockExec.mockClear();
    fireEvent.click(next);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ViraPage />);
    const next = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(next);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('does nothing when no card is selected', async () => {
    // `handlePlay` は選択が 1 枚でなければ何もしない。押しても exec が飛ばない
    // ことを踏まないと、この早期 return が未カバーのまま残る。
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<ViraPage />);
    const play = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    fireEvent.click(play);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
