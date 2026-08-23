import { useCallback, useEffect, useMemo } from 'react';
import { julepeApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { JulepePlayer, JulepeResponse } from '../types/card';
import { JulepePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { JULEPE_HELP, parseJulepeCommand } from '../utils/cli/commands/julepeCommands';
import { formatJulepeState } from '../utils/cli/formatters/julepeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tricks per round (five cards each). */
const TRICKS_PER_ROUND = 5;

/**
 * The extra payment a player owes for entering and falling short of the
 * required trick count.
 *
 * **The requirement is not fixed** — it scales with how many players stay in,
 * so the count itself comes from the server (`requiredTricks`) rather than
 * being hard-coded here alongside this penalty.
 */
const MISS_PENALTY = 5;

/** Guided tutorial steps (pot, decision, trick, hand). */
const JULEPE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="julepe-pot"]', messageKey: 'tutorial.pot', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="julepe-decide"]', messageKey: 'tutorial.decide', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="julepe-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="julepe-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Julepe page (wrapped by `withTutorial`).
 *
 * Renders the pot game for **three to five players** — the table size is
 * configurable, which is why nothing here assumes four seats. Five cards each
 * on a 32-card piquet pack, trump from the card turned after the deal.
 *
 * The distinctive control is the **play-or-drop decision** taken before any
 * card is played: dropping costs only the ante, while entering and falling
 * short of the required trick count costs an extra payment into the pot *and*
 * doubles your ante next hand (a "beast"). The page therefore leads with the
 * pot, the trump card, the required trick count and that penalty, since none of
 * them can be read off the table.
 */
function JulepePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('julepe');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<JulepeResponse, Parameters<typeof julepeApi.exec>>(julepeApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('julepe', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('julepe');
  const cliConfig: CliGameConfig<JulepeResponse, Parameters<typeof julepeApi.exec>> = useMemo(
    () => ({
      gameName: 'julepe',
      parseCommand: parseJulepeCommand,
      formatResponse: formatJulepeState,
      helpText: JULEPE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlayCard = useCallback(
    (idx: number) => {
      void dispatch('card', idx);
    },
    [dispatch],
  );

  const handlePlayIn = useCallback(() => {
    void dispatch('in');
  }, [dispatch]);

  const handlePassOut = useCallback(() => {
    void dispatch('out');
  }, [dispatch]);

  const handleNextRound = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="julepe" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isDecide = state.phase === JulepePhase.DECIDE;
  const isRoundEnd = state.phase === JulepePhase.ROUND_END;
  const isGameEnd = state.phase === JulepePhase.GAME_END || state.gameEndFlag;
  // 降りたラウンドは手番が回ってこない。押せる札も無い。
  const humanIsIn = human?.inRound === true;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isDecide && humanIsIn && state.players[state.currentPlayerIdx]?.isHuman === true;
  // 選択が済んでいて降りているあいだは「見ているだけ」。
  const isWatching = !isGameEnd && !isRoundEnd && !isDecide && human?.decided === true && !humanIsIn;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDecide
        ? t('phase.decide')
        : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  /** Whether a seat is in, out, or has yet to choose. */
  const statusStr = (p: JulepePlayer): string =>
    !p.decided ? t('status.undecided') : p.inRound ? t('status.in') : t('status.out');

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx < 0) return t('result.tie');
    const winner = state.players[state.winnerIdx];
    return state.winnerIdx === 0
      ? t('result.youWin', { chips: String(winner?.chips ?? 0) })
      : t('result.cpuWin', { idx: String(state.winnerIdx), chips: String(winner?.chips ?? 0) });
  })();

  return (
    <GamePageShell
      title={tc('nav.julepe')}
      gameThemeBg={gameTheme.julepe.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/julepe"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === 0}
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
          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4" data-testid="rm-round">
                {t('header.round')}: {state.roundNumber}/{state.config.rounds}
              </span>
              <span className="mr-4" data-testid="rm-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              {/* 人数は可変なので、何人卓かを必ず出す。 */}
              <span data-testid="rm-players">{t('header.players', { count: state.config.playerCnt })}</span>
            </div>

            <div
              className="mb-2 rounded bg-black/30 border border-ds-warning px-3 py-2 text-ds-text-primary text-sm text-center"
              data-testid="rm-pot"
              data-tutorial="julepe-pot"
            >
              {t('header.pot', { pot: String(state.pot) })}
            </div>

            {/* 切り札とリスクは参加判断の材料。盤面からは読めない。 */}
            <div className="mb-3 flex flex-wrap justify-center items-center gap-3">
              {state.upCard && (
                <div className="flex items-center gap-2 rounded bg-black/30 p-2" data-testid="rm-trump">
                  <span className="text-ds-text-muted text-sm">{t('header.trump')}</span>
                  <CardImage card={state.upCard} width={Math.round(cardWidth * 0.8)} />
                </div>
              )}
              <div className="rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm" data-testid="rm-risk">
                {t('header.risk', {
                  penalty: String(MISS_PENALTY),
                  required: String(state.requiredTricks),
                })}
              </div>
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`rm-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {/* **参加判断もリードも親の左隣から始まる** (#5748)。3〜5 人卓で
                      毎ラウンド 1 つ回るので、誰が親かが出ていないと自分が何番目に
                      決断するのか読めない。 */}
                  {p.id === state.dealerIdx && (
                    <span
                      className={`ml-1 rounded px-1.5 py-0.5 text-xs ${badgeWarningColors}`}
                      data-testid={`rm-dealer-${p.id.toString()}`}
                    >
                      <span aria-hidden="true">{t('dealerBadge')}</span>
                      <span className="sr-only">{t('dealerAria')}</span>
                    </span>
                  )}
                  {': '}
                  {t('header.seat', { chips: String(p.chips), tricks: String(p.roundTricks) })} [{statusStr(p)}]
                </div>
              ))}
            </div>

            <div data-tutorial="julepe-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {resultBanner && (
              <div className="text-center text-xl my-4 text-ds-accent font-semibold" role="status">
                {resultBanner}
              </div>
            )}

            {/* 降りたラウンドは操作しない。待ちに見えないよう明示する。 */}
            {isWatching && (
              <div
                className="my-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
                role="status"
                data-testid="rm-watching"
              >
                {t('watching')}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="julepe-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handlePlayCard(idx)}
                      disabled={loading || !isHumanTurn}
                      aria-label={t('actions.playAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${legalRing.has(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''}`}
                    >
                      <CardImage card={card} width={cardWidth} />
                    </button>
                  ))}
                </div>
              </div>
            )}

            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="julepe-decide">
              {/* 参加選択は配り直後の一度きり。 */}
              {isDecide && !isGameEnd && (
                <>
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={handlePlayIn}
                    disabled={loading}
                    data-testid="rm-in-btn"
                  >
                    {t('actions.playIn')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handlePassOut}
                    disabled={loading}
                    data-testid="rm-out-btn"
                  >
                    {t('actions.passOut')}
                  </button>
                </>
              )}
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('actions.nextRound')}
                </button>
              )}
              <button
                type="button"
                className={btnPrimary}
                onClick={() => requestConfirm(handleReset)}
                disabled={loading}
              >
                {t('actions.reset')}
              </button>
              {!isGameEnd && (
                <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                  {t('actions.giveUp')}
                </button>
              )}
            </div>

            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, hintEnabled, setHintEnabled)] }]}
            />
          </div>

          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
        </>
      )}
    </GamePageShell>
  );
}

/** Julepe page wrapped with TutorialProvider. */
export const JulepePage = withTutorial(JulepePageContent, 'julepe', JULEPE_TUTORIAL_STEPS);
