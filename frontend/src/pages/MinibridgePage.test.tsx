import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { minibridgeApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, MinibridgeResponse } from '../types/card';
import { MinibridgePage } from './MinibridgePage';

vi.mock('../api/gameApi', () => ({
  minibridgeApi: { exec: vi.fn() },
  actionLogApi: { minibridge: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(minibridgeApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('HEART', 1), card('SPADE', 9), card('CLOVER', 13), card('DIAMOND', 4), card('SPADE', 2)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 5,
  cards: id === 0 ? hand : [],
  hcp: 10,
  team: id % 2,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<MinibridgeResponse> = {}): MinibridgeResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    contractLevel: 0,
    contractSuit: 0,
    requiredTricks: 0,
    declarerIdx: 0,
    dummyIdx: 2,
    dummyHand: [],
    lastMade: false,
    lastTricks: 0,
    teamScores: [0, 0],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as MinibridgeResponse;
}

/** A settled contract with the human declaring, so the dummy is theirs to play. */
const playing = (over: Partial<MinibridgeResponse> = {}) =>
  makeState({
    phase: 1,
    contractLevel: 2,
    contractSuit: 3,
    requiredTricks: 8,
    declarerIdx: 0,
    dummyIdx: 2,
    dummyHand: [card('SPADE', 1), card('HEART', 7)],
    ...over,
  } as Partial<MinibridgeResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('MinibridgePage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<MinibridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **競りが無いこと自体が規則。**
  it('states that there is no auction', async () => {
    renderWithProviders(<MinibridgePage />);
    expect(await screen.findByTestId('mb-rule')).toHaveTextContent(/競りはありません/);
  });

  // **HCP は公開情報。** 4 席ぶん出て、合計は 40。
  it("shows every seat's HCP", async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { hcp: 13 }), seat(1, { hcp: 9 }), seat(2, { hcp: 11 }), seat(3, { hcp: 7 })] }),
    );
    renderWithProviders(<MinibridgePage />);
    let total = 0;
    for (const [id, hcp] of [
      [0, 13],
      [1, 9],
      [2, 11],
      [3, 7],
    ] as const) {
      expect(await screen.findByTestId(`mb-seat-${id}`)).toHaveTextContent(`HCP${hcp}`);
      total += hcp;
    }
    expect(total).toBe(40);
  });

  it('marks the declarer and the dummy', async () => {
    mockExec.mockResolvedValue(playing({ declarerIdx: 1, dummyIdx: 3 } as Partial<MinibridgeResponse>));
    renderWithProviders(<MinibridgePage />);
    expect(await screen.findByTestId('mb-seat-1')).toHaveTextContent(/デクレアラー/);
    expect(screen.getByTestId('mb-seat-3')).toHaveTextContent(/ダミー/);
    expect(screen.getByTestId('mb-seat-0')).not.toHaveTextContent(/デクレアラー/);
  });

  it('offers all five denominations to the declarer', async () => {
    renderWithProviders(<MinibridgePage />);
    for (const suit of [1, 2, 3, 4, 0]) {
      expect(await screen.findByTestId(`mb-contract-${suit.toString()}-btn`)).toBeEnabled();
    }
  });

  it('hides the contract buttons when the human is not the declarer', async () => {
    mockExec.mockResolvedValue(makeState({ declarerIdx: 1, dummyIdx: 3 }));
    renderWithProviders(<MinibridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('mb-contract-0-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('mb-level-select')).not.toBeInTheDocument();
  });

  // **レベルとスートは 4・5 番目の引数で送る。** ずれると別の契約として届く。
  it('sends the selected level and denomination', async () => {
    renderWithProviders(<MinibridgePage />);
    fireEvent.change(await screen.findByTestId('mb-level-select'), { target: { value: '4' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('mb-contract-3-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', undefined, undefined, 4, 3));
  });

  // **ノートランプは suit 0 で送る。** 「省略」と混ざらないこと。
  it('sends no-trump as suit zero', async () => {
    renderWithProviders(<MinibridgePage />);
    const btn = await screen.findByTestId('mb-contract-0-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', undefined, undefined, 1, 0));
  });

  // **ペアの合計 HCP は契約の大きさを決める材料。**
  it("shows the declaring pair's combined HCP while choosing", async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { hcp: 13 }), seat(1, { hcp: 9 }), seat(2, { hcp: 11 }), seat(3, { hcp: 7 })] }),
    );
    renderWithProviders(<MinibridgePage />);
    expect(await screen.findByTestId('mb-pair-hcp')).toHaveTextContent('24');
  });

  it('shows the contract once chosen, with the tricks it needs', async () => {
    const { unmount } = renderWithProviders(<MinibridgePage />);
    expect(await screen.findByTestId('mb-contract')).toHaveTextContent(/未決定/);
    unmount();

    mockExec.mockResolvedValue(playing());
    renderWithProviders(<MinibridgePage />);
    const contract = await screen.findByTestId('mb-contract');
    expect(contract).toHaveTextContent('♥');
    expect(contract).toHaveTextContent('8');
  });

  // **ダミーは契約が決まってから公開される。**
  it('reveals the dummy only after the contract', async () => {
    const { unmount } = renderWithProviders(<MinibridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('mb-dummy')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(playing());
    renderWithProviders(<MinibridgePage />);
    expect(await screen.findByTestId('mb-dummy')).toBeInTheDocument();
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<MinibridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /^(?!ダミーの).*を出す$/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **ダミーの手番は人間デクレアラーの出番。** 自分の手札ではなくダミーが押せる。
  it('swaps which hand is pressable on the dummy turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 2, validPlays: [1] } as Partial<MinibridgeResponse>));
    renderWithProviders(<MinibridgePage />);

    const dummyCards = await screen.findAllByRole('button', { name: /^ダミーの/ });
    expect(dummyCards[0]).toBeEnabled();
    const ownCards = screen.getAllByRole('button', { name: /^(?!ダミーの).*を出す$/ });
    expect(ownCards[0]).toBeDisabled();

    mockExec.mockClear();
    fireEvent.click(dummyCards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  // 自分の手番ではダミーは押せない。
  it('keeps the dummy unpressable on your own turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 0 } as Partial<MinibridgeResponse>));
    renderWithProviders(<MinibridgePage />);
    const dummyCards = await screen.findAllByRole('button', { name: /^ダミーの/ });
    expect(dummyCards[0]).toBeDisabled();
    const ownCards = screen.getAllByRole('button', { name: /^(?!ダミーの).*を出す$/ });
    expect(ownCards[0]).toBeEnabled();
  });

  // **CPU がデクレアラーなら、ダミーも CPU のもの。**
  it('never lets you play a CPU declarer’s dummy', async () => {
    mockExec.mockResolvedValue(
      playing({ declarerIdx: 1, dummyIdx: 3, currentPlayerIdx: 3 } as Partial<MinibridgeResponse>),
    );
    renderWithProviders(<MinibridgePage />);
    const dummyCards = await screen.findAllByRole('button', { name: /^ダミーの/ });
    expect(dummyCards[0]).toBeDisabled();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<MinibridgeResponse>));
    renderWithProviders(<MinibridgePage />);
    const cards = await screen.findAllByRole('button', { name: /^(?!ダミーの).*を出す$/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the running totals', async () => {
    mockExec.mockResolvedValue(playing({ teamScores: [420, 100] } as Partial<MinibridgeResponse>));
    renderWithProviders(<MinibridgePage />);
    const score = await screen.findByTestId('mb-score');
    expect(score).toHaveTextContent('420');
    expect(score).toHaveTextContent('100');
  });

  it('reports the deal result', async () => {
    for (const [over, expected] of [
      [{ lastMade: true, lastTricks: 9 }, /成立/],
      [{ lastMade: false, lastTricks: 6 }, /失敗/],
    ] as const) {
      mockExec.mockResolvedValue(playing({ phase: 2, ...over } as Partial<MinibridgeResponse>));
      const { unmount } = renderWithProviders(<MinibridgePage />);
      expect(await screen.findByTestId('mb-round-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('advances to the next deal when the button is pressed', async () => {
    mockExec.mockResolvedValue(playing({ phase: 2 } as Partial<MinibridgeResponse>));
    renderWithProviders(<MinibridgePage />);

    const btn = await screen.findByRole('button', { name: '次のディールへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<MinibridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerTeam, expected] of [
      [0, /あなたのペアの勝ち/],
      [1, /相手ペアの勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(playing({ gameEndFlag: true, phase: 3, winnerTeam }));
      const { unmount } = renderWithProviders(<MinibridgePage />);
      expect(await screen.findByTestId('mb-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'dummy-1', reason: 'hint.minibridgeDummy', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<MinibridgePage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/ダミーの手番/);
  });
});
