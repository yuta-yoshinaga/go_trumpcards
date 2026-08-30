import { useMemo } from 'react';
import type { trexApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useTrexGame } from '../hooks/useTrexGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TrexResponse } from '../types/card';
import { TrexPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTrexCommand, TREX_HELP } from '../utils/cli/commands/trexCommands';
import { formatTrexState } from '../utils/cli/formatters/trexFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { trexIsPenaltyCard } from '../utils/trexPenaltyCards';

const TX_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tx-rule"]', messageKey: 'tutorial.contracts', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tx-header"]', messageKey: 'tutorial.kingdom', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tx-table"]', messageKey: 'tutorial.dominoes', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tx-seats"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
];

const SUIT_GLYPHS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Renders the Trex page: the king picks one of five contracts, twenty deals in all. */
export const TrexPage = withTutorial(TrexPageContent, 'trex', TX_TUTORIAL_STEPS);

function TrexPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('trex');
  const game = useTrexGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('trex');
  const cliConfig: CliGameConfig<TrexResponse, Parameters<typeof trexApi.exec>> = useMemo(
    () => ({
      gameName: 'trex',
      parseCommand: parseTrexCommand,
      formatResponse: formatTrexState,
      helpText: TREX_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('trex', state);

  if (!state) {
    return <GameSkeleton gameKey="trex" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === TrexPhase.GAME_END;
  const choosing = state.phase === TrexPhase.CHOOSE;
  const dealOver = state.phase === TrexPhase.DEAL_END;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;
  const humanIsKing = state.kingIdx === 0;

  // Playability comes from the server: each contract follows a different rule,
  // and the dominoes are not a trick game at all.
  const playable = new Set(state.validIndices);

  const contractName = (n: number) => t(`contract${n}`, { defaultValue: t('contractNone') });

  const phaseName = ended
    ? t('phase.end')
    : dealOver
      ? t('phase.dealEnd')
      : choosing
        ? t('phase.choose')
        : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.trex')}
      gameThemeBg={gameTheme.trex.bg}
      phaseName={phaseName}
      gamePath="/trex"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted" data-tutorial="tx-header">
            {t('deal')}: {state.dealNo}/{state.totalDeals}
            {' / '}
            {t('king')}: {t('seat', { n: state.kingIdx })}
            {' / '}
            {t('contract')}: {contractName(state.contract)}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      <SettingsPanel
        title={tc('settings.title')}
        groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
      />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Permanent, not tutorial-only: "once per kingdom" and "the
                dominoes start from the JACK" are the two things a player gets
                wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="tx-rule">
              {t('ruleLine')}
            </div>

            <div className="flex justify-center gap-4 mb-3 flex-wrap" data-tutorial="tx-seats">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                    {' · '}
                    {t('score', { n: o.score })}
                    {o.id === state.kingIdx && ` · ${t('king')}`}
                  </div>
                  <div
                    className="flex gap-1 justify-center flex-wrap"
                    role="img"
                    aria-label={t('opponentHandAriaLabel', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                  >
                    {Array.from({ length: o.cardCount }, (_, i) => (
                      <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="text-center mb-4" data-tutorial="tx-table">
              {state.isTrix ? (
                <>
                  <div className="text-game-text-muted text-xs mb-1">{t('runs')}</div>
                  <div className="flex gap-3 justify-center flex-wrap text-sm">
                    {state.runs.map((run) => (
                      <span key={`run-${run.suit.toString()}`} data-testid="trex-run">
                        {SUIT_GLYPHS[run.suit] ?? '?'}{' '}
                        {run.started ? `${run.low.toString()}–${run.high.toString()}` : t('runEmpty')}
                      </span>
                    ))}
                  </div>
                </>
              ) : (
                <>
                  <div className="text-game-text-muted text-xs mb-1">{t('trick')}</div>
                  <div className="flex gap-1 justify-center flex-wrap">
                    {state.trick.length === 0 ? (
                      <span className="text-game-text-muted text-xs">—</span>
                    ) : (
                      state.trick.map((tc2) => {
                        // **失点対象は契約ごとに違う。**5 種が 1 王国内で切り替わる
                        // ので、どれが危険かを記憶と暗算で追わせることになる (#4911)。
                        const penalty = trexIsPenaltyCard(tc2.card, state.contract);
                        return (
                          <div
                            key={`trick-${tc2.playerIdx.toString()}`}
                            className={penalty ? 'rounded ring-2 ring-ds-error' : ''}
                            data-testid={penalty ? 'trex-penalty-card' : undefined}
                            title={penalty ? t('penaltyCard') : undefined}
                          >
                            <AnimatedCard card={tc2.card} width={cardWidth} draggable={false} />
                          </div>
                        );
                      })
                    )}
                  </div>
                </>
              )}
            </div>

            <div className="text-center" data-tutorial="tx-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('score', { n: human?.score ?? 0 })}
                {' · '}
                {t('dealScore', { n: human?.dealScore ?? 0 })}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  const canPlay = isHumanTurn && !choosing && playable.has(i);
                  const penalty = trexIsPenaltyCard(card, state.contract);
                  return (
                    <button
                      key={`hand-${i.toString()}`}
                      type="button"
                      data-hint-action="play"
                      data-testid={penalty ? 'trex-hand-penalty-card' : undefined}
                      title={penalty ? t('penaltyCard') : undefined}
                      // Kept focusable while it cannot act so the reason is
                      // announced rather than the control leaving the tab order.
                      aria-disabled={!canPlay}
                      onClick={() => canPlay && game.handlePlay(i)}
                      className={[
                        'rounded transition-transform',
                        canPlay ? 'hover:-translate-y-2' : 'opacity-60',
                        frontendHintEnabled && state.hint?.cardIndex === i
                          ? 'ring-2 ring-ds-warning'
                          : penalty
                            ? 'ring-2 ring-ds-error'
                            : '',
                      ].join(' ')}
                    >
                      <AnimatedCard card={card} width={cardWidth} draggable={false} />
                    </button>
                  );
                })}
              </div>
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
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

          <GameFooter className={`${gameTheme.trex.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />

            {choosing && humanIsKing && (
              <div className="mb-2" data-tutorial="tx-choose">
                <div className="text-xs text-ds-text-muted mb-1">{t('choosePrompt')}</div>
                <div className="flex gap-2 flex-wrap items-center">
                  {/* Only what this king has not played. A contract is played
                      once per kingdom, so offering a spent one is a dead end. */}
                  {state.availableContracts.map((n) => (
                    <button
                      key={`contract-${n.toString()}`}
                      type="button"
                      data-hint-action="choose"
                      className={`${btnPrimary} min-h-11`}
                      onClick={() => game.handleChoose(n)}
                    >
                      {contractName(n)}
                    </button>
                  ))}
                </div>
              </div>
            )}
            {choosing && !humanIsKing && (
              <div className="mb-2 text-xs text-ds-text-muted">{t('waitingForKing', { n: state.kingIdx })}</div>
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {/* Enabled only in the dominoes with no legal play. Trick
                  contracts have no pass at all. */}
              {state.isTrix && (
                <button
                  type="button"
                  data-hint-action="pass"
                  aria-disabled={!state.canPass}
                  onClick={() => state.canPass && game.handlePass()}
                  className={[
                    'px-4 py-2 rounded font-bold min-h-11',
                    state.canPass ? 'bg-ds-accent text-ds-on-accent' : 'bg-ds-surface-2 text-ds-text-muted',
                  ].join(' ')}
                >
                  {t('pass')}
                </button>
              )}
              {state.canPass && <span className="text-xs text-ds-text-muted">{t('passHint')}</span>}
              {dealOver && !ended && (
                <button type="button" className={`${btnSecondary} min-h-11`} onClick={game.handleNextDeal}>
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tx-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
