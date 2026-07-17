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

// A hand mixing trumps (♥10, ♦K, ♣Q) with fail cards (♠A, ♣10) for highlight tests.
const mixedHandState = makeDoppelkopfState({
  canAnnounce: false,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART', value: 10 }, // trump (Dulle)
        { design: 'SPADE', value: 1 }, // fail (♠A)
        { design: 'DIAMOND', value: 13 }, // trump (♦K)
        { design: 'CLOVER', value: 10 }, // fail (♣10)
        { design: 'CLOVER', value: 12 }, // trump (♣Q)
      ],
      trickCount: 0,
      chips: 20,
      isRe: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, chips: 20, isRe: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, chips: 20, isRe: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, chips: 20, isRe: false },
  ],
  playableIndices: [0, 1, 2, 3, 4],
});

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

  it('renders the collapsible trump ordering legend', async () => {
    renderWithProviders(<DoppelkopfPage />);
    const legend = await screen.findByTestId('dk-trump-legend');
    expect(legend).toBeInTheDocument();
    expect(legend).toContainHTML('details');
    expect(screen.getByText('切り札序列')).toBeInTheDocument();
    // The strongest and weakest trumps appear in the ordering.
    expect(screen.getByText('♥10')).toBeInTheDocument();
    expect(screen.getByText('♦9')).toBeInTheDocument();
  });

  it('rings trump cards in the hand and leaves fail cards unmarked', async () => {
    mockExec.mockResolvedValue(mixedHandState);
    renderWithProviders(<DoppelkopfPage />);
    await screen.findByAltText('♥ 10');
    // Trumps: ♥10 (Dulle), ♦K, ♣Q.
    for (const alt of ['♥ 10', '♦ K', '♣ Q']) {
      expect(screen.getByAltText(alt).closest('button')).toHaveAttribute('data-trump', 'true');
    }
    // Fail cards: ♠A, ♣10.
    for (const alt of ['♠ A', '♣ 10']) {
      expect(screen.getByAltText(alt).closest('button')).not.toHaveAttribute('data-trump');
    }
  });
});
