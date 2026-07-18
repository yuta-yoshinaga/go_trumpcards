import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ecarteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeEcarteState } from '../test/stateFactories';
import { EcartePage } from './EcartePage';

vi.mock('../api/gameApi', () => ({
  ecarteApi: { exec: vi.fn() },
  actionLogApi: { ecarte: vi.fn() },
}));

const mockExec = vi.mocked(ecarteApi.exec);

// Default fixture: a human Exchange turn at the ElderDecide sub-step (seat 0).
const elderDecideState = makeEcarteState({ phase: 0, negStep: 0, currentPlayerIdx: 0 });
// A human Exchange turn at the DealerRespond sub-step.
const dealerRespondState = makeEcarteState({ phase: 0, negStep: 1, currentPlayerIdx: 0 });
// A human Exchange turn at the ElderDiscard sub-step.
const discardState = makeEcarteState({ phase: 0, negStep: 2, currentPlayerIdx: 0 });
// A human Play turn.
const playPhaseState = makeEcarteState({ phase: 1, currentPlayerIdx: 0 });
// A CPU turn.
const cpuTurnState = makeEcarteState({ phase: 1, currentPlayerIdx: 1 });
const roundEndState = makeEcarteState({ phase: 2, dealPoints: [1, 0] });
const gameEndState = makeEcarteState({
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝利です (5-3)！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(elderDecideState);
});

describe('EcartePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<EcartePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<EcartePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetScore: 5 },
      }),
    );
  });

  it('shows propose/stand buttons on an ElderDecide turn', async () => {
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(screen.getByTestId('ecarte-propose')).toBeInTheDocument());
    expect(screen.getByTestId('ecarte-stand')).toBeInTheDocument();
  });

  it('dispatches propose when the propose button is clicked', async () => {
    renderWithProviders(<EcartePage />);
    const propose = await screen.findByTestId('ecarte-propose');
    mockExec.mockClear();
    mockExec.mockResolvedValue(elderDecideState);
    fireEvent.click(propose);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('propose'));
  });

  it('names the negotiation sub-step in the exchange banner', async () => {
    renderWithProviders(<EcartePage />);
    // elderDecideState is the beforeEach default → "Elder's proposal".
    await waitFor(() => expect(screen.getByTestId('ecarte-neg-step')).toHaveTextContent('エルダーの提案'));
  });

  it('shows the dealer-response sub-step label on a DealerRespond turn', async () => {
    mockExec.mockResolvedValue(dealerRespondState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(screen.getByTestId('ecarte-neg-step')).toHaveTextContent('ディーラーの応答'));
  });

  it('shows the discard sub-step label on a discard turn', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(screen.getByTestId('ecarte-neg-step')).toHaveTextContent('札交換'));
  });

  it('shows accept/refuse buttons on a DealerRespond turn', async () => {
    mockExec.mockResolvedValue(dealerRespondState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(screen.getByTestId('ecarte-accept')).toBeInTheDocument());
    expect(screen.getByTestId('ecarte-refuse')).toBeInTheDocument();
  });

  it('dispatches respond(false) when the refuse button is clicked', async () => {
    mockExec.mockResolvedValue(dealerRespondState);
    renderWithProviders(<EcartePage />);
    const refuse = await screen.findByTestId('ecarte-refuse');
    mockExec.mockClear();
    mockExec.mockResolvedValue(dealerRespondState);
    fireEvent.click(refuse);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', { accept: false }));
  });

  it('shows the discard button on an ElderDiscard turn and dispatches discard', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<EcartePage />);
    const card = await screen.findByAltText('♠ K');
    fireEvent.click(card);
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardState);
    fireEvent.click(screen.getByTestId('ecarte-discard'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [0] }));
  });

  it('disables the discard button and shows the empty reason when no card is selected', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<EcartePage />);
    const discard = await screen.findByTestId('ecarte-discard');
    expect(discard).toBeDisabled();
    expect(screen.getByTestId('ecarte-discard-reason')).toHaveTextContent('カードを選択してください');
    expect(screen.getByTestId('ecarte-discard-guide')).toHaveTextContent('0枚選択中');
  });

  it('enables the discard button and updates the count guide once a card is selected', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<EcartePage />);
    const card = await screen.findByAltText('♠ K');
    fireEvent.click(card);
    expect(screen.getByTestId('ecarte-discard')).toBeEnabled();
    expect(screen.getByTestId('ecarte-discard-guide')).toHaveTextContent('1枚選択中');
    expect(screen.queryByTestId('ecarte-discard-reason')).not.toBeInTheDocument();
  });

  it('disables the discard button and shows the stock reason when selecting more than the stock', async () => {
    mockExec.mockResolvedValue(makeEcarteState({ phase: 0, negStep: 2, currentPlayerIdx: 0, stockRemaining: 0 }));
    renderWithProviders(<EcartePage />);
    const card = await screen.findByAltText('♠ K');
    fireEvent.click(card);
    expect(screen.getByTestId('ecarte-discard')).toBeDisabled();
    expect(screen.getByTestId('ecarte-discard-reason')).toHaveTextContent('山札が足りません');
  });

  it('renders the play phase with the human cards and the play button', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ K')).toBeInTheDocument();
      expect(screen.getByAltText('♦ J')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
  });

  it('dispatches play when a card is selected and the play button clicked', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<EcartePage />);
    const card = await screen.findByAltText('♠ K');
    fireEvent.click(card);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果（獲得ポイント）')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<EcartePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です (5-3)！')).toBeInTheDocument());
  });
});
