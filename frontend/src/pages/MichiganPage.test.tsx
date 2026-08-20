import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { michiganApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMichiganState } from '../test/stateFactories';
import { MichiganPage } from './MichiganPage';

vi.mock('../api/gameApi', () => ({
  michiganApi: { exec: vi.fn() },
  actionLogApi: { michigan: vi.fn() },
}));

const mockExec = vi.mocked(michiganApi.exec);

const betState = makeMichiganState({ phase: 0, isHumanTurn: true, humanBetPlaced: false });
const cpuTurnState = makeMichiganState({ phase: 1, isHumanTurn: false, humanBetPlaced: true });
const playState = makeMichiganState({
  phase: 1,
  isHumanTurn: true,
  humanBetPlaced: true,
  needNewSequence: false,
  seqSuit: 1,
  seqSuitName: 'ハート',
  seqHighValue: 3,
  playableIndices: [0, 2],
});
const resultState = makeMichiganState({
  phase: 2,
  isHumanTurn: false,
  winnerIdx: 0,
  result: 1,
  players: [
    { id: 0, isHuman: true, chips: 210, roundBet: 8, cardCount: 0, cards: [], isCurrent: false, isWinner: true },
    {
      id: 1,
      isHuman: false,
      chips: 190,
      roundBet: 8,
      cardCount: 2,
      cards: [
        { design: 'HEART', value: 9 },
        { design: 'CLOVER', value: 5 },
      ],
      isCurrent: false,
      isWinner: false,
    },
    { id: 2, isHuman: false, chips: 195, roundBet: 8, cardCount: 3, cards: [], isCurrent: false, isWinner: false },
    { id: 3, isHuman: false, chips: 197, roundBet: 8, cardCount: 1, cards: [], isCurrent: false, isWinner: false },
  ],
});
const gameEndState = makeMichiganState({
  phase: 2,
  isHumanTurn: false,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  winnerIdx: 0,
  result: 1,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(betState);
});

describe('MichiganPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MichiganPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MichiganPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        playerCount: 4,
        ante: 8,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });

  // **数字だけでは謎の数でしかない。** 用語もその意味もどこにも説明が
  // 無かった (#5700)。
  it('explains what the dead hand is', async () => {
    renderWithProviders(<MichiganPage />);
    const help = await screen.findByTestId('michigan-deadhand-help');
    expect(help).toHaveTextContent('デッドハンドとは');
    expect(help).toHaveTextContent('誰の手にも入らず');
    // 枚数の表示は残っていること。説明で置き換えると枚数が消える。
    expect(screen.getByText(/デッドハンド: /)).toBeInTheDocument();
  });

  it('shows the place-bets button on the human bet turn', async () => {
    renderWithProviders(<MichiganPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '賭ける' })).toBeInTheDocument());
  });

  it('dispatches bet with the even chip distribution when Place bets is clicked', async () => {
    renderWithProviders(<MichiganPage />);
    const btn = await screen.findByRole('button', { name: '賭ける' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', [2, 2, 2, 2]));
  });

  it('disables the increment buttons and enables Place bets when the full budget is allocated', async () => {
    renderWithProviders(<MichiganPage />);
    const plus0 = await screen.findByRole('button', { name: 'ブードル 0 にチップを 1 枚足す' });
    expect(plus0).toBeDisabled();
    expect(screen.getByRole('button', { name: '賭ける' })).toBeEnabled();
  });

  it('redistributes chips with the +/- steppers and blocks Place bets until the sum matches the budget', async () => {
    renderWithProviders(<MichiganPage />);
    const minus0 = await screen.findByRole('button', { name: 'ブードル 0 からチップを 1 枚減らす' });
    fireEvent.click(minus0); // [1,2,2,2] — one chip freed, sum 7 != 8
    expect(screen.getByRole('button', { name: '賭ける' })).toBeDisabled();
    const plus0 = screen.getByRole('button', { name: 'ブードル 0 にチップを 1 枚足す' });
    expect(plus0).toBeEnabled();
    fireEvent.click(plus0); // back to [2,2,2,2]
    await waitFor(() => expect(screen.getByRole('button', { name: '賭ける' })).toBeEnabled());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭ける' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', [2, 2, 2, 2]));
  });

  it('clamps a boodle bet at zero when decremented past its minimum', async () => {
    renderWithProviders(<MichiganPage />);
    const minus0 = await screen.findByRole('button', { name: 'ブードル 0 からチップを 1 枚減らす' });
    fireEvent.click(minus0); // 2 -> 1
    fireEvent.click(minus0); // 1 -> 0
    fireEvent.click(minus0); // stays 0 (clamped)
    const plus0 = screen.getByRole('button', { name: 'ブードル 0 にチップを 1 枚足す' });
    fireEvent.click(plus0); // 0 -> 1
    fireEvent.click(plus0); // 1 -> 2, budget fully re-allocated
    await waitFor(() => expect(screen.getByRole('button', { name: '賭ける' })).toBeEnabled());
  });

  it('marks a boodle as collectible when the human holds its matching card', async () => {
    // Boodle 0 is A♥ (HEART 1); give the human that card so it becomes recoverable.
    const collectibleState = makeMichiganState({
      phase: 0,
      isHumanTurn: true,
      humanBetPlaced: false,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 192,
          roundBet: 8,
          cardCount: 2,
          cards: [
            { design: 'HEART', value: 1 }, // matches boodle 0 (A♥)
            { design: 'SPADE', value: 2 },
          ],
          isCurrent: true,
          isWinner: false,
        },
        { id: 1, isHuman: false, chips: 192, roundBet: 8, cardCount: 5, cards: [], isCurrent: false, isWinner: false },
        { id: 2, isHuman: false, chips: 192, roundBet: 8, cardCount: 5, cards: [], isCurrent: false, isWinner: false },
        { id: 3, isHuman: false, chips: 192, roundBet: 8, cardCount: 5, cards: [], isCurrent: false, isWinner: false },
      ],
    });
    mockExec.mockResolvedValue(collectibleState);
    renderWithProviders(<MichiganPage />);
    expect(await screen.findByTestId('bet-collectible-0')).toHaveTextContent('回収可能');
    // Boodle 1 (K♣) is not held, so no collectible mark there.
    expect(screen.queryByTestId('bet-collectible-1')).not.toBeInTheDocument();
  });

  it('shows no collectible marks when the human holds none of the boodle cards', async () => {
    renderWithProviders(<MichiganPage />); // default betState hand matches no boodle
    await screen.findByRole('button', { name: '賭ける' });
    expect(screen.queryByTestId('bet-collectible-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bet-collectible-1')).not.toBeInTheDocument();
  });

  it('warns on a boodle whose chips have already been claimed', async () => {
    const claimedState = makeMichiganState({
      phase: 0,
      isHumanTurn: true,
      humanBetPlaced: false,
      boodles: [
        { card: { design: 'HEART', value: 1 }, chips: 2, claimedBy: -1 },
        { card: { design: 'CLOVER', value: 13 }, chips: 2, claimedBy: 2 }, // claimed
        { card: { design: 'DIAMOND', value: 12 }, chips: 2, claimedBy: -1 },
        { card: { design: 'SPADE', value: 11 }, chips: 2, claimedBy: -1 },
      ],
    });
    mockExec.mockResolvedValue(claimedState);
    renderWithProviders(<MichiganPage />);
    expect(await screen.findByTestId('bet-claimed-warning-1')).toHaveTextContent('獲得済み・賭け注意');
    expect(screen.queryByTestId('bet-claimed-warning-0')).not.toBeInTheDocument();
  });

  it('hides the place-bets button when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MichiganPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /リセット|ゲームをリセット/ })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '賭ける' })).not.toBeInTheDocument();
  });

  it('plays a playable card and dispatches play with its index', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MichiganPage />);
    const card = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(card);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('disables non-playable hand cards during the play phase', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MichiganPage />);
    const card = await screen.findByTestId('hand-card-1');
    expect(card).toBeDisabled();
  });

  it('highlights playable cards with a success ring and a next-play badge', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MichiganPage />);
    const playable = await screen.findByTestId('hand-card-0');
    expect(playable.className).toContain('ring-ds-success');
    expect(playable).toHaveAttribute('data-playable', 'true');
    expect(screen.getByTestId('playable-badge-0')).toBeInTheDocument();
  });

  it('does not ring or badge a non-playable card', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MichiganPage />);
    const nonPlayable = await screen.findByTestId('hand-card-1');
    expect(nonPlayable.className).not.toContain('ring-ds-success');
    expect(nonPlayable).toHaveAttribute('data-playable', 'false');
    expect(screen.queryByTestId('playable-badge-1')).not.toBeInTheDocument();
  });

  it('shows the next-required-card hint for an active sequence', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MichiganPage />);
    expect(await screen.findByTestId('michigan-next-hint')).toHaveTextContent('次に出せる: ハート 4');
  });

  it('keeps a highlighted playable card clickable (ring stays additive)', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MichiganPage />);
    const playable = await screen.findByTestId('hand-card-0');
    expect(playable).not.toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(playable);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('shows the next-round button at the result phase and dispatches nextround', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<MichiganPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('hides the place-bets button on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<MichiganPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '賭ける' })).not.toBeInTheDocument();
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MichiganPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });
});
