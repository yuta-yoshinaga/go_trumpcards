import { useEffect, useMemo } from 'react';
import type { primeroApi } from '../api/gameApi';
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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  usePrimeroGame,
} from '../hooks/usePrimeroGame';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PrimeroResponse } from '../types/card';
import { PrimeroPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PRIMERO_HELP, parsePrimeroCommand } from '../utils/cli/commands/primeroCommands';
import { formatPrimeroState } from '../utils/cli/formatters/primeroFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Primero tutorial step definitions. */
const PRIMERO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="primero-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="primero-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="primero-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="primero-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PRIMERO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PrimeroPhase.BETTING]: 'betting',
  [PrimeroPhase.RESULT]: 'result',
};

/** Primero hand categories ordered strongest-first (mirrors the Go domain ranking). */
const PRIMERO_HAND_RANKING = ['fluxus', 'supremus', 'primero', 'numerus'] as const;

/** Renders the Primero game page: a Renaissance 4-card pot-vying game. */
export const PrimeroPage = withTutorial(PrimeroPageContent, 'primero', PRIMERO_TUTORIAL_STEPS);

/** Inner content of the Primero page, wrapped by TutorialProvider. */
function PrimeroPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('primero');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    primeroConfig,
    handleConfigChange,
    reset,
    handleCall,
    handleRaise,
    handleFold,
    handleNextRound,
  } = usePrimeroGame();

  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('primero');
  const cliConfig: CliGameConfig<PrimeroResponse, Parameters<typeof primeroApi.exec>> = useMemo(
    () => ({
      gameName: 'primero',
      parseCommand: parsePrimeroCommand,
      formatResponse: formatPrimeroState,
      helpText: PRIMERO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('primero', state);
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('primero', PRIMERO_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="primero" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isBettingPhase = state.phase === PrimeroPhase.BETTING;
  const isResultPhase = state.phase === PrimeroPhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const playerBadge = (p: PrimeroResponse['players'][number]): string =>
    p.out ? t('badge.out') : p.isWinner ? t('badge.winner') : p.folded ? t('badge.folded') : t('badge.active');

  /** A colour-independent glyph for each player state (WCAG 1.4.1). */
  const playerBadgeIcon = (p: PrimeroResponse['players'][number]): string =>
    p.out ? '—' : p.isWinner ? '👑' : p.folded ? '×' : '●';

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.primero')}
      gameThemeBg={gameTheme.primero.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isBettingPhase && isHumanTurn && !isGameEnd}
      gamePath="/primero"
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
                    value: primeroConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: primeroConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: primeroConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: primeroConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="primero-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span className="mr-4">{t('ante', { amount: state.ante })}</span>
              <span>{t('currentBet', { amount: state.currentBet })}</span>
            </div>

            {isBettingPhase && isHumanTurn && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">{t('bettingNotice')}</div>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="primero-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              <ul className="list-none">
                {state.players.map((p) => {
                  const isCurrentTurn = !isGameEnd && state.currentPlayerIdx === p.id;
                  return (
                    <li
                      key={p.id}
                      data-testid={`primero-player-${p.id}`}
                      aria-label={t('playerRowLabel', { name: playerLabel(p.id, p.isHuman), status: playerBadge(p) })}
                      className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''} ${
                        isCurrentTurn ? 'border-l-2 border-ds-accent pl-1' : ''
                      }`}
                    >
                      {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                      {t('roundBet', { amount: p.roundBet })} · <span aria-hidden="true">{playerBadgeIcon(p)} </span>[
                      {playerBadge(p)}]{p.handName ? ` · ${handName(p.handName)}` : ''}
                    </li>
                  );
                })}
              </ul>
            </div>

            {/* Revealed hands at result */}
            {isResultPhase && (
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
          <GameFooter className={`${gameTheme.primero.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 ? (
              <div className="mb-2" data-tutorial="primero-hand">
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {t('handLabel')}
                  {humanPlayer.handName ? ` — ${handName(humanPlayer.handName)}` : ''}
                </div>
                <div className="flex gap-1">
                  {humanPlayer.cards.map((c, i) => (
                    <CardImage key={`human-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="primero-hand">
                {t('handLabel')}
              </div>
            )}

            {/* Hand-ranking legend (static reference), highlighting the current hand */}
            <details className="mb-2 p-2 rounded bg-black/30" data-testid="primero-hand-legend">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {t('handLegend.title')}
              </summary>
              <div className="mt-1 text-ds-text-muted text-xs">
                <table className="w-full">
                  <thead>
                    <tr>
                      <th scope="col" className="text-left font-normal w-6">
                        {t('handLegend.rankCol')}
                      </th>
                      <th scope="col" className="text-left font-normal">
                        {t('handLegend.handCol')}
                      </th>
                      <th scope="col" className="text-left font-normal">
                        {t('handLegend.descCol')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {PRIMERO_HAND_RANKING.map((key, i) => {
                      const isCurrent = humanPlayer?.handName?.toLowerCase() === key;
                      return (
                        <tr
                          key={key}
                          data-testid={`primero-hand-legend-row-${key}`}
                          aria-current={isCurrent ? 'true' : undefined}
                          className={isCurrent ? 'text-ds-accent font-semibold' : ''}
                        >
                          <td className="align-top">{i + 1}</td>
                          <td className="align-top pr-2 whitespace-nowrap">
                            {t(`hand.${key}`)}
                            {isCurrent ? <span className="ml-1">{`← ${t('handLegend.current')}`}</span> : null}
                          </td>
                          <td className="align-top">{t(`handLegend.${key}`)}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
                <div className="mt-1">{t('handLegend.pointsNote')}</div>
              </div>
            </details>

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="primero-action-buttons">
              {isBettingPhase && isHumanTurn && !isGameEnd && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleCall} disabled={loading}>
                    {t('callButton')}
                  </button>
                  {/* **レイズが消えた理由を書く。**回数上限とチップ不足を
                      区別できないと、突然選択肢を奪われたように見える (#4925)。 */}
                  <span className="text-ds-text-muted text-xs" data-testid="primero-raise-count">
                    {state.raiseCount >= state.maxRaises
                      ? t('raiseCapReached', { max: state.maxRaises })
                      : t('raiseCount', { count: state.raiseCount, max: state.maxRaises })}
                  </span>
                  {state.canRaise && (
                    <button type="button" className={btnSuccess} onClick={handleRaise} disabled={loading}>
                      {t('raiseButton')}
                    </button>
                  )}
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
                dataTutorial="primero-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
