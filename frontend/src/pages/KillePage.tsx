import { useEffect, useMemo, useState } from 'react';
import { killeApi } from '../api/gameApi';
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
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KilleEvent, KillePlayer, KilleResponse } from '../types/card';
import { KillePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KILLE_HELP, parseKilleCommand } from '../utils/cli/commands/killeCommands';
import { formatKilleState } from '../utils/cli/formatters/killeFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Stake options for the Kille settings panel. */
const STAKE_OPTIONS = [1, 2, 5, 10];

/** Buy-backs allowed per player (sync: `KilleMaxReentries` in internal/domain/KillePlayer.go). */
const KILLE_MAX_REENTRIES = 3;

/**
 * The pack in descending strength, for the on-page reference.
 *
 * The numbers 11 down to 2 are elided — they sit between the Inn and the 1 in
 * the obvious order, and listing all twenty-one entries buries the six that
 * actually change what happens.
 */
const KILLE_LADDER = [
  { label: 'Harlequin', color: 'text-ds-accent' },
  { label: 'Cuckoo', color: 'text-ds-info' },
  { label: 'Hussar', color: 'text-ds-danger' },
  { label: 'Pig', color: 'text-ds-warning' },
  { label: 'Cavalier', color: 'text-ds-success' },
  { label: 'Inn', color: 'text-ds-success' },
  { label: '12 … 1', color: 'text-ds-text-muted' },
  { label: 'Wreath', color: 'text-ds-text-muted' },
  { label: 'Flowerpot', color: 'text-ds-text-muted' },
  { label: 'Mask', color: 'text-ds-text-muted' },
];

/** Kille tutorial step definitions. */
const KILLE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="kille-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-card"]',
    messageKey: 'tutorial.card',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kille-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KILLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KillePhase.EXCHANGE]: 'exchange',
  [KillePhase.SHOWDOWN]: 'showdown',
  [KillePhase.GAME_END]: 'gameEnd',
};

/** Renders the Kille game page: the Swedish Cuckoo game on its own 42-card pack. */
export const KillePage = withTutorial(KillePageContent, 'kille', KILLE_TUTORIAL_STEPS);

/** Inner content of the Kille page, wrapped by TutorialProvider. */
function KillePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('kille');
  const { state, loading, error, exec, retry } = useGameApi(killeApi.exec);

  const [stake, setStake] = useState(1);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleStakeChange = (value: string) => {
    const next = Number(value);
    setStake(next);
    exec('reset', { config: { stake: next } });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('kille');
  const cliConfig: CliGameConfig<KilleResponse, Parameters<typeof killeApi.exec>> = useMemo(
    () => ({
      gameName: 'kille',
      parseCommand: parseKilleCommand,
      formatResponse: formatKilleState,
      helpText: KILLE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('kille', KILLE_PHASE_KEYS);

  const kbIsHumanTurn =
    state?.phase === KillePhase.EXCHANGE &&
    state.currentPlayerIdx === 0 &&
    !state.gameEndFlag &&
    !state.players[0].isOut;
  const kbIsShowdown = state?.phase === KillePhase.SHOWDOWN && !state.gameEndFlag;
  const actionBindings = useMemo(
    () => [
      { key: 'e', action: () => exec('exchange'), enabled: !!kbIsHumanTurn },
      { key: 's', action: () => exec('satisfied'), enabled: !!kbIsHumanTurn },
      { key: 'n', action: () => exec('nextround'), enabled: !!kbIsShowdown },
    ],
    [exec, kbIsHumanTurn, kbIsShowdown],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state)
    return <GameSkeleton gameKey="kille" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 1 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isShowdown = state.phase === KillePhase.SHOWDOWN;
  const isGameEnd = state.phase === KillePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    state.phase === KillePhase.EXCHANGE && state.currentPlayerIdx === 0 && !isGameEnd && !humanPlayer?.isOut;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  // The dealer swaps with the stock rather than a neighbour, so its button says
  // something different and it can never be challenged.
  const humanIsDealer = humanPlayer ? state.dealerIdx === humanPlayer.id : false;
  const humanCanReenter = isShowdown && !isGameEnd && !!humanPlayer?.isOut && !!humanPlayer?.canReenter;
  const humanIsOutForGood = isShowdown && !!humanPlayer?.isOut && !humanPlayer?.canReenter;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  /** Why a seat went out. The Hussar and the Pig fire regardless of strength. */
  const outReason = (p: KillePlayer): string => {
    if (p.knockedBy === 'hussar') return t('outHussar');
    if (p.knockedBy === 'pig') return t('outPig');
    return t('outLowest');
  };

  const eventLine = (e: KilleEvent): string =>
    t(`event.${e.kind}`, {
      actor: playerLabel(e.actor, e.actor === 0),
      target: e.target >= 0 ? playerLabel(e.target, e.target === 0) : t('stock', { count: state.stockCount }),
      defaultValue: '',
    });

  const handleManualReset = () => {
    hideActionLog();
    exec('reset', { config: { stake } });
  };

  return (
    <GamePageShell
      title={tc('nav.kille')}
      gameThemeBg={gameTheme.kille.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/kille"
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
                    id: 'stake',
                    label: t('settings.stake'),
                    value: stake,
                    options: STAKE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: handleStakeChange,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="kille-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { count: state.pot })}</span>
              <span className="mr-4">
                {t('dealer')}: {playerLabel(state.dealerIdx, state.dealerIdx === 0)}
              </span>
              <span>{t('stock', { count: state.stockCount })}</span>
            </div>

            {/* The one rule that reliably surprises people. */}
            <div className="mb-2 text-center text-ds-text-muted text-xs" data-testid="kille-rules-note">
              {t('rulesNote')}
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="kille-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="kille-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''} ${p.isFinished ? 'opacity-50' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  {p.id === state.dealerIdx && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  <span>{t('chips', { n: p.chips })}</span>
                  {p.reentries > 0 && (
                    <span>{t('reentriesUsed', { used: p.reentries, max: KILLE_MAX_REENTRIES })}</span>
                  )}
                  {p.isSatisfied && !p.isOut && <span className="text-ds-success">[{t('satisfied')}]</span>}
                  {p.isOut && <span className="text-ds-danger">[{outReason(p)}]</span>}
                  {p.isFinished && <span>({t('eliminated')})</span>}
                  {!p.isHuman && p.card && <CardImage card={p.card} width={cardWidth} />}
                </div>
              ))}
            </div>

            {/* Exchanges this round */}
            {state.events.length > 0 && (
              <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="kille-events">
                <div className="mb-1 text-ds-text-primary">{t('eventsTitle')}</div>
                {state.events.map((e, i) => (
                  <div key={`event-${e.kind}-${e.actor}-${e.target}-${i}`} className="text-ds-text-muted">
                    {eventLine(e)}
                  </div>
                ))}
              </div>
            )}

            {/* Showdown */}
            {(isShowdown || isGameEnd) && state.loserIdxs.length > 0 && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="kille-showdown"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('showdownTitle')}</div>
                {state.loserIdxs.map((idx) => (
                  <div key={`loser-${idx}`}>
                    {t('loserLine', {
                      name: playerLabel(idx, idx === 0),
                      reason: outReason(state.players[idx]),
                    })}
                  </div>
                ))}
              </div>
            )}

            {/* The pack, for reference: a single suit means denomination is everything. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="kille-ladder">
              <div className="mb-1 text-ds-text-primary">{t('deckTitle')}</div>
              <div className="flex flex-wrap gap-x-2 gap-y-1">
                {KILLE_LADDER.map((r) => (
                  <span key={r.label} className={r.color}>
                    {r.label}
                  </span>
                ))}
              </div>
              <div className="mt-1 text-ds-text-muted">{t('deckNote')}</div>
            </div>

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
          <GameFooter className={`${gameTheme.kille.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="kille-card">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourCard')}</div>
              {humanPlayer?.card ? (
                <CardImage card={humanPlayer.card} width={cardWidth} />
              ) : (
                <div className="text-ds-text-muted text-sm">{t('out')}</div>
              )}
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanTurn && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="kille-turn-notice">
                {humanIsDealer ? t('dealerNotice') : t('turnNotice')}
              </div>
            )}
            {humanIsOutForGood && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="kille-reenter-exhausted">
                {t('reenterExhausted')}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="kille-actions">
              {isHumanTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => exec('exchange')} disabled={loading}>
                    {humanIsDealer ? t('exchangeDealerButton') : t('exchangeButton')}
                    <KbdBadge label="E" />
                  </button>
                  <button type="button" className={btnSuccess} onClick={() => exec('satisfied')} disabled={loading}>
                    {t('satisfiedButton')}
                    <KbdBadge label="S" />
                  </button>
                </>
              )}

              {humanCanReenter && (
                <button
                  type="button"
                  className={btnWarning}
                  onClick={() => exec('reenter')}
                  disabled={loading}
                  data-testid="kille-reenter-button"
                >
                  {t('reenterButton', { cost: humanPlayer?.reentryCost ?? 0 })}
                </button>
              )}

              {isShowdown && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('nextround')} disabled={loading}>
                  {t('nextRound')}
                  <KbdBadge label="N" />
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
                dataTutorial="kille-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
