import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ramsApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RamsResponse } from '../types/card';
import { RamsPage } from './RamsPage';

vi.mock('../api/gameApi', () => ({
  ramsApi: { exec: vi.fn() },
  actionLogApi: { rams: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(ramsApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 9), card('CLOVER', 12)] : [],
  chips: 57,
  inRound: false,
  decided: false,
  roundTricks: 0,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<RamsResponse> = {}): RamsResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    pot: 12,
    trumpSuit: 3,
    upCard: card('HEART', 9),
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    activeCount: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4, rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as RamsResponse;
}

/** A state where the human entered the round and it is their turn. */
const playing = (over: Partial<RamsResponse> = {}) =>
  makeState({
    phase: 1,
    players: [seat(0, { decided: true, inRound: true }), seat(1), seat(2), seat(3)],
    ...over,
  } as Partial<RamsResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('RamsPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<RamsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **配り直後は参加選択。** ここを逃すと判断の機会が無い。
  it('opens on the decision phase with both choices', async () => {
    renderWithProviders(<RamsPage />);
    expect(await screen.findByTestId('rm-in-btn')).toBeInTheDocument();
    expect(screen.getByTestId('rm-out-btn')).toBeInTheDocument();
  });

  it('hides the choices once play has started', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<RamsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('rm-in-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('rm-out-btn')).not.toBeInTheDocument();
  });

  // **参加と降りるは別のコマンドを送る。** 取り違えるとラウンドが正反対になる。
  it.each([
    ['rm-in-btn', 'in'],
    ['rm-out-btn', 'out'],
  ])('sends %s as the %s command', async (testId, command) => {
    renderWithProviders(<RamsPage />);
    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<RamsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('card', 2));
  });

  // **ポット・切り札・リスクは参加判断の材料。** 盤面からは読めない。
  it('always shows the pot, the trump card and the penalty', async () => {
    renderWithProviders(<RamsPage />);
    expect(await screen.findByTestId('rm-pot')).toHaveTextContent('12');
    expect(screen.getByTestId('rm-trump')).toBeInTheDocument();
    expect(screen.getByTestId('rm-risk')).toHaveTextContent('5');
  });

  // **人数は可変。** 何人卓かを必ず出す。
  it.each([
    [3, '3人'],
    [5, '5人'],
  ])('reports a %i-player table', async (playerCnt, expected) => {
    mockExec.mockResolvedValue(
      makeState({
        config: { playerCnt, rounds: 4 },
        players: Array.from({ length: playerCnt }, (_, i) => seat(i)),
      } as Partial<RamsResponse>),
    );
    renderWithProviders(<RamsPage />);
    expect(await screen.findByTestId('rm-players')).toHaveTextContent(expected);
    expect(screen.getByTestId(`rm-seat-${playerCnt - 1}`)).toBeInTheDocument();
  });

  // 参加状況が席ごとに出る。3通りすべて。
  it('renders in / out / undecided per seat', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          seat(0, { decided: true, inRound: true, roundTricks: 2, chips: 50 }),
          seat(1, { decided: true, inRound: false }),
          seat(2),
          seat(3),
        ],
      } as Partial<RamsResponse>),
    );
    renderWithProviders(<RamsPage />);

    expect(await screen.findByTestId('rm-seat-0')).toHaveTextContent('参加');
    expect(screen.getByTestId('rm-seat-0')).toHaveTextContent('50チップ / 獲得2');
    expect(screen.getByTestId('rm-seat-1')).toHaveTextContent('降り');
    expect(screen.getByTestId('rm-seat-2')).toHaveTextContent('未定');
  });

  // **降りたラウンドは「見ている」と伝える。** 操作待ちに見えてはいけない。
  it('says it is watching when the player dropped out', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        players: [seat(0, { decided: true, inRound: false }), seat(1), seat(2), seat(3)],
      } as Partial<RamsResponse>),
    );
    renderWithProviders(<RamsPage />);
    expect(await screen.findByTestId('rm-watching')).toBeInTheDocument();
  });

  it('does not say that while the player is in the round', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<RamsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('rm-watching')).not.toBeInTheDocument();
  });

  // 降りたラウンドでは札を押せない。
  it('disables the hand for a player who dropped out', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        players: [seat(0, { decided: true, inRound: false }), seat(1), seat(2), seat(3)],
      } as Partial<RamsResponse>),
    );
    renderWithProviders(<RamsPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<RamsPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<RamsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerIdx, expected] of [
      [0, /あなたの勝ち/],
      [1, /CPU1 の勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 3, winnerIdx }));
      const { unmount } = renderWithProviders(<RamsPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'play-in', reason: 'hint.ramsPlayIn', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<RamsPage />);
    expect(await screen.findByText(/参加する価値があります/)).toBeInTheDocument();
  });
});
