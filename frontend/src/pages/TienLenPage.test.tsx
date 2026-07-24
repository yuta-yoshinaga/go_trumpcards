import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tienlenApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TienLenResponse } from '../types/card';
import { TienLenPage } from './TienLenPage';

vi.mock('../api/gameApi', () => ({
  tienlenApi: { exec: vi.fn() },
  actionLogApi: { tienlen: vi.fn() },
}));

const mockExec = vi.mocked(tienlenApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<TienLenResponse['players'][number]> = {}) {
  return { id, isHuman, isFinished: false, rank: 0, cardCount: cards.length, cards, ...over };
}

function makeState(overrides: Partial<TienLenResponse> = {}): TienLenResponse {
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
    config: { cpuDifficulty: 1 },
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('TienLenPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TienLenPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders skeleton when fewer than 4 players are present', async () => {
    mockExec.mockResolvedValue(makeState({ players: [player(0, true, [])] }));
    renderWithProviders(<TienLenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<TienLenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows play and pass buttons on the human turn', async () => {
    renderWithProviders(<TienLenPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
    expect(screen.getByTestId('play-button')).toBeDisabled(); // nothing selected yet
  });

  it('selecting a card enables play and clicking plays it', async () => {
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('disables play and shows a reason for an invalid combination', async () => {
    renderWithProviders(<TienLenPage />);
    // Select ♠3 + ♥5 (two different ranks) → not a legal combo.
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    expect(screen.getByTestId('play-button')).toBeDisabled();
    expect(screen.getByTestId('tl-invalid-combo')).toBeInTheDocument();
    // The combo-type badge is only shown for valid selections.
    expect(screen.queryByTestId('tl-combo-type')).not.toBeInTheDocument();
  });

  it('shows the combo type name for a single-card selection', async () => {
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const badge = screen.getByTestId('tl-combo-type');
    expect(badge).toHaveTextContent('シングル'); // playType.single (ja locale)
    expect(badge).toHaveTextContent('1枚');
  });

  it('shows the combo type name for a pair selection', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 7), card('HEART', 7)]),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    const badge = screen.getByTestId('tl-combo-type');
    expect(badge).toHaveTextContent('ペア'); // playType.pair
    expect(badge).toHaveTextContent('2枚');
  });

  it('highlights a four-of-a-kind bomb', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 7)]),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('hand-card-2'));
    fireEvent.click(screen.getByTestId('hand-card-3'));
    const badge = screen.getByTestId('tl-combo-type');
    expect(badge).toHaveTextContent('フォーカード'); // playType.fourOfAKind
    expect(badge).toHaveTextContent('爆弾'); // bombLabel emphasis
    expect(badge).toHaveClass('text-ds-warning');
  });

  it('labels the table combo with the CPU player who played it', async () => {
    mockExec.mockResolvedValue(
      makeState({
        tableCards: [card('SPADE', 3)],
        lastPlayPlayerIdx: 2,
        currentTurn: 0,
      }),
    );
    renderWithProviders(<TienLenPage />);
    const owner = await screen.findByTestId('tl-table-owner');
    expect(owner).toHaveTextContent('CPU 2'); // findPlayerName for a non-human player
  });

  it('does not show the table owner label when the table is empty (new round lead)', async () => {
    mockExec.mockResolvedValue(makeState({ tableCards: [], lastPlayPlayerIdx: -1 }));
    renderWithProviders(<TienLenPage />);
    await screen.findByTestId('pass-button');
    expect(screen.queryByTestId('tl-table-owner')).not.toBeInTheDocument();
  });

  it('passes when the pass button is clicked', async () => {
    renderWithProviders(<TienLenPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('toggles card selection on and off', async () => {
    renderWithProviders(<TienLenPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-tienlen', 'true');
    renderWithProviders(<TienLenPage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('play-button')).not.toBeInTheDocument();
  });

  it('shows a retry button when an action fails', async () => {
    renderWithProviders(<TienLenPage />);
    const passBtn = await screen.findByTestId('pass-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(passBtn);
    const retry = await screen.findByText(NETWORK_ERROR_MESSAGE());
    mockExec.mockResolvedValue(makeState());
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
    renderWithProviders(<TienLenPage />);
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
    renderWithProviders(<TienLenPage />);
    await screen.findByTestId('pass-button');
    expect(screen.getByText('— 7')).toBeInTheDocument();
  });
});
