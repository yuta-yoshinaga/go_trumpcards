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
