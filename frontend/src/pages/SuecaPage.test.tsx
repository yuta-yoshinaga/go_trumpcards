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

  it('exposes the trump suit by name in the header and sidebar (symbol hidden from SR)', async () => {
    mockExec.mockResolvedValue(playPhaseState); // trumpSuit 4 = ♦ → ダイヤ
    renderWithProviders(<SuecaPage />);
    const sidebar = await screen.findByTestId('sueca-sidebar-trump');
    expect(sidebar).toHaveAttribute('role', 'img');
    expect(sidebar).toHaveAttribute('aria-label', '切り札: ダイヤ');
    // Header + sidebar both expose a named trump element (glyph is aria-hidden).
    expect(screen.getAllByRole('img', { name: '切り札: ダイヤ' }).length).toBeGreaterThanOrEqual(2);
  });

  it('falls back to the raw symbol when the trump suit is unset', async () => {
    // trumpSuit 0 has no suit-name key → the label falls back to the symbol string.
    mockExec.mockResolvedValue(makeSuecaState({ trumpSuit: 0 }));
    renderWithProviders(<SuecaPage />);
    const sidebar = await screen.findByTestId('sueca-sidebar-trump');
    expect(sidebar).toHaveAttribute('role', 'img');
    expect(sidebar).toHaveAttribute('aria-label', '切り札: ');
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

  it('shows the trump suit in the always-visible sidebar', async () => {
    renderWithProviders(<SuecaPage />); // default trump ♦
    await waitFor(() => expect(screen.getByTestId('sueca-sidebar-trump')).toHaveTextContent('切り札: ♦'));
  });

  it('announces the trick-winning team at trick end (leadPlayerIdx → team)', async () => {
    mockExec.mockResolvedValue(trickEndState); // leadPlayerIdx 0 → Team A
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByTestId('sueca-trick-winner')).toHaveTextContent('チームA がトリック獲得'));
  });

  it('names Team B when an odd seat wins the trick', async () => {
    mockExec.mockResolvedValue(makeSuecaState({ phase: 1, leadPlayerIdx: 1 }));
    renderWithProviders(<SuecaPage />);
    await waitFor(() => expect(screen.getByTestId('sueca-trick-winner')).toHaveTextContent('チームB がトリック獲得'));
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
