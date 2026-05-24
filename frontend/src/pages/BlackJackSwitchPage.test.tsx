import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackjackswitchApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackJackSwitchResponse } from '../types/card';
import { BlackJackSwitchPhase, BlackJackSwitchResult } from '../types/phases';
import { BlackJackSwitchPage } from './BlackJackSwitchPage';

vi.mock('../api/gameApi', () => ({
  blackjackswitchApi: { exec: vi.fn() },
  actionLogApi: { blackjackswitch: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(blackjackswitchApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const betState: BlackJackSwitchResponse = {
  hands: [],
  dealerCards: [],
  dealerScore: 0,
  phase: BlackJackSwitchPhase.BET,
  currentHandIdx: 0,
  chips: 1000,
  switched: false,
  dealerPushed22: false,
  overallResult: 0,
  totalPayout: 0,
  message: '',
};

const switchState: BlackJackSwitchResponse = {
  hands: [
    {
      cards: [
        { design: 'SPADE', value: 10 },
        { design: 'CLOVER', value: 5 },
      ],
      score: 15,
      bet: 100,
      stood: false,
      doubled: false,
      busted: false,
      isBJ: false,
      result: 0,
      payout: 0,
    },
    {
      cards: [
        { design: 'HEART', value: 6 },
        { design: 'DIAMOND', value: 11 },
      ],
      score: 16,
      bet: 100,
      stood: false,
      doubled: false,
      busted: false,
      isBJ: false,
      result: 0,
      payout: 0,
    },
  ],
  dealerCards: [{ design: 'SPADE', value: 10 }, null],
  dealerScore: 10,
  phase: BlackJackSwitchPhase.SWITCH,
  currentHandIdx: 0,
  chips: 800,
  switched: false,
  dealerPushed22: false,
  overallResult: 0,
  totalPayout: 0,
  message: '',
};

const actionState: BlackJackSwitchResponse = {
  ...switchState,
  phase: BlackJackSwitchPhase.ACTION,
};

const dealer22EndState: BlackJackSwitchResponse = {
  ...switchState,
  hands: switchState.hands.map((h) => ({ ...h, stood: true, result: 0, payout: 100 })),
  dealerCards: [
    { design: 'SPADE', value: 5 },
    { design: 'CLOVER', value: 7 },
    { design: 'DIAMOND', value: 10 },
  ],
  dealerScore: 22,
  phase: BlackJackSwitchPhase.END,
  dealerPushed22: true,
  overallResult: 0,
  totalPayout: 200,
  message: 'Dealer 22: hands pushed.',
  messageCode: 'blackjackswitch.result.dealer22Push',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('BlackJackSwitchPage', () => {
  it('issues a reset on mount', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<BlackJackSwitchPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows the bet button in BET phase and posts a bet on click', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<BlackJackSwitchPage />);
    const betButton = await screen.findByRole('button', { name: /Place Bet|ベットする/ });
    fireEvent.click(betButton);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', expect.any(Number)));
  });

  it('shows Switch and Keep buttons in SWITCH phase', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByRole('button', { name: /Switch|スイッチ/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Keep|そのまま/ })).toBeInTheDocument();
  });

  it('hides the dealer hole card before END phase', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    await screen.findByTestId('dealer-area');
    // Two dealer cards rendered, one of them is the face-down placeholder.
    expect(screen.getAllByTestId('card-back').length).toBeGreaterThanOrEqual(1);
  });

  it('renders Hit / Stand / Double Down in ACTION phase', async () => {
    mockApi.mockResolvedValue(actionState);
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByRole('button', { name: /Hit|ヒット/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Stand|スタンド/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Double Down|ダブルダウン/ })).toBeInTheDocument();
  });

  it('shows the dealer-22 push banner in END phase', async () => {
    mockApi.mockResolvedValue(dealer22EndState);
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('dealer22-banner')).toBeInTheDocument();
  });

  it('renders the payout breakdown in END phase', async () => {
    mockApi.mockResolvedValue(dealer22EndState);
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('clicks Switch and posts the switch command', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByRole('button', { name: /Switch|スイッチ/ });
    fireEvent.click(btn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('switch'));
  });

  it('hovering Switch shows post-swap score previews on each hand', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByTestId('switch-button');
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
    fireEvent.mouseEnter(btn);
    // Hand 0 = [S10, C5]=15 -> swap -> [S10, D11]=20; Hand 1 = [H6, D11]=16 -> swap -> [H6, C5]=11
    expect(screen.getByTestId('hand-0-preview')).toHaveTextContent('20');
    expect(screen.getByTestId('hand-1-preview')).toHaveTextContent('11');
    fireEvent.mouseLeave(btn);
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
  });

  it('clicks Keep and posts the keep command', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByRole('button', { name: /Keep|そのまま/ });
    fireEvent.click(btn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('keep'));
  });

  it('clicks Hit and posts the hit command', async () => {
    mockApi.mockResolvedValue(actionState);
    renderWithProviders(<BlackJackSwitchPage />);
    fireEvent.click(await screen.findByRole('button', { name: /Hit|ヒット/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));
  });

  it('clicks Stand and posts the stand command', async () => {
    mockApi.mockResolvedValue(actionState);
    renderWithProviders(<BlackJackSwitchPage />);
    fireEvent.click(await screen.findByRole('button', { name: /Stand|スタンド/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('stand'));
  });

  it('clicks Double Down and posts the doubledown command', async () => {
    mockApi.mockResolvedValue(actionState);
    renderWithProviders(<BlackJackSwitchPage />);
    fireEvent.click(await screen.findByRole('button', { name: /Double Down|ダブルダウン/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('doubledown'));
  });

  it('renders WIN result keys in END phase', async () => {
    mockApi.mockResolvedValue({
      ...dealer22EndState,
      dealerPushed22: false,
      hands: switchState.hands.map((h) => ({ ...h, stood: true, result: BlackJackSwitchResult.WIN, payout: 200 })),
      overallResult: BlackJackSwitchResult.WIN,
      totalPayout: 400,
    });
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('renders LOSE result keys in END phase', async () => {
    mockApi.mockResolvedValue({
      ...dealer22EndState,
      dealerPushed22: false,
      hands: switchState.hands.map((h) => ({ ...h, busted: true, result: BlackJackSwitchResult.LOSE, payout: 0 })),
      overallResult: BlackJackSwitchResult.LOSE,
      totalPayout: 0,
    });
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<BlackJackSwitchPage />);
    // CLI mode replaces the GUI card-area; dealer-area should NOT render.
    await waitFor(() => expect(mockApi).toHaveBeenCalled());
    expect(screen.queryByTestId('dealer-area')).not.toBeInTheDocument();
  });
});
