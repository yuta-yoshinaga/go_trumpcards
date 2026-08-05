import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cariocaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CariocaResponse } from '../types/card';
import { CariocaPage } from './CariocaPage';

vi.mock('../api/gameApi', () => ({
  cariocaApi: { exec: vi.fn() },
  actionLogApi: { carioca: vi.fn() },
}));

const mockExec = vi.mocked(cariocaApi.exec);

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

const drawState: CariocaResponse = {
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
  config: { playerCount: 3, cpuDifficulty: 1, failContractPenalty: 25 },
  message: '',
};

const playState: CariocaResponse = { ...drawState, phase: 1 };

const roundEndState: CariocaResponse = {
  ...drawState,
  phase: 2,
  roundWinnerIdx: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(drawState);
});

describe('CariocaPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the round and contract banner', async () => {
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByText(/Round 1 \/ 7|ラウンド 1 \/ 7/)).toBeInTheDocument());
  });

  it('shows draw-phase action buttons during human draw turn', async () => {
    renderWithProviders(<CariocaPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument();
  });

  it('invokes drawstock when the stock button is clicked', async () => {
    renderWithProviders(<CariocaPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: /Draw from stock|山札から引く/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows play-phase action buttons after drawing', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Submit contract|コントラクトを場に出す/ })).toBeInTheDocument(),
    );
  });

  it('shows next-round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Next round|次のラウンドへ/ })).toBeInTheDocument());
  });

  it('invokes drawdiscard when the discard button is clicked', async () => {
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /Take discard|捨て札を取る/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('lets the human stage cards into a contract slot and submit it', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CariocaPage />);
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
    renderWithProviders(<CariocaPage />);
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

  it('toggles a card off when clicked twice', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CariocaPage />);
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
    renderWithProviders(<CariocaPage />);
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
    const metState: CariocaResponse = {
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
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Lay extra meld|追加メルド/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /Lay off|レイオフ/ })).toBeInTheDocument();
  });

  it('selecting an opponent meld highlights it and shows the layoff target in the footer', async () => {
    const metState: CariocaResponse = {
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
    renderWithProviders(<CariocaPage />);
    const meldButton = await screen.findByRole('button', { name: /メルド1/ });
    fireEvent.click(meldButton);
    expect(meldButton).toHaveAttribute('aria-pressed', 'true');
    expect(meldButton).toHaveClass('bg-ds-warning/20');
    expect(screen.getByTestId('ca-layoff-target')).toBeInTheDocument();
  });

  it('does not let you target an opponent meld before your own contract is met', async () => {
    const notMetState: CariocaResponse = {
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
    renderWithProviders(<CariocaPage />);
    const meldButton = await screen.findByRole('button', { name: /メルド1/ });
    // Not actionable yet: no toggle semantics and clicking does nothing.
    expect(meldButton).not.toHaveAttribute('aria-pressed');
    fireEvent.click(meldButton);
    expect(meldButton).not.toHaveClass('bg-ds-warning/20');
    expect(screen.queryByTestId('ca-layoff-target')).not.toBeInTheDocument();
  });

  it('shows per-slot progress and only enables Submit when both slots satisfy their contract', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress')).toBeInTheDocument());
    // Both slots start empty.
    expect(screen.getByTestId('ca-slot-progress-0')).toHaveAttribute('data-state', 'empty');
    expect(screen.getByTestId('ca-slot-progress-1')).toHaveAttribute('data-state', 'empty');
    expect(screen.getByTestId('ca-submit-contract')).toBeDisabled();

    // Fill slot 0 with the three 5s — valid set.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-0')).toHaveAttribute('data-state', 'satisfied'));
    expect(screen.getByTestId('ca-submit-contract')).toBeDisabled();

    // Fill slot 1 with the three Ks — valid set.
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[4]);
    fireEvent.click(cardButtons[5]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-1')).toHaveAttribute('data-state', 'satisfied'));
    expect(screen.getByTestId('ca-submit-contract')).not.toBeDisabled();
  });

  it('annotates each slot with what is still missing and clears it once satisfied', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress')).toBeInTheDocument());
    // An empty slot shows the "not started" hint.
    expect(screen.getByTestId('ca-slot-shortfall-0')).toHaveTextContent('未着手');

    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    // Place a single 5 into slot 0 — a set of 3 still needs 2 more of the same rank.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-slot-shortfall-0')).toHaveTextContent('あと2枚 同ランク'));

    // Complete slot 0 with the three 5s — the shortfall annotation disappears.
    fireEvent.click(screen.getByRole('button', { name: /Undo last slot|最後のスロットを取り消す/ }));
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-0')).toHaveAttribute('data-state', 'satisfied'));
    expect(screen.queryByTestId('ca-slot-shortfall-0')).not.toBeInTheDocument();
  });

  it('enables Submit for a joker-wild contract meld', async () => {
    // Slot 0 = 5,5,JOKER (a set completed by a wild); slot 1 = three 13s.
    const jokerHand: Card[] = [
      card('SPADE', 5),
      card('HEART', 5),
      { design: 'JOKER', value: 0 },
      card('CLOVER', 13),
      card('SPADE', 13),
      card('HEART', 13),
    ];
    const jokerState: CariocaResponse = {
      ...playState,
      players: [
        { ...playState.players[0], cards: jokerHand, cardCount: jokerHand.length },
        playState.players[1],
        playState.players[2],
      ],
    };
    mockExec.mockResolvedValue(jokerState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-0')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));

    // Slot 0: 5, 5, joker — a set only because the joker acts as a wildcard.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-0')).toHaveAttribute('data-state', 'satisfied'));
    expect(screen.getByTestId('ca-submit-contract')).toBeDisabled();

    // Slot 1: three 13s.
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[4]);
    fireEvent.click(cardButtons[5]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-submit-contract')).not.toBeDisabled());
  });

  it('flags an invalid set as invalid in the slot progress chip', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-0')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.querySelector('img'));
    // Three mismatched ranks (5, K, 2) — not a valid set.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[6]);
    fireEvent.click(screen.getByRole('button', { name: /Add to slot|スロットに追加/ }));
    await waitFor(() => expect(screen.getByTestId('ca-slot-progress-0')).toHaveAttribute('data-state', 'invalid'));
    expect(screen.getByTestId('ca-submit-contract')).toBeDisabled();
  });

  it('marks a meld green for a card that fits and red for one that does not', async () => {
    // Human contract met (no own melds) + opponent 1 has a trío of 5s to lay off onto.
    const metState: CariocaResponse = {
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
    renderWithProviders(<CariocaPage />);
    const meld = await screen.findByTestId('ca-meld-1-0');
    // No preview before a card is staged.
    expect(meld).not.toHaveAttribute('data-layoff-accepts');

    // Hand card buttons carry an image but (unlike meld buttons) no ca-meld test id.
    const handCards = screen
      .getAllByRole('button')
      .filter((b) => b.querySelector('img') && !b.getAttribute('data-testid')?.startsWith('ca-meld'));

    // baseHand[0] is the 5♠ — a valid fourth card for the 5-set.
    fireEvent.click(handCards[0]);
    expect(meld).toHaveAttribute('data-layoff-accepts', 'true');
    expect(meld).toHaveClass('ring-ds-success');

    // Toggle it off and pick baseHand[3] (K♣) — a mismatch that the set rejects.
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[3]);
    expect(meld).toHaveAttribute('data-layoff-accepts', 'false');
    expect(meld).toHaveClass('ring-ds-error');
  });

  it('annotates the layoff summary with whether the staged card fits the target meld', async () => {
    const metState: CariocaResponse = {
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
    renderWithProviders(<CariocaPage />);
    const meld = await screen.findByTestId('ca-meld-1-0');
    fireEvent.click(meld); // target this meld

    const handCards = screen
      .getAllByRole('button')
      .filter((b) => b.querySelector('img') && !b.getAttribute('data-testid')?.startsWith('ca-meld'));

    // Stage the 5♠ — the summary should report the card fits.
    fireEvent.click(handCards[0]);
    const acceptance = screen.getByTestId('ca-layoff-acceptance');
    expect(acceptance).toHaveAttribute('data-accepts', 'true');

    // Switch to the K♣ — the summary flips to "doesn't fit".
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[3]);
    expect(screen.getByTestId('ca-layoff-acceptance')).toHaveAttribute('data-accepts', 'false');
  });

  it('shows winner banner at game end', async () => {
    const endState: CariocaResponse = {
      ...drawState,
      phase: 3,
      gameEndFlag: true,
      winnerIdx: 0,
    };
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<CariocaPage />);
    // Reset becomes "Next Game" at game end.
    await waitFor(() => expect(screen.getByRole('button', { name: /Next Game|次のゲーム/ })).toBeInTheDocument());
  });

  // **押すまで出さない。**押して初めて `hintFactories` の登録行とページ側の
  // 配線が走る。これが無いと「実装したがページに繋いでいない」に気づけない
  // (#4644 のレビュー指摘がまさにそれ)。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<CariocaPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  // **CUI プレゼンターがあるのに Web からその表現へ行けなかった (#4849)。**
  it('switches to CLI mode from the header toggle', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<CariocaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));

    // ターミナルが出て、GUI の操作行は消える。
    expect(await screen.findByRole('textbox')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });
});
