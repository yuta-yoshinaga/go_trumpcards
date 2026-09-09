import { useMemo } from 'react';
import type { quinzeApi } from '../api/gameApi';
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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useQuinzeGame } from '../hooks/useQuinzeGame';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { QuinzeHand, QuinzeResponse } from '../types/card';
import { QuinzePhase } from '../types/games/quinze';
import type { TutorialStep } from '../types/tutorial';
import { parseQuinzeCommand, QUINZE_HELP } from '../utils/cli/commands/quinzeCommands';
import { formatQuinzeState } from '../utils/cli/formatters/quinzeFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';

const BET_OPTIONS = [10, 50, 100, 500];

const QUINZE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="quinze-seats"]', messageKey: 'tutorial.seats', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="quinze-banker"]', messageKey: 'tutorial.banker', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="quinze-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Quinze page and its banking-game controls. */
export const QuinzePage = withTutorial(QuinzePageContent, 'quinze', QUINZE_TUTORIAL_STEPS);

function QuinzePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('quinze');
  const game = useQuinzeGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('quinze');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('quinze', state);
  const cliConfig: CliGameConfig<QuinzeResponse, Parameters<typeof quinzeApi.exec>> = useMemo(
    () => ({
      gameName: 'quinze',
      parseCommand: parseQuinzeCommand,
      formatResponse: formatQuinzeState,
      helpText: QUINZE_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();

  const isPlayerTurn = state?.phase === QuinzePhase.PLAYER_TURN;
  const actionBindings = useMemo(
    () => [
      { key: 'h', action: game.handleHit, label: 'hit' },
      { key: 's', action: game.handleStand, label: 'stand' },
    ],
    [game],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayerTurn && !loading });

  if (!state) {
    return <GameSkeleton gameKey="quinze" layout={{ kind: 'tableau', topRow: 1, tableau: 3 }} />;
  }

  const ended = state.phase === QuinzePhase.END;
  const isBetting = state.phase === QuinzePhase.BET;
  const isBankerTurn = state.phase === QuinzePhase.BANKER_TURN;
  const bankerName = state.isHumanBanker ? t('bankerIsYou') : (state.seats[state.bankerIdx]?.name ?? '');

  /**
   * Render one hand. Visibility comes from the server's `hidden` flag; the page
   * never re-derives it, so there is one place that can be wrong instead of two.
   */
  const renderHand = (hand: QuinzeHand, label: string, keyPrefix: string) => (
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
        </div>
      )}
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.quinze')}
      gameThemeBg={gameTheme.quinze.bg}
      phaseName={
        ended
          ? t('phase.end')
          : isBetting
            ? t('phase.bet')
            : isBankerTurn
              ? t('phase.bankerTurn')
              : t('phase.playerTurn')
      }
      gamePath="/quinze"
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
            {t('target')}: {state.targetPoints}
          </span>
          {/* **相手がいつ引くのをやめるかは、賭け続けるかの判断材料。**
              ブラックジャックの「17 でスタンド」に当たる数字なのに、どの画面にも
              出ていなかった (#5566)。半点はサーバから来るので、5.5 という文字列を
              訳文にも画面にも焼き込まない。 */}
          <span className="text-sm text-ds-text-muted" data-testid="quinze-cpu-stand">
            {t('cpuStand', { total: state.cpuStandPoints })}
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
                    ? t('hiddenBankerHandAriaLabel', { count: state.bankerHand.cards.length })
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
                        // 伏せた手は「伏せられています」としか言わず、晴眼者が
                        // CardBack で数えている枚数が読み上げから落ちていた (#6362)。
                        seat.hand.hidden
                          ? t('hiddenHandAriaLabel', { name: seat.name, count: seat.hand.cards.length })
                          : seat.name,
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

          <GameFooter className={`${gameTheme.quinze.footer} px-4 py-2.5`}>
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

              <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
                <input
                  type="checkbox"
                  checked={frontendHintEnabled}
                  onChange={(e) => setFrontendHintEnabled(e.target.checked)}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="quinze-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
