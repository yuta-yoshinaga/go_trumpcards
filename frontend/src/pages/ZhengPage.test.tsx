import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { zhengApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ZhengResponse } from '../types/card';
import { ZhengPage } from './ZhengPage';

vi.mock('../api/gameApi', () => ({
  zhengApi: { exec: vi.fn() },
  actionLogApi: { zheng: vi.fn() },
}));

const mockExec = vi.mocked(zhengApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<ZhengResponse['players'][number]> = {}) {
  return { id, isHuman, isFinished: false, rank: 0, cardCount: cards.length, cards, ...over };
}

function makeState(overrides: Partial<ZhengResponse> = {}): ZhengResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)]),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

/** State where the human follows a CPU single (table has cards → pass is legal). */
function followState(overrides: Partial<ZhengResponse> = {}): ZhengResponse {
  return makeState({
    tableCards: [card('CLOVER', 4)],
    tablePlayType: 1,
    lastPlayPlayerIdx: 1,
    ...overrides,
  });
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('ZhengPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ZhengPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders skeleton when fewer than 4 players are present', async () => {
    mockExec.mockResolvedValue(makeState({ players: [player(0, true, [])] }));
    renderWithProviders(<ZhengPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ZhengPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('disables pass on a lead (empty table)', async () => {
    renderWithProviders(<ZhengPage />);
    expect(await screen.findByTestId('pass-button')).toBeDisabled();
    expect(screen.getByTestId('play-button')).toBeDisabled(); // nothing selected yet
  });

  it('enables pass when following a table play', async () => {
    mockExec.mockResolvedValue(followState());
    renderWithProviders(<ZhengPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
  });

  it('selecting a card enables play and clicking plays it', async () => {
    renderWithProviders(<ZhengPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('exposes a polite turn/finish live region', async () => {
    renderWithProviders(<ZhengPage />);
    const announce = await screen.findByTestId('zheng-turn-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    // No transition has occurred on first render, so it starts empty.
    expect(announce).toHaveTextContent('');
  });

  it('announces when the human finishes (rank named)', async () => {
    renderWithProviders(<ZhengPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    // The play resolves to a state where the human has gone out first (rank 1).
    mockExec.mockResolvedValue(
      makeState({
        currentTurn: 1,
        players: [
          player(0, true, [], { isFinished: true, rank: 1 }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(screen.getByTestId('zheng-turn-announce')).toHaveTextContent('あがりました'));
  });

  it('disables play and shows a reason for an invalid combination', async () => {
    renderWithProviders(<ZhengPage />);
    // Select ♠3 + ♥5 (two different ranks) → not a legal combo.
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    expect(screen.getByTestId('play-button')).toBeDisabled();
    expect(screen.getByTestId('zheng-invalid-combo')).toBeInTheDocument();
  });

  it('passes when the pass button is clicked', async () => {
    mockExec.mockResolvedValue(followState());
    renderWithProviders(<ZhengPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('toggles card selection on and off', async () => {
    renderWithProviders(<ZhengPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-zheng', 'true');
    renderWithProviders(<ZhengPage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('play-button')).not.toBeInTheDocument();
  });

  it('shows a retry button when an action fails', async () => {
    mockExec.mockResolvedValue(followState());
    renderWithProviders(<ZhengPage />);
    const passBtn = await screen.findByTestId('pass-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(passBtn);
    const retry = await screen.findByText(NETWORK_ERROR_MESSAGE());
    mockExec.mockResolvedValue(followState());
    fireEvent.click(retry);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('shows a finished CPU rank instead of a card count', async () => {
    // A CPU has gone out (rank 1) while the round continues — exercises the
    // `isFinished ? rank : cardCount` branch. Locale is ja in tests, so "1位".
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 3)]),
          player(1, false, [], { isFinished: true, rank: 1, cardCount: 0 }),
          player(2, false, [], { cardCount: 5 }),
          player(3, false, [], { cardCount: 9 }),
        ],
      }),
    );
    renderWithProviders(<ZhengPage />);
    await screen.findByTestId('pass-button');
    expect(screen.getAllByText('1位').length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU remaining-card counts during play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 3)]),
          player(1, false, [], { cardCount: 7 }),
          player(2, false, [], { cardCount: 5 }),
          player(3, false, [], { cardCount: 9 }),
        ],
      }),
    );
    renderWithProviders(<ZhengPage />);
    await screen.findByTestId('pass-button');
    expect(screen.getByText('— 7')).toBeInTheDocument();
  });

  it('hides the standings table until a player has finished', async () => {
    renderWithProviders(<ZhengPage />);
    await screen.findByTestId('pass-button');
    expect(screen.queryByTestId('zheng-rank-table')).not.toBeInTheDocument();
  });

  it('renders a standings table ordering finished players by rank then unfinished by fewest cards', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [], { isFinished: true, rank: 2, cardCount: 0 }),
          player(1, false, [], { isFinished: true, rank: 1, cardCount: 0 }),
          player(2, false, [card('SPADE', 3)], { cardCount: 3 }),
          player(3, false, [], { cardCount: 8 }),
        ],
      }),
    );
    renderWithProviders(<ZhengPage />);
    const table = await screen.findByTestId('zheng-rank-table');
    const rows = within(table).getAllByRole('listitem');
    expect(rows).toHaveLength(4);
    // CPU1 out first (1位), human out second (2位), then still-playing by fewest cards.
    expect(rows[0]).toHaveTextContent('1位');
    expect(rows[1]).toHaveTextContent('2位');
    expect(rows[2]).toHaveTextContent('3枚');
    expect(rows[3]).toHaveTextContent('8枚');
  });

  it('renders the game-end rankings message via messageCode', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        players: [
          player(0, true, [], { isFinished: true, rank: 1, cardCount: 0 }),
          player(1, false, [], { isFinished: true, rank: 2, cardCount: 0 }),
          player(2, false, [], { isFinished: true, rank: 3, cardCount: 0 }),
          player(3, false, [], { isFinished: true, rank: 4, cardCount: 0 }),
        ],
        message: 'fallback',
        messageCode: 'zheng.result.rankings',
        messageParams: { rankings: 'あなた:1位 ' },
      }),
    );
    renderWithProviders(<ZhengPage />);
    expect(await screen.findByText(/ゲーム終了/)).toBeInTheDocument();
  });
});
