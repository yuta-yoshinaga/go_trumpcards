import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { stealingbundlesApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, StealingBundlesResponse } from '../types/card';
import { StealingBundlesPage } from './StealingBundlesPage';

vi.mock('../api/gameApi', () => ({
  stealingbundlesApi: { exec: vi.fn() },
  actionLogApi: { stealingbundles: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(stealingbundlesApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('SPADE', 7), card('HEART', 9), card('CLOVER', 3), card('DIAMOND', 5)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? hand : [],
  bundleSize: 0,
  ...over,
});

function makeState(overrides: Partial<StealingBundlesResponse> = {}): StealingBundlesResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    tableCards: [card('CLOVER', 7)],
    tableMatches: { '0': [0] },
    stealTargets: {},
    canCapture: true,
    deckRemaining: 32,
    lastCaptureIdx: -1,
    currentPlayerIdx: 0,
    turnNumber: 2,
    packsDealt: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as StealingBundlesResponse;
}

const selectCard = (idx: number) => {
  fireEvent.click(screen.getAllByRole('button', { name: /を選ぶ$/ })[idx]);
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('StealingBundlesPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<StealingBundlesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **束の一番上が弱点、というのが規則そのもの。**
  it('states that a bundle can be taken whole', async () => {
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-rule')).toHaveTextContent(/束ごと丸ごと奪えます/);
  });

  // **空の場も情報。** 行が消えると見落としと区別が付きません。
  it('shows the table, empty or not', async () => {
    const { unmount } = renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-table')).toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeState({ tableCards: [] }));
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-table')).toHaveTextContent(/なし/);
  });

  // **束の一番上は CPU の分も見えます。** そこが狙われる場所だからです。
  it('shows every seat bundle size and top card', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0), seat(1, { bundleSize: 5, bundleTop: card('DIAMOND', 9) }), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<StealingBundlesPage />);
    const s1 = await screen.findByTestId('sb-seat-1');
    expect(s1).toHaveTextContent('束5枚');
    expect(s1).not.toHaveTextContent('なし');
    expect(screen.getByTestId('sb-seat-2')).toHaveTextContent('なし');
  });

  it('marks who captured last', async () => {
    const { unmount } = renderWithProviders(<StealingBundlesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('sb-seat-2')).not.toHaveTextContent(/直前に取った/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastCaptureIdx: 2 }));
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-seat-2')).toHaveTextContent(/直前に取った/);
  });

  // **選んだ札でできることだけを出す。** できない手は押させません。
  it('offers only the moves the selected card can make', async () => {
    mockExec.mockResolvedValue(makeState({ tableMatches: { '0': [0] }, stealTargets: { '1': [2] } }));
    renderWithProviders(<StealingBundlesPage />);
    await screen.findByTestId('sb-table');

    selectCard(0);
    expect(screen.getByTestId('sb-take-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('sb-steal-btn-2')).not.toBeInTheDocument();
    // **取れるときは置けません。**
    expect(screen.queryByTestId('sb-trail-btn')).not.toBeInTheDocument();

    selectCard(1);
    expect(screen.getByTestId('sb-steal-btn-2')).toBeInTheDocument();
    expect(screen.queryByTestId('sb-take-btn')).not.toBeInTheDocument();

    // 何もできない札を選んでも、取れる手がある間は置けない。
    selectCard(2);
    expect(screen.queryByTestId('sb-take-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sb-trail-btn')).not.toBeInTheDocument();
  });

  it('captures from the table with the selected card', async () => {
    renderWithProviders(<StealingBundlesPage />);
    await screen.findByTestId('sb-table');
    selectCard(0);

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('sb-take-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('take', 0));
  });

  // **略奪は相手を指名します。**
  it('steals the named seat bundle', async () => {
    mockExec.mockResolvedValue(makeState({ tableMatches: {}, stealTargets: { '1': [2, 3] } }));
    renderWithProviders(<StealingBundlesPage />);
    await screen.findByTestId('sb-table');
    selectCard(1);

    expect(screen.getByTestId('sb-steal-btn-2')).toBeInTheDocument();
    expect(screen.getByTestId('sb-steal-btn-3')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('sb-steal-btn-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('steal', 1, 3));
  });

  // **取れないときだけ場に置けます。**
  it('offers trailing only when nothing can be captured', async () => {
    mockExec.mockResolvedValue(makeState({ canCapture: false, tableMatches: {}, stealTargets: {} }));
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-status')).toHaveTextContent(/取れる手がありません/);

    selectCard(2);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('sb-trail-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trail', 2));
  });

  it('says a capture is compulsory while one exists', async () => {
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-status')).toHaveTextContent(/場に置くことはできません/);
  });

  it('lets you pick another card', async () => {
    renderWithProviders(<StealingBundlesPage />);
    await screen.findByTestId('sb-table');
    selectCard(0);
    expect(screen.getByTestId('sb-actions')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '選び直す' }));
    expect(screen.queryByTestId('sb-actions')).not.toBeInTheDocument();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<StealingBundlesPage />);
    const cards = await screen.findAllByRole('button', { name: /を選ぶ$/ });
    expect(cards[0]).toBeDisabled();
    expect(screen.queryByTestId('sb-status')).not.toBeInTheDocument();
  });

  it('reports who collected the most', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        gameEndFlag: true,
        winnerIdx: 0,
        players: [seat(0, { bundleSize: 30 }), seat(1), seat(2), seat(3)],
      }),
    );
    const { unmount } = renderWithProviders(<StealingBundlesPage />);
    const banner = await screen.findByTestId('sb-result');
    expect(banner).toHaveTextContent(/あなたの勝ち/);
    expect(banner).toHaveTextContent('30');
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        gameEndFlag: true,
        winnerIdx: 2,
        players: [seat(0), seat(1), seat(2, { bundleSize: 22 }), seat(3)],
      }),
    );
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('sb-result')).toHaveTextContent(/CPU2 が 22 枚/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<StealingBundlesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('resets with the chosen table size', async () => {
    renderWithProviders(<StealingBundlesPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // **サーバは 2..4 しか受けない。** 弾かれる値を並べると黙って既定に戻される。
    const options = [...screen.getByTestId('sb-players-select').querySelectorAll('option')].map((o) => o.value);
    expect(options).toEqual(['2', '3', '4']);

    fireEvent.change(screen.getByTestId('sb-players-select'), { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { playerCnt: 2 }));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-1', reason: 'hint.stealingbundlesSteal', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<StealingBundlesPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/丸ごと奪えます/);
  });
});
