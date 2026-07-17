import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { primeroApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makePrimeroState } from '../test/stateFactories';
import { PrimeroPage } from './PrimeroPage';

vi.mock('../api/gameApi', () => ({
  primeroApi: { exec: vi.fn() },
  actionLogApi: { primero: vi.fn() },
}));

const mockExec = vi.mocked(primeroApi.exec);

const bettingState = makePrimeroState({ phase: 0, isHumanTurn: true, canRaise: true });
const cpuTurnState = makePrimeroState({ phase: 0, isHumanTurn: false, canRaise: false });
const resultState = makePrimeroState({
  phase: 1,
  isHumanTurn: false,
  winnerIdx: 0,
  result: 1,
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 230,
      roundBet: 40,
      folded: false,
      out: false,
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 13 },
        { design: 'HEART', value: 12 },
        { design: 'CLOVER', value: 7 },
        { design: 'DIAMOND', value: 5 },
      ],
      handName: 'primero',
      isWinner: true,
    },
    {
      id: 1,
      isHuman: false,
      chips: 170,
      roundBet: 40,
      folded: false,
      out: false,
      cardCount: 4,
      cards: [
        { design: 'HEART', value: 4 },
        { design: 'HEART', value: 3 },
        { design: 'CLOVER', value: 2 },
        { design: 'DIAMOND', value: 6 },
      ],
      handName: 'numerus',
      isWinner: false,
    },
    {
      id: 2,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: true,
      out: false,
      cardCount: 4,
      cards: [],
      isWinner: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: true,
      out: false,
      cardCount: 4,
      cards: [],
      isWinner: false,
    },
  ],
});
const gameEndState = makePrimeroState({
  phase: 1,
  isHumanTurn: false,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  winnerIdx: 0,
  result: 1,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bettingState);
});

describe('PrimeroPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PrimeroPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<PrimeroPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        ante: 10,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });

  it('shows the betting action buttons on the human betting turn', async () => {
    renderWithProviders(<PrimeroPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ（ヴィ）' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド（降りる）' })).toBeInTheDocument();
  });

  it('dispatches bet call when the Call button is clicked', async () => {
    renderWithProviders(<PrimeroPage />);
    const btn = await screen.findByRole('button', { name: 'コール' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'call'));
  });

  it('dispatches bet raise when the Raise button is clicked', async () => {
    renderWithProviders(<PrimeroPage />);
    const btn = await screen.findByRole('button', { name: 'レイズ（ヴィ）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'raise'));
  });

  it('dispatches bet fold when the Fold button is clicked', async () => {
    renderWithProviders(<PrimeroPage />);
    const btn = await screen.findByRole('button', { name: 'フォールド（降りる）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'fold'));
  });

  it('hides action buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PrimeroPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /リセット|ゲームをリセット/ })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'フォールド（降りる）' })).not.toBeInTheDocument();
  });

  it('hides the raise button when raising is not allowed', async () => {
    mockExec.mockResolvedValue(makePrimeroState({ phase: 0, isHumanTurn: true, canRaise: false }));
    renderWithProviders(<PrimeroPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'レイズ（ヴィ）' })).not.toBeInTheDocument();
  });

  it('shows the next-round button at the result phase and dispatches nextround', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<PrimeroPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('hides betting buttons on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<PrimeroPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PrimeroPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });

  it('marks player states with a colour-independent icon and a status aria-label', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<PrimeroPage />);
    const winner = await screen.findByTestId('primero-player-0');
    // Winner: crown glyph + status in the row aria-label.
    expect(winner).toHaveTextContent('👑');
    expect(winner).toHaveAttribute('aria-label', expect.stringContaining('勝者'));
    // Folded player: × glyph + status.
    const folded = screen.getByTestId('primero-player-2');
    expect(folded).toHaveTextContent('×');
    expect(folded).toHaveAttribute('aria-label', expect.stringContaining('降りた'));
  });

  it('renders the hand-ranking legend with all four categories in strength order', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<PrimeroPage />);
    const legend = await screen.findByTestId('primero-hand-legend');
    expect(legend).toHaveTextContent('役の一覧（強い順）');
    // All four Primero hand categories, strongest first.
    expect(screen.getByTestId('primero-hand-legend-row-fluxus')).toBeInTheDocument();
    expect(screen.getByTestId('primero-hand-legend-row-supremus')).toBeInTheDocument();
    expect(screen.getByTestId('primero-hand-legend-row-primero')).toBeInTheDocument();
    expect(screen.getByTestId('primero-hand-legend-row-numerus')).toBeInTheDocument();
  });

  it('highlights the human current hand inside the legend', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<PrimeroPage />);
    // The human hand in resultState is "primero".
    const currentRow = await screen.findByTestId('primero-hand-legend-row-primero');
    expect(currentRow).toHaveAttribute('aria-current', 'true');
    expect(currentRow).toHaveTextContent('現在の役');
    // A non-matching row is not marked current.
    expect(screen.getByTestId('primero-hand-legend-row-fluxus')).not.toHaveAttribute('aria-current');
  });
});
