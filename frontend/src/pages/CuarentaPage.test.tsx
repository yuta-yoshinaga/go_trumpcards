import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cuarentaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CuarentaPlayer, CuarentaResponse } from '../types/card';
import { CuarentaPage } from './CuarentaPage';

vi.mock('../api/gameApi', () => ({
  cuarentaApi: { exec: vi.fn() },
  actionLogApi: { cuarenta: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(cuarentaApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<CuarentaPlayer> = {}): CuarentaPlayer {
  return {
    id: 1,
    team: 1,
    isHuman: false,
    cardCount: 5,
    cards: [],
    capturedCount: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<CuarentaResponse> = {}): CuarentaResponse {
  return {
    players: [
      makePlayer({
        id: 0,
        team: 0,
        isHuman: true,
        cards: [card('SPADE', 5), card('HEART', 11), card('DIAMOND', 1), card('CLOVER', 7), card('SPADE', 2)],
      }),
      makePlayer({ id: 1, team: 1 }),
      makePlayer({ id: 2, team: 0 }),
      makePlayer({ id: 3, team: 1 }),
    ],
    currentTurn: 0,
    tableCards: [card('CLOVER', 7), card('HEART', 3)],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 0,
    teamScores: [12, 8],
    remainingDeck: 16,
    roundWinners: [],
    cpuActions: [],
    humanAction: null,
    lastRoundDetail: null,
    config: { targetScore: 40, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

const playState = makeState();
const emptyTableState = makeState({ tableCards: [] });
const roundEndState = makeState({ phase: 1, currentTurn: -1 });
const gameEndState = makeState({
  phase: 2,
  gameEndFlag: true,
  currentTurn: -1,
  roundWinners: [0],
  teamScores: [40, 31],
});
const cpuWinState = makeState({
  phase: 2,
  gameEndFlag: true,
  currentTurn: -1,
  roundWinners: [1],
  teamScores: [29, 40],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('CuarentaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CuarentaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the remaining stock', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText(/山札: 16枚/)).toBeInTheDocument());
  });

  it('renders team scores with the target', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText(/チームA: 12 \/ 40点/)).toBeInTheDocument());
    expect(screen.getByText(/チームB: 8 \/ 40点/)).toBeInTheDocument();
  });

  it('renders the players list with captured counts', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.getAllByText(/捕獲 0枚/).length).toBe(4);
  });

  it('sums each team captured-card total from its two players', async () => {
    // Team A = seats {0,2}: 8 + 5 = 13; Team B = seats {1,3}: 6 + 1 = 7.
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({ id: 0, team: 0, isHuman: true, capturedCount: 8 }),
          makePlayer({ id: 1, team: 1, capturedCount: 6 }),
          makePlayer({ id: 2, team: 0, capturedCount: 5 }),
          makePlayer({ id: 3, team: 1, capturedCount: 1 }),
        ],
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByTestId('cuarenta-team-captured-0')).toHaveTextContent('獲得 13枚'));
    expect(screen.getByTestId('cuarenta-team-captured-1')).toHaveTextContent('獲得 7枚');
  });

  it('highlights a team counter as it approaches the 20-card bonus', async () => {
    // Team A at 19 (approaching) is emphasized; Team B at 7 is not.
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({ id: 0, team: 0, isHuman: true, capturedCount: 10 }),
          makePlayer({ id: 1, team: 1, capturedCount: 4 }),
          makePlayer({ id: 2, team: 0, capturedCount: 9 }),
          makePlayer({ id: 3, team: 1, capturedCount: 3 }),
        ],
      }),
    );
    renderWithProviders(<CuarentaPage />);
    const teamA = await screen.findByTestId('cuarenta-team-captured-0');
    expect(teamA).toHaveTextContent('獲得 19枚');
    expect(teamA.className).toContain('text-ds-accent');
    const teamB = screen.getByTestId('cuarenta-team-captured-1');
    expect(teamB.className).not.toContain('text-ds-accent');
  });

  it('resets team captured totals to zero on a fresh round', async () => {
    // Default fixture has all capturedCount 0 — both team counters read 0.
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByTestId('cuarenta-team-captured-0')).toHaveTextContent('獲得 0枚'));
    expect(screen.getByTestId('cuarenta-team-captured-1')).toHaveTextContent('獲得 0枚');
  });

  it('renders the human hand cards', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-4')).toBeInTheDocument();
  });

  it('names each hand card in its aria-label', async () => {
    renderWithProviders(<CuarentaPage />);
    // hand[0] is ♠5, hand[2] is ♦A.
    expect(await screen.findByRole('button', { name: '♠ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ A を出す' })).toBeInTheDocument();
  });

  it('plays a hand card on the human turn', async () => {
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
  });

  it('shows the empty-table label when the table is empty', async () => {
    mockExec.mockResolvedValue(emptyTableState);
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('場札なし')).toBeInTheDocument());
  });

  it('does not dispatch play when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2 }));
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders capture-result badges for the human action', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7), card('HEART', 7), card('SPADE', 7)],
          isCaida: true,
          isLimpia: true,
          rondaBonus: 1,
        },
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('カイーダ! +2')).toBeInTheDocument());
    expect(screen.getByText('ロンダ! +1')).toBeInTheDocument();
    expect(screen.getByText('リンピア! +1')).toBeInTheDocument();
    // The human's bonus badges pop (motion-safe) to draw the eye.
    const popped = screen.getAllByTestId('cuarenta-bonus-pop');
    expect(popped.length).toBeGreaterThan(0);
    expect(popped[0].className).toContain('motion-safe:animate-bounce');
    // ...and the bonus row is announced to assistive tech.
    const announce = screen.getByTestId('cuarenta-bonus-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
  });

  it('chimes once when a fresh human bonus lands, but not on a plain play', async () => {
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    // A plain capture (no bonus) must not chime.
    mockExec.mockResolvedValueOnce(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7)],
          isCaida: false,
          isLimpia: false,
          rondaBonus: 0,
        },
      }),
    );
    fireEvent.click(cardBtn);
    await waitFor(() => expect(screen.getByText('直前のプレイ')).toBeInTheDocument());
    expect(mockPlaySound).not.toHaveBeenCalled();

    // A subsequent bonus play chimes exactly once.
    const nextCard = await screen.findByTestId('hand-card-0');
    mockExec.mockResolvedValueOnce(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('SPADE', 5),
          capturedCards: [card('SPADE', 5)],
          isCaida: true,
          isLimpia: false,
          rondaBonus: 0,
        },
      }),
    );
    fireEvent.click(nextCard);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('chipClick', { pitchVariation: 0.1 }));
    expect(mockPlaySound).toHaveBeenCalledTimes(1);
  });

  it('shows the win message when the human team wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('あなたのチームの勝利です！')).toBeInTheDocument());
  });

  it('shows the lose message naming the winning team', async () => {
    mockExec.mockResolvedValue(cpuWinState);
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('チームB の勝利です。')).toBeInTheDocument());
  });

  it('shows a next-round button at round end and dispatches next', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CuarentaPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });
});
