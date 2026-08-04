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

  it('bet phase: pressing "b" dispatches a bet', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<BlackJackSwitchPage />);
    await screen.findByRole('button', { name: /Place Bet|ベットする/ });
    mockApi.mockClear();
    fireEvent.keyDown(document.body, { key: 'b' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', expect.any(Number)));
  });

  it('action phase: pressing "h" dispatches hit', async () => {
    mockApi.mockResolvedValue(actionState);
    renderWithProviders(<BlackJackSwitchPage />);
    await screen.findByRole('button', { name: /Hit|ヒット/ });
    mockApi.mockClear();
    fireEvent.keyDown(document.body, { key: 'h' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('hit'));
  });

  it('advertises the bet keyboard shortcut on the button', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<BlackJackSwitchPage />);
    const betBtn = await screen.findByRole('button', { name: /Place Bet|ベットする/ });
    expect(betBtn).toHaveAttribute('aria-keyshortcuts', 'b');
    expect(betBtn.querySelector('kbd')?.textContent).toBe('B');
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

  it('touching Switch shows then hides the preview for touch devices', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByTestId('switch-button');
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
    fireEvent.touchStart(btn);
    expect(screen.getByTestId('hand-0-preview')).toHaveTextContent('20');
    expect(screen.getByTestId('hand-1-preview')).toHaveTextContent('11');
    fireEvent.touchEnd(btn);
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
  });

  it('clears the preview when a touch is cancelled by the OS', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByTestId('switch-button');
    fireEvent.touchStart(btn);
    expect(screen.getByTestId('hand-0-preview')).toBeInTheDocument();
    fireEvent.touchCancel(btn);
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
  });

  it('keeps the preview visible without hover when "always preview" is enabled', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    await screen.findByTestId('switch-button');
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
    // Open the collapsed settings <details> first, as a real user would.
    fireEvent.click(screen.getByText('設定'));
    fireEvent.click(screen.getByLabelText(/常に表示|Always show/));
    // Visible without any hover/focus.
    expect(screen.getByTestId('hand-0-preview')).toHaveTextContent('20');
    expect(screen.getByTestId('hand-1-preview')).toHaveTextContent('11');
  });

  it('focusing Switch shows the preview for keyboard parity', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByTestId('switch-button');
    fireEvent.focus(btn);
    expect(screen.getByTestId('hand-0-preview')).toBeInTheDocument();
    fireEvent.blur(btn);
    expect(screen.queryByTestId('hand-0-preview')).not.toBeInTheDocument();
  });

  it('paints the bust path in red when the post-swap score exceeds 21', async () => {
    // Hand 0 = [S10, D11]=20; Hand 1 = [H10, S10]=20. After swap: hand0 = [S10, S10] = 20; hand1 = [H10, D11] = 20.
    // Bust requires score > 21. Use [S10, S10] and [H10, S2]: swapped → [S10, S2]=12 and [H10, S10]=20.
    // For a real bust path: Hand 0 = [S10, D11]=20; Hand 1 = [H10, S5]=15 → swap → hand0 [S10,S5]=15; hand1 [H10,D11]=20.
    // Use Hand 0 = [S10, S10, S5]=25 (already busts); Hand 1 = [H10, D11]=20 → swap → hand0 [S10,D11,S5]=25 (still bust).
    const bustState: BlackJackSwitchResponse = {
      ...switchState,
      hands: [
        {
          ...switchState.hands[0],
          cards: [
            { design: 'SPADE', value: 10 },
            { design: 'SPADE', value: 10 },
            { design: 'SPADE', value: 5 },
          ],
          score: 25,
        },
        switchState.hands[1],
      ],
    };
    mockApi.mockResolvedValue(bustState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByTestId('switch-button');
    fireEvent.mouseEnter(btn);
    const preview = screen.getByTestId('hand-0-preview');
    expect(preview).toHaveTextContent('25');
    expect(preview.className).toContain('text-ds-error');
  });

  it('paints the neutral path when the post-swap score is unchanged', async () => {
    // Symmetric swap: hand0 [S5, C5]=10; hand1 [H5, D5]=10 → swap → hand0 [S5, D5]=10; hand1 [H5, C5]=10.
    const flatState: BlackJackSwitchResponse = {
      ...switchState,
      hands: [
        {
          ...switchState.hands[0],
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
          score: 10,
        },
        {
          ...switchState.hands[1],
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'DIAMOND', value: 5 },
          ],
          score: 10,
        },
      ],
    };
    mockApi.mockResolvedValue(flatState);
    renderWithProviders(<BlackJackSwitchPage />);
    const btn = await screen.findByTestId('switch-button');
    fireEvent.mouseEnter(btn);
    const preview = screen.getByTestId('hand-0-preview');
    expect(preview).toHaveTextContent('10');
    expect(preview.className).toContain('text-ds-text-muted');
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
    // Both per-hand results resolve via the new result.handWin key, and the overall key.
    expect(screen.getAllByText(/勝ち/).length).toBeGreaterThan(0);
    expect(screen.getByText('総合: 勝ち越し')).toBeInTheDocument();
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

  it('numbers the hands starting from 1', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('hand-0')).toHaveTextContent('ハンド 1');
    expect(screen.getByTestId('hand-1')).toHaveTextContent('ハンド 2');
  });

  it('shows a BUST badge on a busted hand', async () => {
    mockApi.mockResolvedValue({
      ...actionState,
      hands: [{ ...actionState.hands[0], busted: true, score: 25 }, actionState.hands[1]],
    });
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('hand-0-bust-badge')).toHaveTextContent('バースト');
    expect(screen.queryByTestId('hand-1-bust-badge')).not.toBeInTheDocument();
  });

  it('shows a BJ badge on a blackjack hand', async () => {
    mockApi.mockResolvedValue({
      ...actionState,
      hands: [{ ...actionState.hands[0], isBJ: true, score: 21 }, actionState.hands[1]],
    });
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('hand-0-bj-badge')).toHaveTextContent('BJ');
    expect(screen.queryByTestId('hand-1-bj-badge')).not.toBeInTheDocument();
  });

  it('marks the acting hand with a badge in ACTION phase', async () => {
    mockApi.mockResolvedValue({ ...actionState, currentHandIdx: 0 });
    renderWithProviders(<BlackJackSwitchPage />);
    expect(await screen.findByTestId('hand-0-acting-badge')).toHaveTextContent('操作中');
    expect(screen.queryByTestId('hand-1-acting-badge')).not.toBeInTheDocument();
  });

  it('shows no acting badge outside the ACTION phase', async () => {
    mockApi.mockResolvedValue(switchState);
    renderWithProviders(<BlackJackSwitchPage />);
    await screen.findByTestId('hand-0');
    expect(screen.queryByTestId('hand-0-acting-badge')).not.toBeInTheDocument();
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

  it('renders the i18n skeleton instead of a hardcoded Loading label before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<BlackJackSwitchPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
  });
});
