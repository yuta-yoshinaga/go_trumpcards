import { useCallback, useMemo, useState } from 'react';
import { blackjackswitchApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { badgeErrorColors, badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BlackJackSwitchResponse, Card } from '../types/card';
import { BlackJackSwitchPhase, BlackJackSwitchResult } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { blackjackSwitchPreviewScores } from '../utils/blackjackSwitchPreview';
import { BLACKJACKSWITCH_HELP, parseBlackjackSwitchCommand } from '../utils/cli/commands/blackjackswitchCommands';
import { formatBlackjackSwitchState } from '../utils/cli/formatters/blackjackswitchFormatter';
import type { CliGameConfig } from '../utils/cli/types';

const BJSWITCH_TUTORIAL_STEPS: TutorialStep[] = [];

// Mirrors the backend BJSwitchMaxBet (per-hand cap in internal/domain/BlackJackSwitch.go).
const BJSWITCH_MAX_BET = 10000;

/** Renders the Blackjack Switch game page (#1669). */
export const BlackJackSwitchPage = withTutorial(BlackJackSwitchPageContent, 'blackjackswitch', BJSWITCH_TUTORIAL_STEPS);

function BlackJackSwitchPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('blackjackswitch');

  const [betAmount, setBetAmount] = useState(100);
  const [switchPreview, setSwitchPreview] = useState(false);
  const [alwaysPreview, setAlwaysPreview] = useState(false);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(blackjackswitchApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('blackjackswitch');
  const cliConfig: CliGameConfig<BlackJackSwitchResponse, Parameters<typeof blackjackswitchApi.exec>> = useMemo(
    () => ({
      gameName: 'blackjackswitch',
      parseCommand: parseBlackjackSwitchCommand,
      formatResponse: formatBlackjackSwitchState,
      helpText: BLACKJACKSWITCH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === BlackJackSwitchPhase.BET;
  const isSwitchPhase = state?.phase === BlackJackSwitchPhase.SWITCH;
  const isActionPhase = state?.phase === BlackJackSwitchPhase.ACTION;
  const isEndPhase = state?.phase === BlackJackSwitchPhase.END;

  // Defined above the loading early-return so the keyboard hook keeps a stable
  // call order. Double-down is only offered on a fresh two-card hand.
  const canDoubleDown = isActionPhase && state?.hands[state.currentHandIdx]?.cards.length === 2;
  const handleBet = useCallback(() => execApi('bet', betAmount), [execApi, betAmount]);
  const handleSwitch = useCallback(() => execApi('switch'), [execApi]);
  const handleKeep = useCallback(() => execApi('keep'), [execApi]);
  const handleHit = useCallback(() => execApi('hit'), [execApi]);
  const handleStand = useCallback(() => execApi('stand'), [execApi]);
  const handleDoubleDown = useCallback(() => execApi('doubledown'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleBet, enabled: isBetPhase },
      { key: 's', action: handleSwitch, enabled: isSwitchPhase },
      { key: 'k', action: handleKeep, enabled: isSwitchPhase },
      { key: 'h', action: handleHit, enabled: isActionPhase },
      { key: 't', action: handleStand, enabled: isActionPhase },
      { key: 'd', action: handleDoubleDown, enabled: canDoubleDown },
      { key: 'r', action: handleReset, enabled: isEndPhase },
    ],
    [
      handleBet,
      handleSwitch,
      handleKeep,
      handleHit,
      handleStand,
      handleDoubleDown,
      handleReset,
      isBetPhase,
      isSwitchPhase,
      isActionPhase,
      isEndPhase,
      canDoubleDown,
    ],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.blackjackswitch.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const phaseName = isBetPhase
    ? t('phase.bet')
    : isSwitchPhase
      ? t('phase.switch')
      : isActionPhase
        ? t('phase.action')
        : t('phase.end');

  const previewScores =
    isSwitchPhase && (switchPreview || alwaysPreview) && state.hands.length >= 2
      ? blackjackSwitchPreviewScores(state.hands[0].cards, state.hands[1].cards)
      : null;

  return (
    <GamePageShell
      title={tc('nav.blackjackswitch')}
      gameThemeBg={gameTheme.blackjackswitch.bg}
      phaseName={phaseName}
      gamePath="/blackjackswitch"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.totalPayout > state.hands.reduce((sum, h) => sum + h.bet, 0)}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {t('label.chips')}: {state.chips}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div
            data-testid="card-area"
            className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {isBetPhase && <p className="text-ds-text-muted text-center text-sm py-3">{t('guide.bet')}</p>}
            {isSwitchPhase && <p className="text-ds-text-muted text-center text-sm py-2">{t('guide.switch')}</p>}

            {state.dealerCards.length > 0 && (
              <div className="mb-3" data-testid="dealer-area">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">
                  {t('label.dealer')} ({state.dealerScore})
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.dealerCards.map((c, i) => (
                    <CardSlot key={`dealer-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {state.hands.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-3" data-testid="hands-area">
                {state.hands.map((hand, idx) => {
                  const isCurrent = isActionPhase && idx === state.currentHandIdx;
                  const tone = isCurrent ? 'border-ds-accent' : 'border-white/20';
                  const previewScore = previewScores ? (idx === 0 ? previewScores.a : previewScores.b) : null;
                  const previewDelta = previewScore !== null ? previewScore - hand.score : 0;
                  const previewColor =
                    previewScore === null
                      ? ''
                      : previewScore > 21
                        ? 'text-ds-error'
                        : previewDelta > 0
                          ? 'text-ds-success'
                          : previewDelta < 0
                            ? 'text-ds-error'
                            : 'text-ds-text-muted';
                  return (
                    <div
                      key={`hand-${idx}`}
                      data-testid={`hand-${idx}`}
                      className={`border rounded-lg p-2 ${tone} bg-black/20`}
                    >
                      <div className="text-ds-text-primary text-center text-xs font-bold mb-1">
                        {t('label.hand')} {idx + 1} ({hand.score}
                        {previewScore !== null && (
                          <span data-testid={`hand-${idx}-preview`} className={`ml-1 font-bold ${previewColor}`}>
                            → {previewScore}
                          </span>
                        )}
                        ) — {t('label.bet')}: {hand.bet}
                        {hand.busted && (
                          <span
                            data-testid={`hand-${idx}-bust-badge`}
                            className={`ml-1 inline-block rounded px-1.5 py-0.5 text-[10px] ${badgeErrorColors}`}
                          >
                            {t('badge.bust')}
                          </span>
                        )}
                        {hand.isBJ && (
                          <span
                            data-testid={`hand-${idx}-bj-badge`}
                            className={`ml-1 inline-block rounded px-1.5 py-0.5 text-[10px] ${badgeSuccessColors}`}
                          >
                            {t('badge.bj')}
                          </span>
                        )}
                        {isCurrent && (
                          <span
                            data-testid={`hand-${idx}-acting-badge`}
                            className={`ml-1 inline-block rounded px-1.5 py-0.5 text-[10px] ${badgeWarningColors}`}
                          >
                            {t('badge.acting')}
                          </span>
                        )}
                      </div>
                      <div className="flex justify-center gap-1 flex-wrap">
                        {hand.cards.map((c, j) => (
                          <CardSlot key={`hand-${idx}-${j}`} card={c} width={cardWidth} />
                        ))}
                      </div>
                      {isEndPhase && (
                        <div className="text-ds-text-primary text-center text-xs mt-1">
                          {t(handResultKey(hand.result))} — {t('payout.total')}: {hand.payout}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}

            {state.switched && <div className="text-ds-text-muted text-center text-xs mb-2">{t('label.switched')}</div>}

            {isEndPhase && state.dealerPushed22 && (
              <div className="text-ds-warning text-center text-sm mb-2" data-testid="dealer22-banner">
                {t('result.dealer22Push')}
              </div>
            )}

            {isEndPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
                <div className="font-bold">
                  {t('payout.total')}: {state.totalPayout}
                </div>
                <div className="text-ds-text-muted">{t(overallResultKey(state.overallResult))}</div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.blackjackswitch.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox' as const,
                      id: 'blackjackswitch-always-preview',
                      label: t('label.alwaysPreview'),
                      checked: alwaysPreview,
                      onToggle: setAlwaysPreview,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <ChipBetInput
                  id="blackjackswitch-bet-amount"
                  label={t('label.bet')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={Math.min(Math.floor(state.chips / 2), BJSWITCH_MAX_BET)}
                />
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleBet}
                  disabled={loading}
                  aria-keyshortcuts="b"
                >
                  {t('button.bet')}
                  <KbdBadge label={t('kbd.bet')} />
                </button>
              </div>
            )}
            {isSwitchPhase && (
              <div className="flex justify-center gap-2 pb-2 flex-wrap">
                <button
                  type="button"
                  className={btnWarning}
                  onClick={handleSwitch}
                  onMouseEnter={() => setSwitchPreview(true)}
                  onMouseLeave={() => setSwitchPreview(false)}
                  onFocus={() => setSwitchPreview(true)}
                  onBlur={() => setSwitchPreview(false)}
                  onTouchStart={() => setSwitchPreview(true)}
                  onTouchEnd={() => setSwitchPreview(false)}
                  onTouchCancel={() => setSwitchPreview(false)}
                  data-testid="switch-button"
                  disabled={loading}
                  aria-keyshortcuts="s"
                >
                  {t('button.switch')}
                  <KbdBadge label={t('kbd.switch')} />
                </button>
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={handleKeep}
                  disabled={loading}
                  aria-keyshortcuts="k"
                >
                  {t('button.keep')}
                  <KbdBadge label={t('kbd.keep')} />
                </button>
              </div>
            )}
            {isActionPhase && (
              <div className="flex justify-center gap-2 pb-2 flex-wrap">
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleHit}
                  disabled={loading}
                  aria-keyshortcuts="h"
                >
                  {t('button.hit')}
                  <KbdBadge label={t('kbd.hit')} />
                </button>
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleStand}
                  disabled={loading}
                  aria-keyshortcuts="t"
                >
                  {t('button.stand')}
                  <KbdBadge label={t('kbd.stand')} />
                </button>
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={handleDoubleDown}
                  disabled={loading || !canDoubleDown}
                  aria-keyshortcuts="d"
                >
                  {t('button.doubleDown')}
                  <KbdBadge label={t('kbd.doubleDown')} />
                </button>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2 flex-wrap">
                <GameResetButton
                  isGameEnd={isEndPhase}
                  onReset={handleReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('actionLog.view')}
                </button>
              </div>
            )}
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Renders a face-up card or a face-down placeholder when card is null. */
function CardSlot({ card, width }: { card: Card | null; width: number }) {
  if (!card) {
    return (
      <div
        data-testid="card-back"
        className="rounded-md border border-white/30 bg-game-card-back"
        style={{ width, height: width * 1.4 }}
      />
    );
  }
  return <AnimatedCard card={card} width={width} />;
}

function handResultKey(result: number): string {
  switch (result) {
    case BlackJackSwitchResult.WIN:
      return 'result.handWin';
    case BlackJackSwitchResult.LOSE:
      return 'result.handLose';
    default:
      return 'result.handDraw';
  }
}

function overallResultKey(result: number): string {
  switch (result) {
    case BlackJackSwitchResult.WIN:
      return 'result.overallWin';
    case BlackJackSwitchResult.LOSE:
      return 'result.overallLose';
    default:
      return 'result.overallDraw';
  }
}
