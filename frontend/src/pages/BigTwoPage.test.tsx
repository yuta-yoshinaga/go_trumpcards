import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bigtwoApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BigTwoResponse, Card } from '../types/card';
import { BigTwoPage } from './BigTwoPage';

vi.mock('../api/gameApi', () => ({
  bigtwoApi: { exec: vi.fn() },
  actionLogApi: { bigtwo: vi.fn() },
}));

const mockExec = vi.mocked(bigtwoApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<BigTwoResponse['players'][number]> = {}) {
  return { id, isHuman, isFinished: false, rank: 0, cardCount: cards.length, cards, ...over };
}

function makeState(overrides: Partial<BigTwoResponse> = {}): BigTwoResponse {
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

describe('BigTwoPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BigTwoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders skeleton when fewer than 4 players are present', async () => {
    mockExec.mockResolvedValue(makeState({ players: [player(0, true, [])] }));
    renderWithProviders(<BigTwoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BigTwoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows play and pass buttons on the human turn', async () => {
    renderWithProviders(<BigTwoPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
    expect(screen.getByTestId('play-button')).toBeDisabled(); // nothing selected yet
  });

  it('selecting a card enables play and clicking plays it', async () => {
    renderWithProviders(<BigTwoPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('passes when the pass button is clicked', async () => {
    renderWithProviders(<BigTwoPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('toggles card selection on and off', async () => {
    renderWithProviders(<BigTwoPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeEnabled();
    fireEvent.click(card0);
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    localStorage.setItem('cli-mode-bigtwo', 'true');
    renderWithProviders(<BigTwoPage />);
    expect(await screen.findByPlaceholderText(/コマンド/)).toBeInTheDocument();
    expect(screen.queryByTestId('play-button')).not.toBeInTheDocument();
  });

  it('shows the shared ErrorAlert with a retry button when an action fails', async () => {
    renderWithProviders(<BigTwoPage />);
    const passBtn = await screen.findByTestId('pass-button');
    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(passBtn);
    // The error now surfaces via the shared ErrorAlert (role=alert), not a bare underline button.
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent(NETWORK_ERROR_MESSAGE());
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(screen.getByRole('button', { name: '再試行' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('renders the hand sort buttons', async () => {
    renderWithProviders(<BigTwoPage />);
    expect(await screen.findByTestId('bt-sort-strength')).toBeInTheDocument();
    expect(screen.getByTestId('bt-sort-suit')).toBeInTheDocument();
    expect(screen.getByTestId('bt-sort-number')).toBeInTheDocument();
  });

  it('keeps the selected card index stable across hand sorting', async () => {
    renderWithProviders(<BigTwoPage />);
    // Select ♦7 (original index 2), then re-sort by suit (which moves ♦ to the front).
    fireEvent.click(await screen.findByTestId('hand-card-2'));
    fireEvent.click(screen.getByTestId('bt-sort-suit'));
    // The play command still references the original dealt index.
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [2]));
  });

  it('previews the combo type of a single selected card', async () => {
    renderWithProviders(<BigTwoPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const preview = screen.getByTestId('bt-selected-playtype');
    expect(preview).toHaveTextContent('選択中');
    expect(preview).toHaveTextContent('シングル');
  });

  it('previews a pair when two same-rank cards are selected', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, [card('SPADE', 7), card('HEART', 7), card('DIAMOND', 3)]),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<BigTwoPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    expect(screen.getByTestId('bt-selected-playtype')).toHaveTextContent('ペア');
    expect(screen.getByTestId('play-button')).toBeEnabled();
  });

  it('warns and disables play for an invalid combination', async () => {
    renderWithProviders(<BigTwoPage />);
    // Default hand is ♠3, ♥5, ♦7 — selecting two of them is not a valid pair.
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    expect(screen.getByTestId('bt-selected-playtype')).toHaveTextContent('無効な組み合わせ');
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('hides the selection preview when nothing is selected', async () => {
    renderWithProviders(<BigTwoPage />);
    await screen.findByTestId('hand-card-0');
    expect(screen.queryByTestId('bt-selected-playtype')).not.toBeInTheDocument();
  });

  it('shows the table play-type label for the cards in play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        tableCards: [card('SPADE', 3), card('SPADE', 5), card('SPADE', 7), card('SPADE', 9), card('SPADE', 11)],
        tablePlayType: 5, // flush
        currentTurn: 1,
      }),
    );
    renderWithProviders(<BigTwoPage />);
    const label = await screen.findByTestId('bt-table-playtype');
    expect(label).toHaveTextContent('フラッシュ');
  });

  // **currentTurn は届いていたのに isHumanTurn の判定にしか使われていなかった。**
  // 誰の番かが画面に出ておらず、Daifugo / Sevens だけがハイライトしていた (#5478)。
  it('highlights the CPU whose turn it is', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2 }));
    renderWithProviders(<BigTwoPage />);

    const active = await screen.findByTestId('bt-cpu-2');
    expect(active.className).toContain('border-game-status-waiting');
    // 負のコントロール: 手番でない CPU には付かない。ここを見ないと
    // 「全員に枠が付く」実装でも通る。
    expect(screen.getByTestId('bt-cpu-1').className).not.toContain('border-game-status-waiting');
    expect(screen.getByTestId('bt-cpu-3').className).not.toContain('border-game-status-waiting');
  });

  it('does not highlight anyone once the game is over', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2, gameEndFlag: true }));
    renderWithProviders(<BigTwoPage />);
    await waitFor(() => expect(screen.getByTestId('bt-cpu-2')).toBeInTheDocument());
    expect(screen.getByTestId('bt-cpu-2').className).not.toContain('border-game-status-waiting');
  });

  it('dims a finished CPU instead of highlighting it', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentTurn: 2,
        players: [
          player(0, true, [card('SPADE', 3)]),
          player(1, false, [card('HEART', 4)]),
          player(2, false, [], { isFinished: true, rank: 1 }),
          player(3, false, [card('CLOVER', 6)]),
        ],
      }),
    );
    renderWithProviders(<BigTwoPage />);
    const finished = await screen.findByTestId('bt-cpu-2');
    expect(finished.className).toContain('opacity-50');
    expect(finished.className).not.toContain('border-game-status-waiting');
  });

  it('translates the CPU difficulty options instead of showing raw English', async () => {
    renderWithProviders(<BigTwoPage />);
    await screen.findByTestId('bt-cpu-2');
    const select = document.getElementById('cpuDifficulty');
    if (!(select instanceof HTMLSelectElement)) throw new Error('cpuDifficulty select not rendered');
    const labels = Array.from(select.options).map((o) => o.textContent);
    expect(labels).toEqual(['イージー', 'ノーマル', 'ハード']);
    // The value each label maps to must not shift: 0=Easy, 1=Normal, 2=Hard.
    expect(Array.from(select.options).map((o) => o.value)).toEqual(['0', '1', '2']);
  });
});
