import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { suecaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSuecaState } from '../test/stateFactories';
import { SuecaPage } from './SuecaPage';

vi.mock('../api/gameApi', () => ({
  suecaApi: { exec: vi.fn() },
  actionLogApi: { sueca: vi.fn() },
}));

const mockExec = vi.mocked(suecaApi.exec);

const playPhaseState = makeSuecaState();
const trickEndState = makeSuecaState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeSuecaState({ phase: 2, roundCardPoints: [70, 50] });
const gameEndState = makeSuecaState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeSuecaState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('SuecaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SuecaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SuecaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetGamePoints: 4 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<SuecaPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<SuecaPage />);
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
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
