import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gutsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGutsState } from '../test/stateFactories';
import { GutsPage } from './GutsPage';

vi.mock('../api/gameApi', () => ({
  gutsApi: { exec: vi.fn() },
  actionLogApi: { guts: vi.fn() },
}));

const mockExec = vi.mocked(gutsApi.exec);

const declareState = makeGutsState({ phase: 0 });
const resultState = makeGutsState({
  phase: 1,
  winnerIdx: 0,
  result: 1,
  matchers: [1],
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 230,
      in: true,
      out: false,
      roundBet: 10,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
      ],
      handName: 'highcard',
      isWinner: true,
      isMatcher: false,
    },
    {
      id: 1,
      isHuman: false,
      chips: 170,
      in: true,
      out: false,
      roundBet: 90,
      cardCount: 2,
      cards: [
        { design: 'HEART', value: 5 },
        { design: 'CLOVER', value: 3 },
      ],
      handName: 'highcard',
      isWinner: false,
      isMatcher: true,
    },
    {
      id: 2,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 0,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 0,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
  ],
});
const gameEndState = makeGutsState({
  phase: 1,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  winnerIdx: 0,
  result: 1,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(declareState);
});

describe('GutsPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GutsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<GutsPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        ante: 10,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });

  it('shows the declare action buttons on the declare phase', async () => {
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'イン（残る）' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アウト（降りる）' })).toBeInTheDocument();
  });

  it('dispatches declare in when the In button is clicked', async () => {
    renderWithProviders(<GutsPage />);
    const btn = await screen.findByRole('button', { name: 'イン（残る）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 1));
  });

  it('dispatches declare out when the Out button is clicked', async () => {
    renderWithProviders(<GutsPage />);
    const btn = await screen.findByRole('button', { name: 'アウト（降りる）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 0));
  });

  it('shows the next-round button at the result phase and dispatches nextround', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('hides declare buttons on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'イン（残る）' })).not.toBeInTheDocument();
  });

  it('shows the matcher extra payment in the round result', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    const line = await screen.findByTestId('guts-matcher-payment');
    // roundBet 90 minus ante 10 = 80 chips matched into the next pot.
    expect(line).toHaveTextContent('マッチ支払い: CPU 1 が -80');
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });
});
