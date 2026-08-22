import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ramschApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RamschResponse } from '../types/card';
import { RamschPhase } from '../types/phases';
import { RamschPage } from './RamschPage';

vi.mock('../api/gameApi', async () => {
  const actual = await vi.importActual<typeof import('../api/gameApi')>('../api/gameApi');
  return { ...actual, ramschApi: { exec: vi.fn() } };
});

const mockExec = vi.mocked(ramschApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const player = (id: number, overrides: Partial<RamschResponse['players'][number]> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 10,
  cards: id === 0 ? [card('SPADE', 7), card('SPADE', 8), card('HEART', 10)] : [],
  cardPoints: 0,
  roundsWon: 0,
  roundsLost: 0,
  roundScore: 0,
  cumulativeScore: 0,
  trickCount: 0,
  ...overrides,
});

const baseState: RamschResponse = {
  players: [player(0), player(1), player(2)],
  phase: RamschPhase.PLAY,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  forehandIdx: 0,
  middlehandIdx: 1,
  rearhandIdx: 2,
  dealerIdx: 2,
  loserIdx: -1,
  durchmarsch: false,
  durchmarschIdx: -1,
  gameEndFlag: false,
  leadPlayerIdx: 0,
  config: { cpuDifficulty: 1, targetScore: 500 },
  message: '',
} as RamschResponse;

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(baseState);
});

describe('RamschPage', () => {
  // **切り札と得点の向きを常に出す。** スカート系のつもりで来た人が真っ先に
  // 取り違えるところで、無ければ盤が読めない。
  it('always states that jacks are trump and that points are penalties', async () => {
    renderWithProviders(<RamschPage />);
    await waitFor(() => expect(screen.getByTestId('ramsch-trump-note')).toBeInTheDocument());
    expect(screen.getByTestId('ramsch-trump-note')).toHaveTextContent('ジャック');
    expect(screen.getByTestId('ramsch-scoring-note')).toHaveTextContent('失います');
  });

  // **伏せ札はサーバが返すまで出さない。** 最終トリックの獲得者が受け取る 2 枚で、
  // 途中で見えると終盤の判断が完全情報になる。
  it('does not show the skat while the round is running', async () => {
    renderWithProviders(<RamschPage />);
    await waitFor(() => expect(screen.getByTestId('ramsch-trump-note')).toBeInTheDocument());
    expect(screen.queryByTestId('ramsch-skat-reveal')).not.toBeInTheDocument();
  });

  it('shows the skat once the server reveals it', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      phase: RamschPhase.ROUND_END,
      skat: [card('SPADE', 1), card('HEART', 10)],
    });
    renderWithProviders(<RamschPage />);
    await waitFor(() => expect(screen.getByTestId('ramsch-skat-reveal')).toBeInTheDocument());
  });

  // **誰が負けたかを表の中で言う。** 罰点なので数字だけ並べると、多い人が
  // 勝っているように読める。
  it('marks the player who took the most points as the loser', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      phase: RamschPhase.ROUND_END,
      loserIdx: 1,
      players: [player(0, { cardPoints: 30 }), player(1, { cardPoints: 78 }), player(2, { cardPoints: 12 })],
    });
    renderWithProviders(<RamschPage />);
    await waitFor(() => expect(screen.getByTestId('ramsch-loser-1')).toBeInTheDocument());
    // **負のコントロール**: 点の少ない席には印が付かない。
    expect(screen.queryByTestId('ramsch-loser-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('ramsch-loser-0')).not.toBeInTheDocument();
  });

  // Durchmarsch は別の印。敗者の印を出すと、総取りした人が負けたように読める。
  it('marks a Durchmarsch instead of a loser', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      phase: RamschPhase.ROUND_END,
      loserIdx: -1,
      durchmarsch: true,
      durchmarschIdx: 0,
      players: [player(0, { cardPoints: 120 }), player(1), player(2)],
    });
    renderWithProviders(<RamschPage />);
    await waitFor(() => expect(screen.getByTestId('ramsch-durchmarsch-0')).toBeInTheDocument());
    expect(screen.queryByTestId('ramsch-loser-0')).not.toBeInTheDocument();
  });

  it('plays the selected card', async () => {
    renderWithProviders(<RamschPage />);
    await waitFor(() => expect(screen.getByTestId('ramsch-trump-note')).toBeInTheDocument());

    const playButton = screen.getByRole('button', { name: 'カードを出す' });
    // 手札から 1 枚選ぶまでは押せない（選択なしで出せてしまわないこと）。
    expect(playButton).toBeDisabled();

    const cards = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-label')?.startsWith('♠'));
    expect(cards.length).toBeGreaterThan(0);
    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByRole('button', { name: 'カードを出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', expect.objectContaining({ cardIndex: 0 })));
  });

  it('advances the trick', async () => {
    mockExec.mockResolvedValue({ ...baseState, phase: RamschPhase.TRICK_END });
    renderWithProviders(<RamschPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances the round', async () => {
    mockExec.mockResolvedValue({ ...baseState, phase: RamschPhase.ROUND_END });
    renderWithProviders(<RamschPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
