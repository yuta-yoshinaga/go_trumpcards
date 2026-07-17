import { useEffect, useMemo, useState } from 'react';
import { cuckooApi } from '../api/gameApi';
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
import { KbdBadge } from '../components/KbdBadge';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CuckooResponse } from '../types/card';
import { CuckooPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CUCKOO_HELP, parseCuckooCommand } from '../utils/cli/commands/cuckooCommands';
import { formatCuckooState } from '../utils/cli/formatters/cuckooFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** CPU difficulty options for the Cuckoo settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
];

/** Cuckoo tutorial step definitions. */
const CUCKOO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cuckoo-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuckoo-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuckoo-card"]',
    messageKey: 'tutorial.card',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuckoo-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuckoo-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Rank value of a King (A=1 … K=13). Only a King holder may refuse an incoming
 * swap; mirrors `CuckooKingValue` in `internal/domain/Cuckoo.go`. */
const CUCKOO_KING_VALUE = 13;

const CUCKOO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CuckooPhase.TURN]: 'turn',
  [CuckooPhase.REFUSE]: 'refuse',
  [CuckooPhase.ROUND_END]: 'roundEnd',
  [CuckooPhase.GAME_END]: 'gameEnd',
};

/** Renders the Cuckoo game page: a 4-player life-survival game (Chase the Ace). */
export const CuckooPage = withTutorial(CuckooPageContent, 'cuckoo', CUCKOO_TUTORIAL_STEPS);

/** Inner content of the Cuckoo page, wrapped by TutorialProvider. */
function CuckooPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cuckoo');
  const { state, loading, error, exec, retry } = useGameApi(cuckooApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleDifficultyChange = (value: string) => {
    const level = Number(value);
    setCpuDifficulty(level);
    exec('reset', { config: { cpuDifficulty: level } });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cuckoo');
  const cliConfig: CliGameConfig<CuckooResponse, Parameters<typeof cuckooApi.exec>> = useMemo(
    () => ({
      gameName: 'cuckoo',
      parseCommand: parseCuckooCommand,
      formatResponse: formatCuckooState,
      helpText: CUCKOO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('cuckoo', CUCKOO_PHASE_KEYS);

  // Keyboard shortcuts: keep/swap on a TURN, refuse/accept when targeted for a swap.
  // Enabled only on the human's active turn (see the per-binding `enabled` flags).
  const kbIsHumanTurn = state?.phase === CuckooPhase.TURN && state.currentPlayerIdx === 0 && !state.gameEndFlag;
  const kbIsRefuseTarget = state?.phase === CuckooPhase.REFUSE && state.pendingSwapTo === 0 && !state.gameEndFlag;
  // Refusing is only legal for a King holder; the 'r' shortcut is gated on it so
  // it cannot fire a server error, while 'a' (accept) stays available regardless.
  const kbHumanHasKing = state?.players.find((p) => p.isHuman)?.card?.value === CUCKOO_KING_VALUE;
  const actionBindings = useMemo(
    () => [
      { key: 'k', action: () => exec('keep'), enabled: kbIsHumanTurn },
      { key: 's', action: () => exec('swap'), enabled: kbIsHumanTurn },
      { key: 'r', action: () => exec('refuse'), enabled: kbIsRefuseTarget && kbHumanHasKing },
      { key: 'a', action: () => exec('accept'), enabled: kbIsRefuseTarget },
    ],
    [exec, kbIsHumanTurn, kbIsRefuseTarget, kbHumanHasKing],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state)
    return <GameSkeleton gameKey="cuckoo" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 1 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isTurn = state.phase === CuckooPhase.TURN;
  const isRefuse = state.phase === CuckooPhase.REFUSE;
  const isRoundEnd = state.phase === CuckooPhase.ROUND_END;
  const isGameEnd = state.phase === CuckooPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isTurn && state.currentPlayerIdx === 0 && !isGameEnd;
  const isHumanRefuseTarget = isRefuse && state.pendingSwapTo === 0 && !isGameEnd;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  // The dealer swaps with the stock instead of the next player.
  const humanIsDealer = humanPlayer ? state.dealerIdx === humanPlayer.id : false;
  // Only a King holder may refuse an incoming swap; drives the refuse button's
  // dynamic label and disabled state.
  const humanHasKing = humanPlayer?.card?.value === CUCKOO_KING_VALUE;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handleManualReset = () => {
    hideActionLog();
    exec('reset', { config: { cpuDifficulty } });
  };

  return (
    <GamePageShell
      title={tc('nav.cuckoo')}
      gameThemeBg={gameTheme.cuckoo.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanRefuseTarget) && !isGameEnd}
      gamePath="/cuckoo"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
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
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    onSelect: handleDifficultyChange,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="cuckoo-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">
                {t('dealer')}: {playerLabel(state.dealerIdx, state.dealerIdx === 0)}
              </span>
              <span>{t('stock', { count: state.stockCount })}</span>
            </div>

            {/* Players (lives / eliminated / current-turn) */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="cuckoo-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''} ${p.isEliminated ? 'opacity-50' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span role="img" aria-label={p.lives > 0 ? t('livesCount', { count: p.lives }) : t('eliminated')}>
                    {p.lives > 0 ? '♥'.repeat(p.lives) : `(${t('eliminated')})`}
                  </span>
                  {p.kingRevealed && <span className="text-ds-accent">[{t('kingRevealed')}]</span>}
                  {!p.isHuman && p.card && !p.isEliminated && <CardImage card={p.card} width={cardWidth} />}
                </div>
              ))}
            </div>

            {/* Round losers */}
            {(isRoundEnd || isGameEnd) && state.roundLosers.length > 0 && (
              <div
                className="mb-2 p-2 rounded bg-ds-warning/20 text-ds-warning text-sm"
                data-tutorial="cuckoo-losers"
                data-testid="cuckoo-losers"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('roundLosersTitle')}</div>
                {state.roundLosers.map((idx) => (
                  <div key={`loser-${idx}`}>{t('roundLoser', { name: playerLabel(idx, idx === 0) })}</div>
                ))}
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
          <GameFooter className={`${gameTheme.cuckoo.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="cuckoo-card">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourCard')}</div>
              {humanPlayer?.card && !humanPlayer.isEliminated ? (
                <CardImage card={humanPlayer.card} width={cardWidth} />
              ) : (
                <div className="text-ds-text-muted text-sm">{t('eliminated')}</div>
              )}
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanTurn && <div className="text-ds-text-muted text-xs mb-2">{t('turnNotice')}</div>}
            {isHumanRefuseTarget && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="cuckoo-refuse-notice">
                {humanHasKing ? t('refuseNoticeKing') : t('refuseNoticeNoKing')}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="cuckoo-actions">
              {isHumanTurn && (
                <>
                  <button type="button" className={btnSuccess} onClick={() => exec('keep')} disabled={loading}>
                    {t('keepButton')}
                    <KbdBadge label="K" />
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => exec('swap')} disabled={loading}>
                    {humanIsDealer ? t('swapDealerButton') : t('swapButton')}
                    <KbdBadge label="S" />
                  </button>
                </>
              )}

              {isHumanRefuseTarget && (
                <>
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={() => exec('refuse')}
                    disabled={loading || !humanHasKing}
                    title={humanHasKing ? undefined : t('refuseNoKingReason')}
                    data-testid="cuckoo-refuse-button"
                  >
                    {humanHasKing ? t('refuseKingButton') : t('refuseButton')}
                    <KbdBadge label="R" />
                  </button>
                  <button type="button" className={btnSuccess} onClick={() => exec('accept')} disabled={loading}>
                    {t('acceptButton')}
                    <KbdBadge label="A" />
                  </button>
                </>
              )}

              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('nextround')} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cuckoo-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
