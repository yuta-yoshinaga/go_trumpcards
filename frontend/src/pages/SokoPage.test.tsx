import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fiveCardStudApi, sokoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { FiveCardStudResponse } from '../types/card';
import { FiveCardStudPhase } from '../types/phases';
import { SokoPage } from './SokoPage';

vi.mock('../api/gameApi', () => ({
  fiveCardStudApi: { exec: vi.fn() },
  sokoApi: { exec: vi.fn() },
  actionLogApi: { soko: vi.fn() },
}));

const mockSoko = vi.mocked(sokoApi.exec);
const mockStud = vi.mocked(fiveCardStudApi.exec);

const card = (design: 'SPADE' | 'HEART' | 'CLOVER' | 'DIAMOND', value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, handName = '') {
  return {
    id,
    isHuman,
    name: isHuman ? 'あなた' : `CPU ${id.toString()}`,
    chips: 1000,
    currentBet: 0,
    folded: false,
    allIn: false,
    // ♠2 ♠5 ♠9 ♠K ♥K -- a pair of kings that is also four spades, i.e. the hand
    // Soko ranks as a four-card flush and plain stud ranks as one pair.
    holeCards: [card('SPADE', 2)],
    doorCards: [card('SPADE', 5), card('SPADE', 9), card('SPADE', 13), card('HEART', 13)],
    handName,
    totalHands: 0,
  };
}

function makeState(overrides?: Partial<FiveCardStudResponse>): FiveCardStudResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    communityCard: null,
    pot: 400,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: FiveCardStudPhase.FIFTH_STREET,
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
  } as unknown as FiveCardStudResponse;
}

describe('SokoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSoko.mockResolvedValue(makeState());
    mockStud.mockResolvedValue(makeState());
  });

  // The page is shared with Five Card Stud, so getting the key wrong silently
  // drives the other game's endpoint — no error, just the wrong session.
  it('drives the soko endpoint, not the plain five card stud one', async () => {
    renderWithProviders(<SokoPage />);
    await waitFor(() => expect(mockSoko).toHaveBeenCalled());
    expect(mockStud).not.toHaveBeenCalled();
  });

  it('renders the Soko heading rather than Five Card Stud', async () => {
    renderWithProviders(<SokoPage />);
    // Heading role, matching the e2e spec: this render has no NavBar so plain
    // text would pass, but the assertion should say the same thing in both.
    await waitFor(() => expect(screen.getByRole('heading', { name: 'ソッコ' })).toBeInTheDocument());
    expect(screen.queryByRole('heading', { name: 'ファイブカードスタッド' })).not.toBeInTheDocument();
  });

  // The Soko-only ranks are resolved server-side into handName, so the page's
  // job is only to show what it is given. This pins that it does.
  it('shows the four-card flush hand name the server sends at showdown', async () => {
    mockSoko.mockResolvedValue(
      makeState({
        phase: FiveCardStudPhase.SHOWDOWN,
        players: [seat(0, true, 'Four-Card Flush'), seat(1, false, 'One Pair')],
      } as unknown as Partial<FiveCardStudResponse>),
    );
    renderWithProviders(<SokoPage />);
    await waitFor(() => expect(screen.getByText('Four-Card Flush')).toBeInTheDocument());
  });

  it('renders the shared Five Card Stud controls', async () => {
    renderWithProviders(<SokoPage />);
    await waitFor(() => expect(screen.getByText('ソッコ')).toBeInTheDocument());
    // The testid still carries the Five Card Stud name because the component IS
    // the Five Card Stud page — that is the point of sharing it.
    expect(screen.getByTestId('five-card-stud-kbd-shortcuts')).toBeInTheDocument();
  });
});
