import { useMemo } from 'react';
import type { settemezzoApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSetteEMezzoGame } from '../hooks/useSetteEMezzoGame';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SetteEMezzoHand, SetteEMezzoResponse } from '../types/card';
import { SetteEMezzoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSetteEMezzoCommand, SETTEMEZZO_HELP } from '../utils/cli/commands/settemezzoCommands';
import { formatSetteEMezzoState } from '../utils/cli/formatters/settemezzoFormatter';
import type { CliGameConfig } from '../utils/cli/types';

const BET_OPTIONS = [10, 50, 100, 500];

/**
 * The matta's choices, in HALVES. 1 is the half-point; 2..14 are the whole
 * numbers 1 through 7. Halves all the way through means no rounding anywhere.
 */
const MATTA_CHOICES = [1, 2, 4, 6, 8, 10, 12, 14];

const SM_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sm-seats"]', messageKey: 'tutorial.halfPoints', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sm-controls"]', messageKey: 'tutorial.matta', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sm-banker"]', messageKey: 'tutorial.target', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sm-banker"]', messageKey: 'tutorial.banker', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sm-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Sette e Mezzo page: half-point cards, a wild matta, a rotating bank. */
export const SetteEMezzoPage = withTutorial(SetteEMezzoPageContent, 'settemezzo', SM_TUTORIAL_STEPS);

/** Render a halves value as points, e.g. 15 -> "7.5". */
function halvesToLabel(halves: number): string {
  return halves % 2 === 0 ? String(halves / 2) : `${Math.floor(halves / 2)}.5`;
}

function SetteEMezzoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('settemezzo');
  const game = useSetteEMezzoGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('settemezzo');
  const cliConfig: CliGameConfig<SetteEMezzoResponse, Parameters<typeof settemezzoApi.exec>> = useMemo(
    () => ({
      gameName: 'settemezzo',
      parseCommand: parseSetteEMezzoCommand,
      formatResponse: formatSetteEMezzoState,
      helpText: SETTEMEZZO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();

  const isPlayerTurn = state?.phase === SetteEMezzoPhase.PLAYER_TURN;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: game.handleHit, label: 'hit' },
      { key: 's', action: game.handleStand, label: 'stand' },
    ],
    [game],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayerTurn && !loading });

  if (!state) {
    return <GameSkeleton gameKey="settemezzo" layout={{ kind: 'tableau', topRow: 1, tableau: 3 }} />;
  }

  const ended = state.phase === SetteEMezzoPhase.END;
  const isBetting = state.phase === SetteEMezzoPhase.BET;
  const isBankerTurn = state.phase === SetteEMezzoPhase.BANKER_TURN;
  const bankerName = state.isHumanBanker ? t('bankerIsYou') : (state.seats[state.bankerIdx]?.name ?? '');

  /**
   * Render one hand. Visibility comes from the server's `hidden` flag; the page
   * never re-derives it, so there is one place that can be wrong instead of two.
   */
  const renderHand = (hand: SetteEMezzoHand, label: string, keyPrefix: string) => (
    <div className="text-center">
      <div
        className="flex gap-1 justify-center"
        role="img"
        aria-label={hand.hidden ? label : t('seatAriaLabel', { name: label, total: hand.totalLabel })}
      >
        {hand.cards.map((card, i) =>
          hand.hidden || !card ? (
            <CardBack key={`${keyPrefix}-c${i.toString()}`} width={cardWidth} />
          ) : (
            <AnimatedCard key={`${keyPrefix}-c${i.toString()}`} card={card} width={cardWidth} draggable={false} />
          ),
        )}
      </div>
      {!hand.hidden && (
        <div className="text-game-text-muted text-xs mt-1">
          {t('total')}: {hand.totalLabel}
          {/* The matta stays adjustable until the hand stands, so its current
              value has to be visible or the choice is blind. */}
          {hand.hasMatta && (
            <span className="ml-1 text-ds-warning">
              {t('mattaValue', { value: halvesToLabel(hand.mattaHalves || 1) })}
            </span>
          )}
        </div>
      )}
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.settemezzo')}
      gameThemeBg={gameTheme.settemezzo.bg}
      phaseName={
        ended
          ? t('phase.end')
          : isBetting
            ? t('phase.bet')
            : isBankerTurn
              ? t('phase.bankerTurn')
              : t('phase.playerTurn')
      }
      gamePath="/settemezzo"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('chips')}: {state.chips}
          </span>
          <span className="text-sm text-ds-text-muted">
            {t('banker')}: {bankerName}
          </span>
          <span className="text-sm text-ds-text-muted">
            {t('target')}: {halvesToLabel(state.targetHalves)}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="text-center mb-4" data-tutorial="sm-banker">
              <div className="text-game-text-muted text-xs mb-1">{t('bankerHand')}</div>
              {state.bankerHand ? (
                renderHand(
                  state.bankerHand,
                  state.bankerHand.hidden
                    ? t('hiddenBankerHandAriaLabel')
                    : t('bankerHandAriaLabel', { total: state.bankerHand.totalLabel }),
                  'banker',
                )
              ) : (
                <div className="text-game-text-muted text-sm">—</div>
              )}
            </div>

            <div className="flex flex-wrap justify-center gap-4 sm:gap-8" data-tutorial="sm-seats">
              {state.seats.map((seat, seatIdx) => {
                if (seatIdx === state.bankerIdx || !seat.hand) return null;
                const onTurn = isPlayerTurn && seatIdx === state.activeSeat;
                return (
                  <div key={`seat-${seatIdx.toString()}`} className="text-center">
                    <div className="text-game-text-muted text-xs mb-1">{seat.name}</div>
                    <div className={onTurn ? 'ring-2 ring-ds-warning rounded p-1' : 'p-1'}>
                      {renderHand(
                        seat.hand,
                        seat.hand.hidden ? t('hiddenHandAriaLabel', { name: seat.name }) : seat.name,
                        `s${seatIdx.toString()}`,
                      )}
                      <div className="text-game-text-muted text-xs mt-1">
                        {t('bet')}: {seat.hand.bet}
                        {ended && seat.hand.payout !== 0 && (
                          <span className={seat.hand.payout > 0 ? ' text-ds-success' : ' text-ds-danger'}>
                            {' '}
                            {seat.hand.payout > 0 ? `+${seat.hand.payout}` : seat.hand.payout}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={ended}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.settemezzo.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="sm-controls">
              {isBetting && !state.isHumanBanker && (
                <>
                  <span className="text-sm text-ds-text-muted">{t('betLabel')}</span>
                  {BET_OPTIONS.map((amount) => (
                    <button
                      key={`bet-${amount.toString()}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => game.handleBet(amount)}
                      disabled={loading || amount > state.chips}
                    >
                      {t('betAmount', { amount })}
                    </button>
                  ))}
                </>
              )}

              {isBetting && state.isHumanBanker && (
                <button type="button" className={btnPrimary} onClick={game.handleDeal} disabled={loading}>
                  {t('actions.deal')}
                </button>
              )}

              {isPlayerTurn && state.canHit && (
                <button type="button" className={btnPrimary} onClick={game.handleHit} disabled={loading}>
                  {t('actions.hit')}
                </button>
              )}
              {isPlayerTurn && state.canStand && (
                <button type="button" className={btnSuccess} onClick={game.handleStand} disabled={loading}>
                  {t('actions.stand')}
                </button>
              )}

              {/* The matta's value is a choice, not a consequence, so it gets
                  its own row of buttons rather than being inferred. */}
              {isPlayerTurn && state.canSetMatta && (
                <>
                  <span className="text-sm text-ds-text-muted">{t('mattaLabel')}</span>
                  {MATTA_CHOICES.map((halves) => (
                    <button
                      key={`matta-${halves.toString()}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => game.handleMatta(halves)}
                      disabled={loading}
                      data-testid={`matta-${halves.toString()}`}
                    >
                      {halvesToLabel(halves)}
                    </button>
                  ))}
                </>
              )}

              {isBankerTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={game.handleBankerHit} disabled={loading}>
                    {t('actions.bankerHit')}
                  </button>
                  <button type="button" className={btnDanger} onClick={game.handleBankerStand} disabled={loading}>
                    {t('actions.bankerStand')}
                  </button>
                </>
              )}

              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sm-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="settemezzo-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
