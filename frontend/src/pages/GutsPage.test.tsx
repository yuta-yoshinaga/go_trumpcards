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

  it('shows a strong (high) win-chance guideline and the pair name for a paired hand', async () => {
    const pairState = makeGutsState({
      phase: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 200,
          in: false,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 9 },
            { design: 'HEART', value: 9 },
          ],
          isWinner: false,
          isMatcher: false,
        },
        ...declareState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(pairState);
    renderWithProviders(<GutsPage />);
    const guide = await screen.findByTestId('guts-declare-guide');
    expect(guide).toHaveTextContent('手役: ペア');
    expect(screen.getByTestId('guts-guide-tier')).toHaveTextContent('高い');
    // Pot (40) match-loss risk is surfaced next to the buttons.
    expect(screen.getByTestId('guts-guide-risk')).toHaveTextContent('ポット 40 相当');
  });

  it('shows a weak (low) win-chance guideline and the high-card name for a low hand', async () => {
    const weakState = makeGutsState({
      phase: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 200,
          in: false,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 2 },
            { design: 'HEART', value: 7 },
          ],
          isWinner: false,
          isMatcher: false,
        },
        ...declareState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(weakState);
    renderWithProviders(<GutsPage />);
    const guide = await screen.findByTestId('guts-declare-guide');
    expect(guide).toHaveTextContent('手役: ハイカード');
    expect(screen.getByTestId('guts-guide-tier')).toHaveTextContent('低い');
  });

  it('hides the declaration guideline on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('guts-declare-guide')).not.toBeInTheDocument();
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });

  // **誰も残らなかったラウンドは表示が丸ごと無かった。**CUI は最初から
  // guts.result.carry で持ち越し額と連続回数を出している (#4847)。
  it('shows the carried pot when nobody stayed in', async () => {
    mockExec.mockResolvedValue(makeGutsState({ phase: 1, winnerIdx: -1, carryPot: 40, carryCount: 2 }));
    renderWithProviders(<GutsPage />);

    const panel = await screen.findByTestId('guts-carry-result');
    expect(panel).toHaveTextContent('ポット 40');
    expect(panel).toHaveTextContent('2 回連続');
  });

  it('shows the winner panel instead when the round had a winner', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('guts-carry-result')).not.toBeInTheDocument();
  });
});
