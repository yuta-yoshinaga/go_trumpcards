import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { twentyNineApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTwentyNineState } from '../test/stateFactories';
import { TwentyNinePage } from './TwentyNinePage';

vi.mock('../api/gameApi', () => ({
  twentyNineApi: { exec: vi.fn() },
  actionLogApi: { twentynine: vi.fn() },
}));

const mockExec = vi.mocked(twentyNineApi.exec);

// Default fixture: a human bid turn (bid phase).
const bidPhaseState = makeTwentyNineState();
// A human play turn with playable cards (so the play control is shown). Trump is revealed.
const playPhaseState = makeTwentyNineState({
  phase: 1,
  declarerIdx: 0,
  contract: 20,
  trumpSuit: 3,
  trumpRevealed: true,
  isHumanBidTurn: false,
  isHumanTurn: true,
  playableIndices: [0, 1, 2],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 8,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
  ],
});
const cpuTurnState = makeTwentyNineState({
  phase: 1,
  declarerIdx: 1,
  isHumanBidTurn: false,
  isHumanTurn: false,
  currentPlayerIdx: 1,
});
const trickEndState = makeTwentyNineState({ phase: 2, isHumanBidTurn: false });
const roundEndState = makeTwentyNineState({ phase: 3, isHumanBidTurn: false, roundTeamPoints: [18, 11] });
const gameEndState = makeTwentyNineState({
  phase: 4,
  isHumanBidTurn: false,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidPhaseState);
});

describe('TwentyNinePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TwentyNinePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 6 },
      }),
    );
  });

  it('shows bid buttons on a human bid turn', async () => {
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByTestId('bid-0')).toBeInTheDocument());
    expect(screen.getByTestId('bid-16')).toBeInTheDocument();
    expect(screen.getByTestId('bid-20')).toBeInTheDocument();
    expect(screen.getByTestId('bid-24')).toBeInTheDocument();
    expect(screen.getByTestId('bid-28')).toBeInTheDocument();
  });

  it('shows "no bids yet" when no one has bid during the bid phase', async () => {
    renderWithProviders(<TwentyNinePage />);
    const readout = await screen.findByTestId('tn29-highest-bid');
    expect(readout).toHaveTextContent('まだ入札なし');
  });

  it('shows the current highest bid and the bidder name during the bid phase', async () => {
    // CPU 2 (seat index 2) holds the highest bid of 20.
    mockExec.mockResolvedValue(makeTwentyNineState({ bids: [0, 0, 20, 0] }));
    renderWithProviders(<TwentyNinePage />);
    const readout = await screen.findByTestId('tn29-highest-bid');
    expect(readout).toHaveTextContent('現在の最高ビッド: 20（CPU 2）');
  });

  it('updates the highest-bid readout as bids change', async () => {
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByTestId('tn29-highest-bid')).toHaveTextContent('まだ入札なし'));
    // The next server response reflects the human's own bid of 16.
    mockExec.mockResolvedValue(makeTwentyNineState({ bids: [16, 0, 0, 0] }));
    fireEvent.click(screen.getByTestId('bid-20'));
    await waitFor(() =>
      expect(screen.getByTestId('tn29-highest-bid')).toHaveTextContent('現在の最高ビッド: 16（あなた）'),
    );
  });

  it('dispatches a bid when a bid button is clicked', async () => {
    renderWithProviders(<TwentyNinePage />);
    const bid16 = await screen.findByTestId('bid-16');
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bid16);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 16 }));
  });

  it('disables bids that do not beat the current highest bid', async () => {
    mockExec.mockResolvedValue(makeTwentyNineState({ bids: [20, 0, 0, 0] }));
    renderWithProviders(<TwentyNinePage />);
    // Highest bid is 20: 16 and 20 are disabled, 24/28 and Pass (0) are enabled.
    await waitFor(() => expect(screen.getByTestId('bid-16')).toBeDisabled());
    expect(screen.getByTestId('bid-20')).toBeDisabled();
    expect(screen.getByTestId('bid-24')).toBeEnabled();
    expect(screen.getByTestId('bid-28')).toBeEnabled();
    expect(screen.getByTestId('bid-0')).toBeEnabled();
  });

  it('explains why a disabled bid cannot be chosen via title and aria-label', async () => {
    mockExec.mockResolvedValue(makeTwentyNineState({ bids: [20, 0, 0, 0] }));
    renderWithProviders(<TwentyNinePage />);
    const bid16 = await screen.findByTestId('bid-16');
    const reason = '現在の最高ビッド 20 を超える必要があります';
    // aria-disabled + reason-bearing aria-label on the button.
    expect(bid16).toHaveAttribute('aria-disabled', 'true');
    expect(bid16).toHaveAttribute('aria-label', `16 — ${reason}`);
    // The hover tooltip lives on the wrapping span (browsers suppress it on disabled buttons).
    expect(screen.getByTestId('bid-wrap-16')).toHaveAttribute('title', reason);
    // An enabled bid gets neither a reason label nor a title.
    expect(screen.getByTestId('bid-24')).not.toHaveAttribute('aria-label');
    expect(screen.getByTestId('bid-wrap-24')).not.toHaveAttribute('title');
  });

  it('hides the trump suit until it is revealed', async () => {
    // Bid-phase fixture has trumpRevealed false; the trump should read "非公開" (hidden).
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByText(/切り札: 非公開/)).toBeInTheDocument());
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('落札者')).toBeInTheDocument();
  });

  it('flashes a trump-reveal banner when the hidden trump becomes revealed', async () => {
    // Start on a human play turn with the trump still hidden.
    mockExec.mockResolvedValue({ ...playPhaseState, trumpRevealed: false });
    renderWithProviders(<TwentyNinePage />);
    const card = await screen.findByAltText('♥ Q');
    expect(screen.queryByTestId('tn-trump-reveal-banner')).not.toBeInTheDocument();

    // Play a card; the resolved state now has the trump revealed (♥ = suit 3).
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);

    const banner = await screen.findByTestId('tn-trump-reveal-banner');
    expect(banner).toHaveTextContent('♥');
    expect(banner).toHaveAttribute('role', 'status');
  });

  it('shows live round card points during play', async () => {
    mockExec.mockResolvedValue(makeTwentyNineState({ phase: 1, isHumanBidTurn: false, roundTeamPoints: [12, 7] }));
    renderWithProviders(<TwentyNinePage />);
    const live = await screen.findByTestId('tn29-round-points');
    expect(live).toHaveTextContent('12');
    expect(live).toHaveTextContent('7');
  });

  it('hides the live round-points block at round end (the result block takes over)', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByText('ラウンド結果（カード点）')).toBeInTheDocument());
    expect(screen.queryByTestId('tn29-round-points')).not.toBeInTheDocument();
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果（カード点）')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<TwentyNinePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });
});
