import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fortyFivesApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeFortyFivesState } from '../test/stateFactories';
import { FortyFivesPage } from './FortyFivesPage';

vi.mock('../api/gameApi', () => ({
  fortyFivesApi: { exec: vi.fn() },
  actionLogApi: { fortyfives: vi.fn() },
}));

const mockExec = vi.mocked(fortyFivesApi.exec);

// Default fixture: a human bid turn (bid phase).
const bidPhaseState = makeFortyFivesState();
// A human play turn with playable cards (so the play control is shown).
const playPhaseState = makeFortyFivesState({
  phase: 1,
  declarerIdx: 0,
  contract: 20,
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
      teamScore: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
  ],
});
const cpuTurnState = makeFortyFivesState({
  phase: 1,
  declarerIdx: 1,
  isHumanBidTurn: false,
  isHumanTurn: false,
  currentPlayerIdx: 1,
});
const trickEndState = makeFortyFivesState({ phase: 2, isHumanBidTurn: false });
const roundEndState = makeFortyFivesState({ phase: 3, isHumanBidTurn: false, roundTeamPoints: [15, 10] });
const gameEndState = makeFortyFivesState({
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

describe('FortyFivesPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FortyFivesPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 45 },
      }),
    );
  });

  it('shows bid buttons on a human bid turn', async () => {
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByTestId('bid-0')).toBeInTheDocument());
    expect(screen.getByTestId('bid-15')).toBeInTheDocument();
    expect(screen.getByTestId('bid-20')).toBeInTheDocument();
    expect(screen.getByTestId('bid-25')).toBeInTheDocument();
  });

  it('dispatches a bid when a bid button is clicked', async () => {
    renderWithProviders(<FortyFivesPage />);
    const bid15 = await screen.findByTestId('bid-15');
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bid15);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 15 }));
  });

  it('disables bids that do not beat the current highest bid', async () => {
    mockExec.mockResolvedValue(makeFortyFivesState({ bids: [20, 0, 0, 0] }));
    renderWithProviders(<FortyFivesPage />);
    // Highest bid is 20: 15 and 20 are disabled, 25 and Pass (0) are enabled.
    await waitFor(() => expect(screen.getByTestId('bid-15')).toBeDisabled());
    expect(screen.getByTestId('bid-20')).toBeDisabled();
    expect(screen.getByTestId('bid-25')).toBeEnabled();
    expect(screen.getByTestId('bid-0')).toBeEnabled();
  });

  it('shows the current highest bid and a reason tooltip on too-low bids', async () => {
    mockExec.mockResolvedValue(makeFortyFivesState({ bids: [20, 0, 0, 0] }));
    renderWithProviders(<FortyFivesPage />);
    const info = await screen.findByTestId('ff-highest-bid');
    expect(info).toHaveTextContent('20');
    expect(info).not.toHaveTextContent('まだ入札なし');
    // The reason tooltip lives on the wrapping span (disabled buttons suppress native titles).
    expect(screen.getByTestId('bid-wrap-15')).toHaveAttribute('title');
    expect(screen.getByTestId('bid-15')).toHaveAttribute('aria-label');
    expect(screen.getByTestId('bid-wrap-25')).not.toHaveAttribute('title');
  });

  it('shows "no bids yet" before anyone has bid', async () => {
    mockExec.mockResolvedValue(makeFortyFivesState({ bids: [0, 0, 0, 0] }));
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByTestId('ff-highest-bid')).toHaveTextContent('まだ入札なし'));
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('落札者')).toBeInTheDocument();
  });

  it('shows the localized bid name (Jink) in the contract line for a 25 contract', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, contract: 25 });
    renderWithProviders(<FortyFivesPage />);
    // Contract line shows "25 (Jink)", not the bare number.
    await waitFor(() => expect(screen.getByText(/落札者:.*25 \(Jink\)/)).toBeInTheDocument());
  });

  it('shows the localized bid name for the current highest bid', async () => {
    mockExec.mockResolvedValue(makeFortyFivesState({ bids: [25, 0, 0, 0] }));
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByText(/現在の最高ビッド: 25 \(Jink\)/)).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('shows the live round points panel during play with team points', async () => {
    // Team A has won 2 tricks (10 pts), team B 1 trick (5 pts) so far this round.
    mockExec.mockResolvedValue({ ...playPhaseState, roundTeamPoints: [10, 5] });
    renderWithProviders(<FortyFivesPage />);
    const panel = await screen.findByTestId('ff-live-points');
    expect(panel).toHaveTextContent('ライブ得点');
    expect(panel).toHaveTextContent('チームA: 10点');
    expect(panel).toHaveTextContent('チームB: 5点');
  });

  it('shows the declarer contract progress with a "needMore" status while still reachable', async () => {
    // Declarer team A (declarerIdx 0) has 10 of a 20-pt contract; 15 pts remain, so it is still on track.
    mockExec.mockResolvedValue({ ...playPhaseState, contract: 20, declarerIdx: 0, roundTeamPoints: [10, 5] });
    renderWithProviders(<FortyFivesPage />);
    const progress = await screen.findByTestId('ff-contract-progress');
    expect(progress).toHaveTextContent('契約20点');
    expect(progress).toHaveTextContent('あと10点');
  });

  it('marks the contract as made once the declarer team reaches the contract', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, contract: 15, declarerIdx: 0, roundTeamPoints: [15, 5] });
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByTestId('ff-contract-progress')).toHaveTextContent('契約達成'));
  });

  it('marks the contract as unreachable when even every remaining trick would fall short', async () => {
    // Contract 25 for team A: A has 5, B has 15 (4 tricks resolved), only 1 trick (5 pts) left → max 10 < 25.
    mockExec.mockResolvedValue({ ...playPhaseState, contract: 25, declarerIdx: 0, roundTeamPoints: [5, 15] });
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByTestId('ff-contract-progress')).toHaveTextContent('達成不能'));
  });

  it('hides the live points panel at round end in favor of the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(screen.getByText('ラウンド結果')).toBeInTheDocument());
    expect(screen.queryByTestId('ff-live-points')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<FortyFivesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });
});
