import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gaigelApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, GaigelResponse } from '../types/card';
import { GaigelPhase } from '../types/phases';
import { GaigelPage } from './GaigelPage';

vi.mock('../api/gameApi', () => ({
  gaigelApi: { exec: vi.fn() },
  actionLogApi: { gaigel: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(gaigelApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<GaigelResponse> = {}): GaigelResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 13), card('SPADE', 12), card('HEART', 10)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: GaigelPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    stockRemaining: 28,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 101 },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: GaigelPhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [101, 60],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(initialState);
});

describe('GaigelPage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 101,
      }),
    );
  });

  it('renders the team score table and stock readout during play', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.getByText('山札: 28')).toBeInTheDocument();
  });

  it('renders the face-up turn-up card when the stock still holds it', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: card('HEART', 10) }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByTestId('gaigel-trump-card')).toBeInTheDocument());
    expect(screen.getByText('めくり札')).toBeInTheDocument();
  });

  it('omits the turn-up card once the stock is exhausted', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, stockRemaining: 0 }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('gaigel-trump-card')).not.toBeInTheDocument();
  });

  it('dispatches play with the selected card index during play', async () => {
    renderWithProviders(<GaigelPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('shows the marriage button when the selected card can declare and dispatches marriage', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0] }));
    renderWithProviders(<GaigelPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    const marriageBtn = await screen.findByRole('button', { name: 'マリッジ' });
    mockExec.mockClear();
    fireEvent.click(marriageBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', undefined, 0));
  });

  it('hides the marriage button when the selected card cannot declare', async () => {
    renderWithProviders(<GaigelPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'マリッジ' })).not.toBeInTheDocument();
  });

  it('badges both King and Queen when a marriage is available on the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0, 1] }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByTestId('card-role-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('card-role-badge-1')).toBeInTheDocument();
    // The non-marriage card (index 2) gets no badge.
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument();
    expect(screen.getByTestId('card-role-badge-0')).toHaveTextContent('💍');
  });

  it('shows no marriage badge when no marriage is available', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-1')).not.toBeInTheDocument();
  });

  it('shows no marriage badge when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0, 1], currentPlayerIdx: 1 }));
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-1')).not.toBeInTheDocument();
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GaigelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });
});
