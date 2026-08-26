import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { chicagoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevenCardStudPlayerData, SevenCardStudResponse } from '../types/card';
import { SevenCardStudPhase } from '../types/phases';
import { ChicagoPage } from './ChicagoPage';

vi.mock('../api/gameApi', () => ({
  sevenCardStudApi: { exec: vi.fn() },
  chicagoApi: { exec: vi.fn() },
  actionLogApi: { chicago: vi.fn() },
}));

const mockExec = vi.mocked(chicagoApi.exec);

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
    isChicago: true,
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

describe('ChicagoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('drives the chicago endpoint, not the plain stud one', async () => {
    // 同じページを共有しているので、キーの取り違えは静かに別ゲームを叩く。
    renderWithProviders(<ChicagoPage />);
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
          wonSpade: 200,
          spadeCard: card('SPADE', 1),
        },
      ] as SevenCardStudResponse['roundResults']),
    );
    renderWithProviders(<ChicagoPage />);

    await waitFor(() => expect(screen.getByTestId('studchicago-split')).toBeInTheDocument());
    // 奇数チップは役の側。201/200 がそのまま出ていること。
    expect(screen.getByTestId('studchicago-hi-badge')).toHaveTextContent('201');
    const spade = screen.getByTestId('studchicago-spade-badge');
    expect(spade).toHaveTextContent('200');
    // **どの 1 枚で半分を取ったのかを出す。**
    expect(spade).toHaveTextContent('♠');
  });

  it('states that the high took it all when nobody held a spade in the hole', async () => {
    mockExec.mockResolvedValue(
      showdown([
        { playerIdx: 0, handRank: 2, handName: 'Two Pair', kickers: '', bestHand: [], wonAmount: 400, mucked: false },
      ] as SevenCardStudResponse['roundResults']),
    );
    renderWithProviders(<ChicagoPage />);

    await waitFor(() => expect(screen.getByTestId('studchicago-hi-takes-all')).toBeInTheDocument());
    expect(screen.queryByTestId('studchicago-spade-badge')).not.toBeInTheDocument();
  });

  it('does not render the split for a plain stud response', async () => {
    // isChicago はサーバーが立てる。ページがルート名から推測すると、通常の
    // スタッドにも空の内訳が出る。
    mockExec.mockResolvedValue(
      makeState({
        isChicago: false,
        phase: SevenCardStudPhase.SHOWDOWN,
        roundResults: [
          { playerIdx: 0, handRank: 2, handName: 'Two Pair', kickers: '', bestHand: [], wonAmount: 400, mucked: false },
        ] as SevenCardStudResponse['roundResults'],
      }),
    );
    renderWithProviders(<ChicagoPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('studchicago-split')).not.toBeInTheDocument();
  });
});

// #5522 と同型: Chicago はページを共有するが i18n の名前空間は別。style.* が
// 解決しないと生キーが画面に出る。
describe('ChicagoPage HUD stats', () => {
  it('shows the CPU HUD with translated style names', async () => {
    const cpu = {
      ...seat(1, false),
      totalHands: 20,
      vpip: 42,
      pfr: 8,
      threeBet: 3,
      af: '2.5',
    } as unknown as SevenCardStudPlayerData;
    mockExec.mockResolvedValue(makeState({ players: [seat(0, true), cpu] }));
    renderWithProviders(<ChicagoPage />);
    const hud = await screen.findByTestId('hud-stats');
    expect(hud).toHaveTextContent('42%');
    const style = screen.getByTestId('hud-overall-style');
    expect(style).toHaveTextContent('LP');
    expect(style.textContent).not.toContain('style.');
  });
});
