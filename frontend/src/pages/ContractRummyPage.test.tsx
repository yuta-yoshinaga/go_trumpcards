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
});
