import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mariasApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMariasState } from '../test/stateFactories';
import { MariasPage } from './MariasPage';

vi.mock('../api/gameApi', () => ({
  mariasApi: { exec: vi.fn() },
  actionLogApi: { marias: vi.fn() },
}));

const mockExec = vi.mocked(mariasApi.exec);

const playPhaseState = makeMariasState();
const trickEndState = makeMariasState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeMariasState({
  phase: 2,
  roundCardPoints: [55, 35, 30],
  roundMarriage: [40, 0, 0],
});
const gameEndState = makeMariasState({
  phase: 3,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeMariasState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('MariasPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MariasPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MariasPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 10 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Soloist badge', async () => {
    renderWithProviders(<MariasPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default Soloist.
    expect(screen.getByText('ソリスト')).toBeInTheDocument();
  });

  it('shows a marriage banner during play (trump-suit K-Q scores +40)', async () => {
    // Default hand: ♥K + ♥Q with trump ♥ → a +40 marriage.
    renderWithProviders(<MariasPage />);
    const banner = await screen.findByTestId('marias-marriage');
    expect(banner).toHaveTextContent('マリッジ可能');
    expect(banner).toHaveTextContent('♥ K-Q (+40)');
  });

  it('labels a non-trump marriage as +20', async () => {
    mockExec.mockResolvedValue(makeMariasState({ trumpSuit: 1 })); // trump ♠, ♥ K-Q now +20
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByTestId('marias-marriage')).toHaveTextContent('♥ K-Q (+20)'));
  });

  it('announces the marriage banner with a spoken suit name (symbol hidden from SR)', async () => {
    // Default hand: ♥K + ♥Q with trump ♥ → +40. SR reads the suit name, not the glyph.
    renderWithProviders(<MariasPage />);
    const banner = await screen.findByTestId('marias-marriage');
    expect(banner).toHaveAttribute('role', 'status');
    expect(banner).toHaveAttribute('aria-live', 'polite');
    expect(banner).toHaveAttribute('aria-label', 'マリッジ可能: ハート K-Q +40');
  });

  it('shows no marriage banner when the hand has no K-Q pair', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [
              { design: 'HEART', value: 13 },
              { design: 'SPADE', value: 12 },
            ],
            trickCount: 0,
            score: 0,
            isSoloist: true,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
          { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
        ],
      }),
    );
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByAltText('♥ K')).toBeInTheDocument());
    expect(screen.queryByTestId('marias-marriage')).not.toBeInTheDocument();
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<MariasPage />);
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
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
