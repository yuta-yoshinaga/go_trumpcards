import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tressetteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTressetteState } from '../test/stateFactories';
import { TressettePage } from './TressettePage';

vi.mock('../api/gameApi', () => ({
  tressetteApi: { exec: vi.fn() },
  actionLogApi: { tressette: vi.fn() },
}));

const mockExec = vi.mocked(tressetteApi.exec);

const playPhaseState = makeTressetteState();
const trickEndState = makeTressetteState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 3 } },
    { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
  ],
});
const roundEndState = makeTressetteState({ phase: 2 });
const gameEndState = makeTressetteState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ チームAの勝ち！',
});
const cpuTurnState = makeTressetteState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('TressettePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TressettePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with config', async () => {
    renderWithProviders(<TressettePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetPoints: 21,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<TressettePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<TressettePage />);
    const card = await screen.findByAltText('♠ 3');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('renders trick end with next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ チームAの勝ち！')).toBeInTheDocument());
  });

  it('does not show play button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
