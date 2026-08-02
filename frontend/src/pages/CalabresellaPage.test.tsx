import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { calabresellaApi } from '../api/gameApi';
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
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeCalabresellaState({
  phase: 4,
  roundThirds: [20, 8, 5],
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

  it('renders the bid phase with chiamo, solo and pass buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'キアーモ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ソロ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
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
    expect(screen.getByRole('button', { name: /カードを捨てる/ })).toBeInTheDocument();
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
    expect(screen.getByTestId('calabresella-discard-button')).toHaveTextContent('(4)');
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

  it('selecting a card then discarding dispatches discard', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CalabresellaPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const discardBtn = await screen.findByRole('button', { name: /カードを捨てる/ });
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 0 }));
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
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CalabresellaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
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
});
