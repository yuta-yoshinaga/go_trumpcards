import { useEffect, useMemo } from 'react';
import type { anacondaApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useAnacondaGame,
} from '../hooks/useAnacondaGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardSelection } from '../hooks/useCardSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, focusRingAccent } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AnacondaResponse } from '../types/card';
import { AnacondaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { ANACONDA_HELP, parseAnacondaCommand } from '../utils/cli/commands/anacondaCommands';
import { formatAnacondaState } from '../utils/cli/formatters/anacondaFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Anaconda tutorial step definitions. */
const ANACONDA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="anaconda-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="anaconda-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="anaconda-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="anaconda-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const ANACONDA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [AnacondaPhase.PASS]: 'pass',
  [AnacondaPhase.SET]: 'set',
  [AnacondaPhase.ROLL]: 'roll',
  [AnacondaPhase.RESULT]: 'result',
};

/** Number of cards a player keeps at the Set phase. */
const KEEP_SIZE = 5;

/** Renders the Anaconda (Pass the Trash) game page: a multi-phase poker pot game. */
export const AnacondaPage = withTutorial(AnacondaPageContent, 'anaconda', ANACONDA_TUTORIAL_STEPS);

/** Inner content of the Anaconda page, wrapped by TutorialProvider. */
function AnacondaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('anaconda');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    anacondaConfig,
    handleConfigChange,
    reset,
    handlePass,
    handleKeep,
    handleCall,
    handleRaise,
    handleFold,
    handleNextRound,
  } = useAnacondaGame();

  const { selected, toggle, clear } = useCardSelection();

  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('anaconda');
  const cliConfig: CliGameConfig<AnacondaResponse, Parameters<typeof anacondaApi.exec>> = useMemo(
    () => ({
      gameName: 'anaconda',
      parseCommand: parseAnacondaCommand,
      formatResponse: formatAnacondaState,
      helpText: ANACONDA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('anaconda', state);
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('anaconda', ANACONDA_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="anaconda" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isPassPhase = state.phase === AnacondaPhase.PASS;
  const isSetPhase = state.phase === AnacondaPhase.SET;
  const isRollPhase = state.phase === AnacondaPhase.ROLL;
  const isResultPhase = state.phase === AnacondaPhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const humanTurn = state.isHumanTurn && !isGameEnd;
  const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);

  const canSelectCards = (isPassPhase || isSetPhase) && humanTurn;
  const passReady = isPassPhase && humanTurn && selected.length === state.passCount;
  const keepReady = isSetPhase && humanTurn && selected.length === KEEP_SIZE;

  // Immediate over/under feedback on the selected-card count for the Pass/Set phases.
  const requiredSelection = isPassPhase ? state.passCount : KEEP_SIZE;
  const selectionDiff = selected.length - requiredSelection;
  const selectionFeedback =
    selectionDiff < 0
      ? {
          text: t('selectionRemaining', { n: -selectionDiff, selected: selected.length, required: requiredSelection }),
          cls: 'text-ds-warning',
        }
      : selectionDiff > 0
        ? {
            text: t('selectionOver', { n: selectionDiff, selected: selected.length, required: requiredSelection }),
            cls: 'text-ds-danger',
          }
        : {
            text: t('selectionReady', { selected: selected.length, required: requiredSelection }),
            cls: 'text-ds-success',
          };

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const playerBadge = (p: AnacondaResponse['players'][number]): string =>
    p.out
      ? t('badge.out')
      : p.folded
        ? t('badge.folded')
        : p.isWinner
          ? t('badge.winner')
          : p.id === state.currentPlayer && !isResultPhase
            ? t('badge.turn')
            : t('badge.waiting');

  const handleManualReset = () => {
    hideActionLog();
    clear();
    reset();
  };

  const onPass = () => {
    handlePass(selected);
    clear();
  };

  const onKeep = () => {
    handleKeep(selected);
    clear();
  };

  return (
    <GamePageShell
      title={tc('nav.anaconda')}
      gameThemeBg={gameTheme.anaconda.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={humanTurn && !isResultPhase}
      gamePath="/anaconda"
      gameEndFlag={isGameEnd}
      winShow={isResultPhase && (humanWonMatch || state.result > 0)}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>{t('chips', { amount: state.chips })}</span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: anacondaConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: anacondaConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: anacondaConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: anacondaConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="anaconda-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span className="mr-4">{t('ante', { amount: state.ante })}</span>
              {isRollPhase && <span>{t('currentBet', { amount: state.currentBet })}</span>}
            </div>

            {isPassPhase && humanTurn && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">
                {t('passNotice', { count: state.passCount })}
              </div>
            )}
            {isSetPhase && humanTurn && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">{t('setNotice')}</div>
            )}
            {isRollPhase && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">
                {t('rollNotice', { revealed: state.rollIndex })}
              </div>
            )}
            {canSelectCards && (
              <div
                data-testid="anaconda-selection-feedback"
                aria-live="polite"
                className={`text-center mb-2 text-sm font-semibold ${selectionFeedback.cls}`}
              >
                {selectionFeedback.text}
              </div>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="anaconda-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                  {t('roundBet', { amount: p.roundBet })} · [{playerBadge(p)}]
                  {p.handName ? ` · ${handName(p.handName)}` : ''}
                </div>
              ))}
            </div>

            {/* Revealed CPU hands during Roll / at Result */}
            {(isRollPhase || isResultPhase) && (
              <div className="mb-2 p-2 rounded bg-black/30">
                {state.players
                  .filter((p) => !p.isHuman && p.cards.length > 0)
                  .map((p) => (
                    <div key={p.id} className="mb-1">
                      <div className="text-ds-text-muted text-xs mb-0.5">
                        {playerLabel(p.id, p.isHuman)}
                        {p.handName ? ` — ${handName(p.handName)}` : ''}
                      </div>
                      <div className="flex gap-1">
                        {p.cards.map((c, i) => (
                          <CardImage key={`${p.id}-${i}`} card={c} width={cardWidth} />
                        ))}
                      </div>
                    </div>
                  ))}
              </div>
            )}

            {/* Round result */}
            {isResultPhase && state.winnerIdx >= 0 && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div>
                  {t('roundResult.winner', {
                    name: playerLabel(state.winnerIdx, state.winnerIdx === humanIdx),
                    pot: state.pot,
                  })}
                </div>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.anaconda.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 ? (
              <div className="mb-2" data-tutorial="anaconda-hand">
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {t('handLabel')}
                  {humanPlayer.handName ? ` — ${handName(humanPlayer.handName)}` : ''}
                  {canSelectCards ? ` · ${t('selectedCount', { n: selected.length })}` : ''}
                </div>
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards.map((c, i) => {
                    const isSelected = selected.includes(i);
                    return (
                      <button
                        key={`human-${c.design}-${c.value}`}
                        type="button"
                        aria-pressed={isSelected}
                        disabled={!canSelectCards || loading}
                        onClick={() => toggle(i)}
                        className={`${focusRingAccent} rounded`}
                        style={{
                          background: 'none',
                          padding: 0,
                          cursor: canSelectCards ? 'pointer' : 'default',
                          borderRadius: 8,
                          ...selectedCardStyle(isSelected),
                          boxSizing: 'border-box',
                        }}
                      >
                        <CardImage card={c} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="anaconda-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="anaconda-action-buttons">
              {isPassPhase && humanTurn && (
                <button type="button" className={btnPrimary} onClick={onPass} disabled={loading || !passReady}>
                  {t('passButton')}
                </button>
              )}

              {isSetPhase && humanTurn && (
                <button type="button" className={btnPrimary} onClick={onKeep} disabled={loading || !keepReady}>
                  {t('keepButton')}
                </button>
              )}

              {isRollPhase && humanTurn && (
                <>
                  <button type="button" className={btnSecondary} onClick={handleCall} disabled={loading}>
                    {t('callButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleRaise}
                    disabled={loading || !state.canRaise}
                  >
                    {t('raiseButton')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                    {t('foldButton')}
                  </button>
                </>
              )}

              {isResultPhase && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="anaconda-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
