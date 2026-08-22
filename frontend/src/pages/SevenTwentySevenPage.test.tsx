import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevenTwentySevenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SevenTwentySevenResponse } from '../types/card';
import { SevenTwentySevenPhase } from '../types/phases';
import { SevenTwentySevenPage } from './SevenTwentySevenPage';

vi.mock('../api/gameApi', async () => {
  const actual = await vi.importActual<typeof import('../api/gameApi')>('../api/gameApi');
  return { ...actual, sevenTwentySevenApi: { exec: vi.fn() } };
});

const mockExec = vi.mocked(sevenTwentySevenApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const player = (
  id: number,
  overrides: Partial<SevenTwentySevenResponse['players'][number]> = {},
): SevenTwentySevenResponse['players'][number] => ({
  id,
  isHuman: id === 0,
  chips: 200,
  standing: false,
  out: false,
  roundBet: 10,
  cardCount: 2,
  cards: id === 0 ? [card('SPADE', 4), card('HEART', 13)] : [],
  lowScore: id === 0 ? '4.5' : '',
  highScore: id === 0 ? '4.5' : '',
  wonLow: false,
  wonHigh: false,
  ...overrides,
});

const baseState: SevenTwentySevenResponse = {
  players: [player(0), player(1), player(2), player(3)],
  phase: SevenTwentySevenPhase.DRAW,
  roundNumber: 1,
  drawRound: 1,
  pot: 40,
  carryPot: 0,
  carryCount: 0,
  ante: 10,
  chips: 200,
  lowWinner: -1,
  highWinner: -1,
  matchWinnerIdx: -1,
  result: 0,
  gameEndFlag: false,
  config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
  message: '',
} as SevenTwentySevenResponse;

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(baseState);
});

describe('SevenTwentySevenPage', () => {
  // **2 つの目標を常に書く。** 7 と 27 のどちらに寄せるかがこのゲームそのもの。
  it('always states both targets and the unusual card values', async () => {
    renderWithProviders(<SevenTwentySevenPage />);
    const note = await screen.findByTestId('s27-targets-note');
    expect(note).toHaveTextContent('7');
    expect(note).toHaveTextContent('27');
    expect(note).toHaveTextContent('0.5');
  });

  // **両側の得点を出す。** 片方だけでは、いま何を狙えるのかが読めない。
  it('shows the low and high totals for the human', async () => {
    renderWithProviders(<SevenTwentySevenPage />);
    await waitFor(() => expect(screen.getByTestId('s27-your-score')).toHaveTextContent('4.5 / 4.5'));
  });

  // 超過した側はサーバが "-" を返す。数字を出すと勝負が残っているように読める。
  it('renders a busted side as a dash', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [player(0, { lowScore: '-', highScore: '19' }), player(1), player(2), player(3)],
    });
    renderWithProviders(<SevenTwentySevenPage />);
    await waitFor(() => expect(screen.getByTestId('s27-your-score')).toHaveTextContent('- / 19'));
  });

  it('takes a card and stands pat', async () => {
    renderWithProviders(<SevenTwentySevenPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'カードを引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('card'));

    fireEvent.click(screen.getByRole('button', { name: '止まる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  // **止まったら打つ手は無い。** サーバがラウンドを回し切るので、
  // ボタンを残すと押せるのに何も起きない。
  it('hides the actions once the human has stood pat', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [player(0, { standing: true }), player(1), player(2), player(3)],
    });
    renderWithProviders(<SevenTwentySevenPage />);
    await waitFor(() => expect(screen.getByTestId('s27-targets-note')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'カードを引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '止まる' })).not.toBeInTheDocument();
  });

  // **両側の勝者を名指しする。** どちらを取ったのかが分からないと、
  // なぜ半分なのかが読めない。
  it('names both side winners at the result', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      phase: SevenTwentySevenPhase.RESULT,
      lowWinner: 0,
      highWinner: 2,
      players: [
        player(0, { wonLow: true, lowScore: '6', highScore: '6' }),
        player(1),
        player(2, { wonHigh: true, lowScore: '-', highScore: '27', cards: [card('SPADE', 10)] }),
        player(3),
      ],
    });
    renderWithProviders(<SevenTwentySevenPage />);
    await waitFor(() => expect(screen.getByTestId('s27-low-result')).toHaveTextContent('6'));
    expect(screen.getByTestId('s27-high-result')).toHaveTextContent('27');
    expect(screen.queryByTestId('s27-scoop-result')).not.toBeInTheDocument();
  });

  // 総取りは専用の表示。半分ずつの表記だと誤解する。
  it('announces a scoop', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      phase: SevenTwentySevenPhase.RESULT,
      lowWinner: 0,
      highWinner: 0,
      players: [player(0, { wonLow: true, wonHigh: true }), player(1), player(2), player(3)],
    });
    renderWithProviders(<SevenTwentySevenPage />);
    await waitFor(() => expect(screen.getByTestId('s27-scoop-result')).toBeInTheDocument());
    expect(screen.queryByTestId('s27-low-result')).not.toBeInTheDocument();
  });

  // 両側とも全滅したら持ち越しと言う。何も描かないとチップが動かない理由が消える。
  it('says the pot carries over when everyone busts', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      phase: SevenTwentySevenPhase.RESULT,
      lowWinner: -1,
      highWinner: -1,
      carryPot: 40,
      carryCount: 1,
    });
    renderWithProviders(<SevenTwentySevenPage />);
    await waitFor(() => expect(screen.getByTestId('s27-carry-result')).toBeInTheDocument());
  });

  it('advances to the next round', async () => {
    mockExec.mockResolvedValue({ ...baseState, phase: SevenTwentySevenPhase.RESULT, lowWinner: 0 });
    renderWithProviders(<SevenTwentySevenPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
