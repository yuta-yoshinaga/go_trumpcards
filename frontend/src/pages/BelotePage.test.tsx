import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beloteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BeloteResponse, Card } from '../types/card';
import { BelotePhase } from '../types/phases';
import { BelotePage } from './BelotePage';

vi.mock('../api/gameApi', () => ({
  beloteApi: { exec: vi.fn() },
  actionLogApi: { belote: vi.fn() },
}));

const mockExec = vi.mocked(beloteApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<BeloteResponse> = {}): BeloteResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 11), card('HEART', 10), card('CLOVER', 9)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: BelotePhase.BID_PICK_UP,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 0,
    faceUpCard: card('HEART', 11),
    makerTeam: 0,
    makerPlayerIdx: -1,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundBeloteBonus: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: {
      cpuDifficulty: 1,
      targetScore: 1000,
      dixDeDer: 10,
      enableBeloteRebelote: true,
    },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: BelotePhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [1010, 600],
});

beforeEach(() => {
  mockExec.mockResolvedValue(initialState);
});

describe('BelotePage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 1000,
        dixDeDer: 10,
        enableBeloteRebelote: true,
      }),
    );
  });

  it('shows the orderup and pass buttons in pick-up phase when human is bid turn', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '取る' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('dispatches orderup when "取る" is clicked', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() => screen.getByRole('button', { name: '取る' }));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('orderup'));
  });

  it('shows trump suit buttons during call-trump phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.BID_CALL_TRUMP }));
    renderWithProviders(<BelotePage />);
    // 3 callable suits (face-up suit HEART excluded): ♠ ♣ ♦
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♣ クラブ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ ダイヤ' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '♥ ハート' })).not.toBeInTheDocument();
  });

  it('dispatches calltrump with the picked suit', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.BID_CALL_TRUMP }));
    renderWithProviders(<BelotePage />);
    await waitFor(() => screen.getByRole('button', { name: '♠ スペード' }));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♠ スペード' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('calltrump', undefined, 1));
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('renders bonus trackers (dim by default) during play after trump is set', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 3, makerTeam: 0 }));
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('belote-bonus-trackers')).toBeInTheDocument());
    expect(screen.getByTestId('dix-de-der-badge')).not.toHaveAttribute('data-active');
    expect(screen.getByTestId('belote-rebelote-badge')).not.toHaveAttribute('data-active');
  });

  it('activates the dix-de-der badge on the 8th trick', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 8, makerTeam: 0 }));
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('dix-de-der-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('activates the belote/rebelote badge once the maker team earns the bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [20, 0] }),
    );
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('belote-rebelote-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });
});
