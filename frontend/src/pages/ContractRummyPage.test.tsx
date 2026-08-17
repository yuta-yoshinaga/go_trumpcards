import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { contractrummyApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ContractRummyResponse } from '../types/card';
import { ContractRummyPage } from './ContractRummyPage';

vi.mock('../api/gameApi', () => ({
  contractrummyApi: { exec: vi.fn() },
  actionLogApi: { contractrummy: vi.fn() },
}));

const mockExec = vi.mocked(contractrummyApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseHand: Card[] = [
  card('SPADE', 5),
  card('HEART', 5),
  card('DIAMOND', 5),
  card('CLOVER', 13),
  card('SPADE', 13),
  card('HEART', 13),
  card('SPADE', 2),
  card('SPADE', 3),
  card('SPADE', 4),
  card('SPADE', 6),
  card('SPADE', 7),
];

const drawState: ContractRummyResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 11,
      cards: baseHand,
      melds: [],
      contractMet: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 11,
      cards: [],
      melds: [],
      contractMet: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 11,
      cards: [],
      melds: [],
      contractMet: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  totalRounds: 7,
  currentPlayerIdx: 0,
  discardTop: card('HEART', 7),
  drawPileCount: 60,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  contractSlots: [
    { kind: 0, size: 3 },
    { kind: 0, size: 3 },
  ],
  config: { cpuDifficulty: 1, failContractPenalty: 25 },
  message: '',
};

const playState: ContractRummyResponse = { ...drawState, phase: 1 };

const roundEndState: ContractRummyResponse = {
  ...drawState,
  phase: 2,
  roundWinnerIdx: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(drawState);
});

describe('ContractRummyPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the round and contract banner', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByText(/Round 1 \/ 7|ラウンド 1 \/ 7/)).toBeInTheDocument());
  });

  it('shows draw-phase action buttons during human draw turn', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument();
  });

  it('invokes drawstock when the stock button is clicked', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: /Draw from stock|山札から引く/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows play-phase action buttons after drawing', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Submit contract|コントラクトを場に出す/ })).toBeInTheDocument(),
    );
  });

  it('exposes hand cards as accessible buttons with aria-pressed selection state', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    // cardAlt('DIAMOND', 5) => "♦ 5"
    const cardBtn = await screen.findByRole('button', { name: '♦ 5' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '♦ 5' })).toHaveAttribute('aria-pressed', 'true'));
  });

  it('shows next-round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Next round|次のラウンドへ/ })).toBeInTheDocument());
  });

  it('invokes drawdiscard when the discard button is clicked', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /Take discard|捨て札を取る/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('lets the human stage cards into a contract slot and submit it', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add to slot|スロットに追加/ })).toBeInTheDocument());

    // Select first three cards (the three 5s) by clicking their hand buttons.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);

    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));

    // Stage second slot (next three buttons should be Ks).
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[4]);
    fireEvent.click(cardButtons[5]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));

    fireEvent.click(screen.getByRole('button', { name: /Submit contract|コントラクトを場に出す/ }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'meldcontract',
        expect.objectContaining({ indicesPerSlot: expect.any(Array) }),
      ),
    );
  });

  it('undoing a slot pop returns the cards to the hand', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add to slot|スロットに追加/ })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));

    const undoBtn = screen.getByRole('button', { name: /Undo last slot|最後のスロットを取り消す/ });
    expect(undoBtn).not.toBeDisabled();
    fireEvent.click(undoBtn);
    // After undo, the undo button is disabled again (no slots staged).
    expect(undoBtn).toBeDisabled();
  });

  it('removes a single staged card from its slot without clearing the rest', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add to slot|スロットに追加/ })).toBeInTheDocument());

    // Stage the three 5s (indices 0-2) into slot 0.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress-0')).toHaveAttribute('data-state', 'satisfied'));

    // Remove just the first staged card — the slot keeps the other two.
    fireEvent.click(screen.getByTestId('cr-slot-card-0'));
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress-0')).toHaveAttribute('data-state', 'partial'));
    expect(screen.getByTestId('cr-slot-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('cr-slot-card-2')).toBeInTheDocument();
    // The slot is not deleted, so slotsBuilt still reports one staged slot.
    expect(screen.getByText(/Slots staged|組成済スロット/)).toBeInTheDocument();
  });

  it('makes a removed staged card selectable again', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add to slot|スロットに追加/ })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('cr-slot-card-0')).toBeInTheDocument());

    // Remove the ♠5 (index 0) from the slot; it returns to the hand as a plain, selectable card.
    fireEvent.click(screen.getByTestId('cr-slot-card-0'));
    const reselectable = await screen.findByRole('button', { name: '♠ 5' });
    expect(reselectable).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(reselectable);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 5' })).toHaveAttribute('aria-pressed', 'true'));
  });

  it('auto-deletes a staged slot once its last card is removed', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add to slot|スロットに追加/ })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByText(/Slots staged|組成済スロット/)).toBeInTheDocument());

    // Remove all three staged cards; the emptied slot disappears.
    fireEvent.click(screen.getByTestId('cr-slot-card-0'));
    fireEvent.click(screen.getByTestId('cr-slot-card-1'));
    fireEvent.click(screen.getByTestId('cr-slot-card-2'));
    await waitFor(() => expect(screen.queryByText(/Slots staged|組成済スロット/)).not.toBeInTheDocument());
    expect(screen.getByTestId('cr-slot-progress-0')).toHaveAttribute('data-state', 'empty');
  });

  it('toggles a card off when clicked twice', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Add to slot|スロットに追加/ })).toBeInTheDocument());

    const addSlot = screen.getByRole('button', { name: /Add to slot|スロットに追加/ });
    expect(addSlot).toBeDisabled();

    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    expect(addSlot).not.toBeDisabled();
    fireEvent.click(cardButtons[0]); // toggle off
    expect(addSlot).toBeDisabled();
  });

  it('shows discard button after card is selected in play phase', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Discard card|カードを捨てる/ })).toBeInTheDocument(),
    );

    const discardBtn = screen.getByRole('button', { name: /Discard card|カードを捨てる/ });
    expect(discardBtn).toBeDisabled(); // no card selected

    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    expect(discardBtn).not.toBeDisabled();
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 0 }));
  });

  it('shows extra-meld + layoff buttons when contract is met', async () => {
    const metState: ContractRummyResponse = {
      ...playState,
      players: [
        {
          ...playState.players[0],
          contractMet: true,
          melds: [{ cards: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5)] }],
        },
        playState.players[1],
        playState.players[2],
      ],
    };
    mockExec.mockResolvedValue(metState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Lay extra meld|追加メルド/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /Lay off|レイオフ/ })).toBeInTheDocument();
  });

  it('selecting an opponent meld highlights it and shows the layoff target in the footer', async () => {
    const metState: ContractRummyResponse = {
      ...playState,
      players: [
        { ...playState.players[0], contractMet: true },
        {
          ...playState.players[1],
          contractMet: true,
          melds: [{ cards: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5)] }],
        },
        playState.players[2],
      ],
    };
    mockExec.mockResolvedValue(metState);
    renderWithProviders(<ContractRummyPage />);
    const meldButton = await screen.findByRole('button', { name: /メルド1/ });
    fireEvent.click(meldButton);
    expect(meldButton).toHaveAttribute('aria-pressed', 'true');
    expect(meldButton).toHaveClass('bg-ds-warning/20');
    expect(screen.getByTestId('cr-layoff-target')).toBeInTheDocument();
  });

  it('does not let you target an opponent meld before your own contract is met', async () => {
    const notMetState: ContractRummyResponse = {
      ...playState,
      players: [
        { ...playState.players[0], contractMet: false }, // human has NOT met their contract
        {
          ...playState.players[1],
          contractMet: true,
          melds: [{ cards: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5)] }],
        },
        playState.players[2],
      ],
    };
    mockExec.mockResolvedValue(notMetState);
    renderWithProviders(<ContractRummyPage />);
    const meldButton = await screen.findByRole('button', { name: /メルド1/ });
    // Not actionable yet: no toggle semantics and clicking does nothing.
    expect(meldButton).not.toHaveAttribute('aria-pressed');
    fireEvent.click(meldButton);
    expect(meldButton).not.toHaveClass('bg-ds-warning/20');
    expect(screen.queryByTestId('cr-layoff-target')).not.toBeInTheDocument();
  });

  it('shows per-slot progress and only enables Submit when both slots satisfy their contract', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress')).toBeInTheDocument());
    // Both slots start empty.
    expect(screen.getByTestId('cr-slot-progress-0')).toHaveAttribute('data-state', 'empty');
    expect(screen.getByTestId('cr-slot-progress-1')).toHaveAttribute('data-state', 'empty');
    expect(screen.getByTestId('cr-submit-contract')).toBeDisabled();

    // Fill slot 0 with the three 5s — valid set.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress-0')).toHaveAttribute('data-state', 'satisfied'));
    expect(screen.getByTestId('cr-submit-contract')).toBeDisabled();

    // Fill slot 1 with the three Ks — valid set.
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[4]);
    fireEvent.click(cardButtons[5]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress-1')).toHaveAttribute('data-state', 'satisfied'));
    expect(screen.getByTestId('cr-submit-contract')).not.toBeDisabled();
  });

  it('flags an invalid set as invalid in the slot progress chip', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress-0')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    // Three mismatched ranks (5, K, 2) — not a valid set.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[6]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('cr-slot-progress-0')).toHaveAttribute('data-state', 'invalid'));
    expect(screen.getByTestId('cr-submit-contract')).toBeDisabled();
  });

  it('shows winner banner at game end', async () => {
    const endState: ContractRummyResponse = {
      ...drawState,
      phase: 3,
      gameEndFlag: true,
      winnerIdx: 0,
    };
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<ContractRummyPage />);
    // Reset becomes "Next Game" at game end.
    await waitFor(() => expect(screen.getByRole('button', { name: /Next Game|次のゲーム/ })).toBeInTheDocument());
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<ContractRummyPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('refuses an extra meld that is not a set or a run', async () => {
    const metState: ContractRummyResponse = {
      ...playState,
      players: [{ ...playState.players[0], contractMet: true }, playState.players[1], playState.players[2]],
    };
    mockExec.mockResolvedValue(metState);
    renderWithProviders(<ContractRummyPage />);
    const meldExtra = await screen.findByRole('button', { name: /Lay extra meld|追加メルド/ });
    // Three cards, but neither a set nor a run — counting to three used to be enough.
    fireEvent.click(screen.getByRole('button', { name: '♠ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ K' }));
    expect(meldExtra).toBeDisabled();
    expect(screen.getByTestId('cr-invalid-extra-meld')).toBeInTheDocument();
  });

  it('allows an extra meld that is a real set', async () => {
    const metState: ContractRummyResponse = {
      ...playState,
      players: [{ ...playState.players[0], contractMet: true }, playState.players[1], playState.players[2]],
    };
    mockExec.mockResolvedValue(metState);
    renderWithProviders(<ContractRummyPage />);
    const meldExtra = await screen.findByRole('button', { name: /Lay extra meld|追加メルド/ });
    fireEvent.click(screen.getByRole('button', { name: '♠ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 5' }));
    fireEvent.click(screen.getByRole('button', { name: '♦ 5' }));
    expect(meldExtra).toBeEnabled();
    expect(screen.queryByTestId('cr-invalid-extra-meld')).not.toBeInTheDocument();
  });

  // #5588: 難易度はドメインにもレスポンス型にも API にもあるのに、**Web からは
  // 触れなかった** (CUI の `sd` だけ)。
  describe('cpu difficulty setting', () => {
    it('shows the current difficulty from the response', async () => {
      mockExec.mockResolvedValue({ ...drawState, config: { cpuDifficulty: 2, failContractPenalty: 25 } });
      renderWithProviders(<ContractRummyPage />);
      const select = await screen.findByLabelText('CPU難易度');
      expect((select as HTMLSelectElement).value).toBe('2');
    });

    // **数値は CUI の `sd` と同じ 0/1/2** (受け入れ条件2)。reset に載せて渡す ──
    // 難易度は配り直しでしか効かない。
    it('resets with the chosen difficulty', async () => {
      mockExec.mockResolvedValue(drawState);
      renderWithProviders(<ContractRummyPage />);
      const select = await screen.findByLabelText('CPU難易度');
      mockExec.mockClear();
      fireEvent.change(select, { target: { value: '0' } });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 0 } }));
    });

    it('offers exactly the three levels', async () => {
      mockExec.mockResolvedValue(drawState);
      renderWithProviders(<ContractRummyPage />);
      const select = await screen.findByLabelText('CPU難易度');
      const values = [...(select as HTMLSelectElement).options].map((o) => o.value);
      expect(values).toEqual(['0', '1', '2']);
    });
  });
});
