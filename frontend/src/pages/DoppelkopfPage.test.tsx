import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doppelkopfApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeDoppelkopfState } from '../test/stateFactories';
import { DoppelkopfPage } from './DoppelkopfPage';

vi.mock('../api/gameApi', () => ({
  doppelkopfApi: { exec: vi.fn() },
  actionLogApi: { doppelkopf: vi.fn() },
}));

const mockExec = vi.mocked(doppelkopfApi.exec);

const playPhaseState = makeDoppelkopfState({ canAnnounce: false });
const announceState = makeDoppelkopfState({ canAnnounce: true, youAreRe: true });
const kontraAnnounceState = makeDoppelkopfState({ canAnnounce: true, youAreRe: false });
const trickEndState = makeDoppelkopfState({
  phase: 1,
  canAnnounce: false,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 10 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 12 } },
  ],
});
const roundEndState = makeDoppelkopfState({
  phase: 2,
  canAnnounce: false,
  teamsRevealed: true,
  reTeam: [true, false, true, false],
  roundRePoints: 130,
  roundReWon: true,
  roundGamePoints: 2,
});
const gameEndState = makeDoppelkopfState({
  phase: 3,
  canAnnounce: false,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeDoppelkopfState({ currentPlayerIdx: 1, canAnnounce: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('DoppelkopfPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DoppelkopfPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, baseChips: 2, startChips: 20, targetChips: 40 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ 10')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const card = await screen.findByAltText('♥ 10');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('shows the hint button on the human turn and requests a hint', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('shows the Re announce button when the human is Re and can announce', async () => {
    mockExec.mockResolvedValue(announceState);
    renderWithProviders(<DoppelkopfPage />);
    const btn = await screen.findByRole('button', { name: /Re を宣言/ });
    // aria-label adds the timing/scoring context while still containing the visible label.
    expect(btn.getAttribute('aria-label')).toContain('Re を宣言');
    expect(btn.getAttribute('aria-label')).toContain('得点');
    expect(btn).toHaveAttribute('title');
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('announce'));
  });

  it('shows the Kontra announce button when the human is not Re', async () => {
    mockExec.mockResolvedValue(kontraAnnounceState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Kontra を宣言/ })).toBeInTheDocument());
  });

  it('does not show the announce button when canAnnounce is false', async () => {
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 10')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /宣言/ })).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DoppelkopfPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 10')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });
});
