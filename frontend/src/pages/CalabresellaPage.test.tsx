import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { calabresellaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCalabresellaState } from '../test/stateFactories';
import { CalabresellaPage } from './CalabresellaPage';

vi.mock('../api/gameApi', () => ({
  calabresellaApi: { exec: vi.fn() },
  actionLogApi: { calabresella: vi.fn() },
}));

const mockExec = vi.mocked(calabresellaApi.exec);

const playPhaseState = makeCalabresellaState();
const bidPhaseState = makeCalabresellaState({
  phase: 0,
  currentBidderIdx: 0,
  isHumanTurn: true,
  winningBid: 0,
  highestBid: 0,
});
// A realistic discard phase: the soloist holds 16 cards (took the 4-card monte) and must
// discard down to the regulation 12, so 4 discards remain. Keeps ♥Q for selection tests.
const soloist16CardHand = [
  { design: 'HEART' as const, value: 12 },
  { design: 'HEART' as const, value: 13 },
  { design: 'SPADE' as const, value: 1 },
  { design: 'SPADE' as const, value: 2 },
  { design: 'SPADE' as const, value: 3 },
  { design: 'SPADE' as const, value: 4 },
  { design: 'SPADE' as const, value: 5 },
  { design: 'SPADE' as const, value: 6 },
  { design: 'SPADE' as const, value: 7 },
  { design: 'CLOVER' as const, value: 1 },
  { design: 'CLOVER' as const, value: 2 },
  { design: 'CLOVER' as const, value: 3 },
  { design: 'CLOVER' as const, value: 4 },
  { design: 'CLOVER' as const, value: 5 },
  { design: 'CLOVER' as const, value: 6 },
  { design: 'CLOVER' as const, value: 7 },
];
const discardPhaseState = makeCalabresellaState({
  phase: 1,
  soloistIdx: 0,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 16,
      cards: soloist16CardHand,
      trickCount: 0,
      score: 0,
      isSoloist: true,
      roundThirds: 0,
    },
    { id: 1, isHuman: false, cardCount: 12, cards: [], trickCount: 0, score: 0, isSoloist: false, roundThirds: 0 },
    { id: 2, isHuman: false, cardCount: 12, cards: [], trickCount: 0, score: 0, isSoloist: false, roundThirds: 0 },
  ],
});
const trickEndState = makeCalabresellaState({
  phase: 3,
  lastTrickWinner: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeCalabresellaState({
  phase: 4,
  roundThirds: [20, 8, 5],
  roundScoreChanges: [2, -1, 0],
  soloistWon: true,
});
const gameEndState = makeCalabresellaState({
  phase: 5,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeCalabresellaState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CalabresellaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CalabresellaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 21 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Soloist badge', async () => {
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default Soloist.
    expect(screen.getByText('ソリスト')).toBeInTheDocument();
  });

  it('shows the lead suit while the current trick has a card', async () => {
    mockExec.mockResolvedValueOnce(
      makeCalabresellaState({ currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 12 } }] }),
    );
    renderWithProviders(<CalabresellaPage />);
    expect(await screen.findByText('リードスート: ハート')).toBeInTheDocument();
  });

  it('does not show a lead suit when the current trick is empty', async () => {
    mockExec.mockResolvedValueOnce(makeCalabresellaState({ currentTrick: [] }));
    renderWithProviders(<CalabresellaPage />);
    await screen.findByText('あなたのターン');
    expect(screen.queryByTestId('trick-lead-suit')).not.toBeInTheDocument();
  });

  it('renders the bid phase with chiamo, solo and pass buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'キアーモ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ソロ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('marks bids at or below the winning bid unavailable while keeping pass available', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 1 });
    renderWithProviders(<CalabresellaPage />);
    const chiamo = await screen.findByRole('button', { name: 'キアーモ' });
    const solo = screen.getByRole('button', { name: 'ソロ' });
    const pass = screen.getByRole('button', { name: 'パス' });
    expect(chiamo).toHaveAttribute('aria-disabled', 'true');
    expect(chiamo).toHaveAttribute('aria-describedby', 'calabresella-bid-unavailable-reason');
    expect(solo).not.toHaveAttribute('aria-disabled', 'true');
    expect(pass).not.toHaveAttribute('aria-disabled', 'true');
    mockExec.mockClear();
    fireEvent.click(chiamo);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(solo);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 2 }));
  });

  it('makes both bids unavailable at solo while keeping pass available', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 2 });
    renderWithProviders(<CalabresellaPage />);
    expect(await screen.findByRole('button', { name: 'キアーモ' })).toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByRole('button', { name: 'ソロ' })).toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByRole('button', { name: 'パス' })).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('allows both bids when no declaration has been made', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 0 });
    renderWithProviders(<CalabresellaPage />);
    expect(await screen.findByRole('button', { name: 'キアーモ' })).not.toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByRole('button', { name: 'ソロ' })).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('declaring chiamo dispatches bid with bid=1', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);
    const chiamoBtn = await screen.findByRole('button', { name: 'キアーモ' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(chiamoBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1 }));
  });

  it('renders the discard phase with the discard-card button and prompt', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByTestId('calabresella-discard-prompt')).toBeInTheDocument());
    expect(screen.getByTestId('calabresella-discard-button')).toBeInTheDocument();
  });

  it('reveals the monte (widow) cards once the Soloist has taken them', async () => {
    mockExec.mockResolvedValue({
      ...discardPhaseState,
      monte: [
        { design: 'DIAMOND', value: 3 },
        { design: 'SPADE', value: 11 },
        { design: 'CLOVER', value: 7 },
        { design: 'HEART', value: 2 },
      ],
    });
    renderWithProviders(<CalabresellaPage />);
    const monte = await screen.findByTestId('calabresella-monte');
    expect(monte).toBeInTheDocument();
    expect(monte).toHaveTextContent('モンテ');
    // All four widow cards are rendered as face-up images inside the monte row.
    expect(monte.querySelectorAll('img')).toHaveLength(4);
  });

  it('does not render the monte row during the bid phase (widow not yet taken)', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'キアーモ' })).toBeInTheDocument());
    expect(screen.queryByTestId('calabresella-monte')).not.toBeInTheDocument();
  });

  it('shows the remaining discard count in the prompt and on the button', async () => {
    // 16-card soloist hand → 4 discards remain before reaching the regulation 12.
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);
    const prompt = await screen.findByTestId('calabresella-discard-prompt');
    expect(prompt).toHaveTextContent('残り 4 枚');
    expect(screen.getByTestId('calabresella-discard-button')).toHaveTextContent('4枚を捨てる');

    fireEvent.click(await screen.findByAltText('♥ Q'));
    fireEvent.click(await screen.findByAltText('♥ K'));
    expect(prompt).toHaveTextContent('残り 2 枚');

    fireEvent.click(await screen.findByAltText('♠ A'));
    fireEvent.click(await screen.findByAltText('♠ 2'));
    expect(prompt).toHaveTextContent('残り 0 枚');
  });

  it('hides the discard prompt once the hand is down to the regulation 12', async () => {
    mockExec.mockResolvedValue({
      ...discardPhaseState,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 12,
          cards: soloist16CardHand.slice(0, 12),
          trickCount: 0,
          score: 0,
          isSoloist: true,
          roundThirds: 0,
        },
        ...discardPhaseState.players.slice(1),
      ],
    });
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByTestId('calabresella-discard-button')).toBeInTheDocument());
    expect(screen.queryByTestId('calabresella-discard-prompt')).not.toBeInTheDocument();
    // With nothing left to discard the button no longer carries a count suffix.
    expect(screen.getByTestId('calabresella-discard-button')).toHaveTextContent('カードを捨てる');
    expect(screen.getByTestId('calabresella-discard-button').textContent).not.toContain('(');
  });

  it('selecting four cards then discarding dispatches them together', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    fireEvent.click(await screen.findByAltText('♥ K'));
    fireEvent.click(await screen.findByAltText('♠ A'));
    fireEvent.click(await screen.findByAltText('♠ 2'));
    const discardBtn = await screen.findByTestId('calabresella-discard-button');
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 1, 2, 3] }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<CalabresellaPage />);
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
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('CPU 1 が獲得');
  });

  it('does not render a winner badge while a trick is in progress', async () => {
    mockExec.mockResolvedValue(
      makeCalabresellaState({
        phase: 2,
        lastTrickWinner: 1,
        currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 12 } }],
      }),
    );
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByTestId('trick-display-cards')).toBeInTheDocument());
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
    expect(screen.getByText('ソリストの契約達成')).toBeInTheDocument();
    expect(screen.getByText('あなた 今回の精算: +2点')).toBeInTheDocument();
    expect(screen.getByText('CPU 1 今回の精算: -1点')).toBeInTheDocument();
    expect(screen.getByText('CPU 2 今回の精算: ±0点')).toBeInTheDocument();
  });

  it('shows the round result on mobile and game end layouts', async () => {
    const originalWidth = window.innerWidth;
    try {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      mockExec.mockResolvedValue({
        ...gameEndState,
        soloistWon: true,
        roundThirds: [20, 8, 5],
        roundScoreChanges: [2, -1, 0],
      });
      renderWithProviders(<CalabresellaPage />);
      expect(await screen.findByText('ソリストの契約達成')).toBeInTheDocument();
      expect(screen.getByText('あなた 今回の精算: +2点')).toBeInTheDocument();
      expect(screen.getByText('CPU 1 今回の精算: -1点')).toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('does not show a round result during play', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, soloistWon: true, roundScoreChanges: [2, -1, 0] });
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByText('ソリストの契約達成')).not.toBeInTheDocument();
    expect(screen.queryByText(/今回の精算/)).not.toBeInTheDocument();
  });

  it('shows when the Soloist misses the contract', async () => {
    mockExec.mockResolvedValue(
      makeCalabresellaState({
        phase: 4,
        soloistWon: false,
        roundScoreChanges: [-4, 2, 2],
      }),
    );
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByText('ソリストの契約未達')).toBeInTheDocument());
    expect(screen.getByText('あなた 今回の精算: -4点')).toBeInTheDocument();
    expect(screen.getByText('CPU 1 今回の精算: +2点')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<CalabresellaPage />);
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
      // このページのバナーは `cardIndices` を並べる。`cardIndex` を渡しても
      // 何も出ないので、押した側のテストにならない。
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'calabresella.hintRequested',
    });
    renderWithProviders(<CalabresellaPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **催促は常設のライブ領域から出す** ── 出現と同時に付けた領域は変化として
  // 扱われず読み上げられない (#5955)。同じファイルの hint-live は既にこの形
  // だったのに、ビッド/捨て札の催促だけ素の div のままだった (#6484)。
  it('keeps the prompt live region mounted and announces the bid prompt', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);

    const live = await screen.findByTestId('calabresella-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 催促は**その領域の中**にある ── 隣に置いただけでは読み上げられない。
    expect(live).toContainElement(screen.getByTestId('calabresella-bid-prompt'));
  });

  it('announces the discard prompt from the same region', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);

    const live = await screen.findByTestId('calabresella-prompt-live');
    expect(live).toContainElement(screen.getByTestId('calabresella-discard-prompt'));
    expect(live).toHaveAttribute('aria-live', 'polite');
  });

  // **領域は催促が無くても残る。**これが #5955 の教訓そのもの ── 領域ごと
  // 出し入れすると、中身が入った瞬間が「変化」にならない。
  it('keeps the region mounted and empty while no prompt applies', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<CalabresellaPage />);

    const live = await screen.findByTestId('calabresella-prompt-live');
    expect(live).toBeEmptyDOMElement();
    expect(screen.queryByTestId('calabresella-bid-prompt')).not.toBeInTheDocument();
    expect(screen.queryByTestId('calabresella-discard-prompt')).not.toBeInTheDocument();
  });
});
