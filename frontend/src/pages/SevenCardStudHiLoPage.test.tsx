import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevenCardStudHiLoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevenCardStudPlayerData, SevenCardStudResponse } from '../types/card';
import { SevenCardStudPhase } from '../types/phases';
import { SevenCardStudHiLoPage } from './SevenCardStudHiLoPage';

vi.mock('../api/gameApi', () => ({
  sevenCardStudApi: { exec: vi.fn() },
  sevenCardStudHiLoApi: { exec: vi.fn() },
  actionLogApi: { sevencardstudhilo: vi.fn() },
}));

const mockExec = vi.mocked(sevenCardStudHiLoApi.exec);

const card = (design: 'SPADE' | 'HEART' | 'CLOVER' | 'DIAMOND', value: number) => ({ design, value });

function seat(id: number, isHuman: boolean): SevenCardStudPlayerData {
  return {
    id,
    isHuman,
    name: isHuman ? 'あなた' : `CPU ${id.toString()}`,
    chips: 1000,
    currentBet: 0,
    folded: false,
    allIn: false,
    holeCards: [card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3)],
    doorCards: [card('CLOVER', 4), card('SPADE', 5), card('HEART', 12), card('DIAMOND', 13)],
    handName: '',
    totalHands: 0,
  } as unknown as SevenCardStudPlayerData;
}

function makeState(overrides?: Partial<SevenCardStudResponse>): SevenCardStudResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    communityCard: null,
    pot: 400,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: 5,
    isHiLo: true,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 20,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 1000,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 10,
    bringIn: 20,
    smallBet: 20,
    bigBet: 40,
    tournamentMode: false,
    anteLevelHands: 0,
    anteMultiplier: 1,
    tableSize: 2,
    bringInPlayerIdx: 0,
    rebuyAvailable: false,
    addonAvailable: false,
    rebuyCounts: [0, 0],
    addonUsed: [false, false],
    rebuyEnabled: false,
    addonEnabled: false,
    rebuyMaxCount: 0,
    rebuyChips: 0,
    addonChips: 0,
    message: '',
    ...overrides,
  } as SevenCardStudResponse;
}

// SHOWDOWN is 6, not 5 -- 5 is SEVENTH_STREET, where no result is rendered.
const showdown = (results: SevenCardStudResponse['roundResults']) =>
  makeState({ phase: SevenCardStudPhase.SHOWDOWN, roundResults: results });

describe('SevenCardStudHiLoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('drives the hi-lo endpoint, not the plain stud one', async () => {
    // 同じページを共有しているので、キーの取り違えは静かに別ゲームを叩く。
    renderWithProviders(<SevenCardStudHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('shows the split breakdown at showdown', async () => {
    mockExec.mockResolvedValue(
      showdown([
        { playerIdx: 0, handRank: 2, handName: 'Two Pair', kickers: '', bestHand: [], wonAmount: 201, mucked: false },
        {
          playerIdx: 1,
          handRank: 0,
          handName: 'High Card',
          kickers: '',
          bestHand: [],
          wonAmount: 200,
          mucked: false,
          wonLow: 200,
          lowQualifies: true,
          lowBestHand: [card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3), card('CLOVER', 4), card('SPADE', 5)],
        },
      ] as SevenCardStudResponse['roundResults']),
    );
    renderWithProviders(<SevenCardStudHiLoPage />);

    await waitFor(() => expect(screen.getByTestId('studhilo-split')).toBeInTheDocument());
    // 奇数チップはハイ側。201/200 がそのまま出ていること。
    expect(screen.getByTestId('studhilo-hi-badge')).toHaveTextContent('201');
    expect(screen.getByTestId('studhilo-lo-badge')).toHaveTextContent('200');
  });

  it('states that the high took it all when no low qualified', async () => {
    mockExec.mockResolvedValue(
      showdown([
        { playerIdx: 0, handRank: 2, handName: 'Two Pair', kickers: '', bestHand: [], wonAmount: 400, mucked: false },
      ] as SevenCardStudResponse['roundResults']),
    );
    renderWithProviders(<SevenCardStudHiLoPage />);

    await waitFor(() => expect(screen.getByTestId('studhilo-hi-takes-all')).toBeInTheDocument());
    expect(screen.queryByTestId('studhilo-lo-badge')).not.toBeInTheDocument();
  });

  it('does not render the split for a plain stud response', async () => {
    // isHiLo はサーバーが立てる。ページがルート名から推測すると、通常の
    // スタッドにも空の内訳が出る。
    mockExec.mockResolvedValue(
      makeState({
        isHiLo: false,
        phase: SevenCardStudPhase.SHOWDOWN,
        roundResults: [
          { playerIdx: 0, handRank: 2, handName: 'Two Pair', kickers: '', bestHand: [], wonAmount: 400, mucked: false },
        ] as SevenCardStudResponse['roundResults'],
      }),
    );
    renderWithProviders(<SevenCardStudHiLoPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('studhilo-split')).not.toBeInTheDocument();
  });
});
