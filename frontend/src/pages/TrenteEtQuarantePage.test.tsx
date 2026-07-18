import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trenteetquaranteApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTrenteEtQuaranteState } from '../test/stateFactories';
import type { Card, CardDesign } from '../types/card';
import { TrenteEtQuaranteBetType, TrenteEtQuarantePhase } from '../types/phases';
import { TrenteEtQuarantePage } from './TrenteEtQuarantePage';

vi.mock('../api/gameApi', () => ({
  trenteetquaranteApi: { exec: vi.fn() },
  actionLogApi: { trenteetquarante: vi.fn() },
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

const mockApi = vi.mocked(trenteetquaranteApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betState = makeTrenteEtQuaranteState();

const endState = makeTrenteEtQuaranteState({
  phase: TrenteEtQuarantePhase.RESULT,
  stake: 100,
  currentBet: TrenteEtQuaranteBetType.NOIR,
  noirRow: [card('SPADE', 10), card('CLOVER', 13), card('SPADE', 8)],
  rougeRow: [card('HEART', 9), card('DIAMOND', 13), card('HEART', 10)],
  noirTotal: 31,
  rougeTotal: 39,
  winningRow: 0, // Noir row wins
  firstCardRed: false,
  result: 1,
  payout: 200,
  gameEndFlag: true,
  message: '',
});

const refaitState = makeTrenteEtQuaranteState({
  phase: TrenteEtQuarantePhase.RESULT,
  stake: 100,
  noirRow: [card('SPADE', 1)],
  rougeRow: [card('HEART', 1)],
  noirTotal: 31,
  rougeTotal: 31,
  winningRow: -1,
  refait: true,
  result: -1,
  payout: 50,
  gameEndFlag: true,
});

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockResolvedValue(betState);
});

describe('TrenteEtQuarantePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<TrenteEtQuarantePage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders the bet phase with the four bet buttons and a deal button', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<TrenteEtQuarantePage />);
    await waitFor(() => expect(screen.getByTestId('teq-deal-button')).toBeInTheDocument());
    expect(screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.NOIR}`)).toBeInTheDocument();
    expect(screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.ROUGE}`)).toBeInTheDocument();
    expect(screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.COULEUR}`)).toBeInTheDocument();
    expect(screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.INVERSE}`)).toBeInTheDocument();
  });

  it('highlights the selected bet with a check mark and always shows the descriptions', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<TrenteEtQuarantePage />);
    // Couleur/Inverse meanings are shown as static text (not only tooltips).
    await waitFor(() => expect(screen.getByText('最初のカードの色が勝ち色と一致')).toBeInTheDocument());
    expect(screen.getByText('最初のカードの色が勝ち色と不一致')).toBeInTheDocument();
    // Selecting Rouge marks it pressed with a ✓; the previously selected Noir is not.
    fireEvent.click(screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.ROUGE}`));
    const rouge = screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.ROUGE}`);
    expect(rouge).toHaveAttribute('aria-pressed', 'true');
    expect(rouge).toHaveTextContent('✓');
    expect(rouge.className).toContain('ring-2');
    expect(screen.getByTestId(`teq-bet-${TrenteEtQuaranteBetType.NOIR}`)).toHaveAttribute('aria-pressed', 'false');
  });

  it('deals with the selected bet type and stake', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<TrenteEtQuarantePage />);
    const rougeBtn = await screen.findByTestId(`teq-bet-${TrenteEtQuaranteBetType.ROUGE}`);
    fireEvent.click(rougeBtn);
    mockApi.mockClear();
    fireEvent.click(screen.getByTestId('teq-deal-button'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', TrenteEtQuaranteBetType.ROUGE, 100));
  });

  it('renders both rows and a win result at the result phase', async () => {
    mockApi.mockResolvedValue(endState);
    renderWithProviders(<TrenteEtQuarantePage />);
    await waitFor(() => expect(screen.getByTestId('teq-result')).toBeInTheDocument());
    expect(screen.getByTestId('teq-next-round-button')).toBeInTheDocument();
    expect(screen.getByTestId('teq-result')).toHaveTextContent('200');
  });

  it('shows the refait message on a tie at 31', async () => {
    mockApi.mockResolvedValue(refaitState);
    renderWithProviders(<TrenteEtQuarantePage />);
    // Scope to the result banner so the rule-explainer summary (which also mentions Refait) is excluded.
    await waitFor(() => expect(screen.getByTestId('teq-result')).toHaveTextContent(/ルフェ/));
  });

  it('renders the static refait rule explainer with the half-stake detail', async () => {
    mockApi.mockResolvedValue(refaitState);
    renderWithProviders(<TrenteEtQuarantePage />);
    const explainer = await screen.findByTestId('teq-refait-explainer');
    expect(explainer).toBeInTheDocument();
    // The explainer must accurately state that the house takes half the stake.
    expect(explainer).toHaveTextContent('賭け金の半分を失う');
    expect(explainer).toHaveTextContent(/ハウスエッジ/);
  });

  it('starts the next round from the result phase', async () => {
    mockApi.mockResolvedValue(endState);
    renderWithProviders(<TrenteEtQuarantePage />);
    const nextBtn = await screen.findByTestId('teq-next-round-button');
    mockApi.mockClear();
    fireEvent.click(nextBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('nextround'));
  });

  it('reads from useCliMode', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<TrenteEtQuarantePage />);
    await waitFor(() => expect(mockUseCliMode).toHaveBeenCalledWith('trenteetquarante'));
  });

  it('renders the CLI terminal when useCliMode reports cliEnabled', async () => {
    mockUseCliMode.mockReturnValueOnce({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<TrenteEtQuarantePage />);
    await waitFor(() => expect(screen.queryByTestId('teq-deal-button')).not.toBeInTheDocument());
  });
});
