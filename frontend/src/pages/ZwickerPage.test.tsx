import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { zwickerApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, ZwickerCard, ZwickerPlayer, ZwickerResponse } from '../types/card';
import { ZwickerPhase } from '../types/phases';
import { ZwickerPage } from './ZwickerPage';

vi.mock('../api/gameApi', () => ({
  zwickerApi: { exec: vi.fn() },
  actionLogApi: { zwicker: vi.fn() },
}));

const mockExec = vi.mocked(zwickerApi.exec);

const card = (design: CardDesign, value: number, values: number[]): ZwickerCard => ({ design, value, values });

function seat(id: number, isHuman: boolean, overrides?: Partial<ZwickerPlayer>): ZwickerPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 3,
    // ♠K は 4/14 の 2 択、♣7 は 7 のみ、大ジョーカーは 25 固定。
    cards: isHuman ? [card('SPADE', 13, [4, 14]), card('CLOVER', 7, [7]), card('JOKER', 3, [25])] : [],
    capturedCount: 6,
    zwicks: 1,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ZwickerResponse>): ZwickerResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: ZwickerPhase.PLAY,
    currentPlayerIdx: 0,
    stockCount: 20,
    tableCards: [card('HEART', 4, [4]), card('DIAMOND', 3, [3])],
    builds: [
      {
        owner: 1,
        value: 9,
        cards: [
          { design: 'SPADE', value: 5 },
          { design: 'HEART', value: 4 },
        ],
      },
    ],
    teamScores: [12, 8],
    targetScore: 61,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  };
}

/** Click the hand card at the given index. */
function pickHand(i: number) {
  const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'discard');
  fireEvent.click(hand[i]);
}

describe('ZwickerPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently and the running score', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/A=1\/11/)).toBeInTheDocument();
    expect(screen.getByText(/場を空にするとZwickで1点/)).toBeInTheDocument();
    expect(screen.getByText(/味方12 相手8/)).toBeInTheDocument();
  });

  // **A と絵札は 2 択を持つ**ので、札を選んだだけでは取れない。
  it('asks which value a court card is used as before it will capture', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    pickHand(0); // ♠K
    fireEvent.click(screen.getAllByTestId('zwicker-table-card')[0]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    const options = screen.getAllByTestId('zwicker-value-option');
    expect(options).toHaveLength(2);
    fireEvent.click(options[0]); // 4 として使う
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        cardIndex: 0,
        playedValue: 4,
        tableIndices: [0],
        buildIndices: [],
      }),
    );
  });

  // 値が 1 つしかない札で選ばせるのは無駄な一手。
  it('does not ask for a value when the card has only one', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    pickHand(1); // ♣7
    expect(screen.queryByTestId('zwicker-value-picker')).not.toBeInTheDocument();
    fireEvent.click(screen.getAllByTestId('zwicker-table-card')[0]);
    fireEvent.click(screen.getAllByTestId('zwicker-table-card')[1]);
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        cardIndex: 1,
        playedValue: 7,
        tableIndices: [0, 1],
        buildIndices: [],
      }),
    );
  });

  it('keeps table and build selections in their own lists', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    pickHand(1); // ♣7
    fireEvent.click(screen.getAllByTestId('zwicker-build')[0]);
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        cardIndex: 1,
        playedValue: 7,
        tableIndices: [],
        buildIndices: [0],
      }),
    );
  });

  it('will not capture nothing', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    pickHand(1);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('builds from a declared value', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    pickHand(1);
    fireEvent.click(screen.getAllByTestId('zwicker-table-card')[0]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ビルド' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText('宣言値'), { target: { value: '11' } });
    fireEvent.click(screen.getByRole('button', { name: 'ビルド' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('build', { cardIndex: 1, tableIndices: [0], declaredValue: 11 }),
    );
  });

  it('trails the selected card', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '置く' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    pickHand(2);
    fireEvent.click(screen.getByRole('button', { name: '置く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trail', { cardIndex: 2 }));
  });

  it('shows the build with its declared value', async () => {
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(screen.getAllByTestId('zwicker-build')[0]).toHaveTextContent('9'));
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: ZwickerPhase.ROUND_END,
        lastRound: { cardPoints: [17, 10], cards: [30, 25], majorityTeam: 0, zwicks: [1, 0], total: [21, 10] },
      }),
    );
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(screen.getByTestId('zwicker-round-result')).toHaveTextContent('21'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // 同数で 3 点が宙に浮いたことを黙って通すと、合計が合わないように見える。
  it('says when the card counts were level', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: ZwickerPhase.ROUND_END,
        lastRound: { cardPoints: [10, 10], cards: [27, 27], majorityTeam: -1, zwicks: [0, 0], total: [10, 10] },
      }),
    );
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(screen.getByTestId('zwicker-round-result')).toHaveTextContent('同数'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'zwicker.win', '勝ちました'],
      [1, 'zwicker.lose', '負けました'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: ZwickerPhase.GAME_END, gameEndFlag: true, winnerTeam: winner, messageCode: code }),
      );
      renderWithProviders(<ZwickerPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  it('highlights the table cards the hint says to take with the named card', async () => {
    localStorage.clear();
    // The lift/ring assist follows the hint setting, which defaults to off.
    localStorage.setItem('hint_enabled_zwicker', 'true');
    mockExec.mockReset();
    mockExec.mockResolvedValue(
      makeState({
        hint: { take: true, cardIndex: 0, value: 7, tableIndices: [1], reason: 'zwicker.hint.take' },
        messageCode: 'zwicker.hintRequested',
      }),
    );
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(screen.getAllByTestId('zwicker-table-card').length).toBeGreaterThan(1));
    // Naming the hand card alone left the multi-card capture unexplained.
    const tableCards = screen.getAllByTestId('zwicker-table-card');
    expect(tableCards[1]).toHaveAttribute('data-hinted-table');
    expect(tableCards[0]).not.toHaveAttribute('data-hinted-table');
  });

  it('highlights nothing until the hint is actually requested', async () => {
    localStorage.clear();
    localStorage.setItem('hint_enabled_zwicker', 'true');
    mockExec.mockReset();
    // Every response has carried state.hint since #4483, so an ungated read
    // lights the board up for a player who never pressed the button (#4605).
    mockExec.mockResolvedValue(
      makeState({
        hint: { take: true, cardIndex: 0, value: 7, tableIndices: [1], reason: 'zwicker.hint.take' },
        messageCode: 'zwicker.playing',
      }),
    );
    renderWithProviders(<ZwickerPage />);
    await waitFor(() => expect(screen.getAllByTestId('zwicker-table-card').length).toBeGreaterThan(1));
    expect(document.querySelectorAll('[data-hinted-table]')).toHaveLength(0);
  });
});
