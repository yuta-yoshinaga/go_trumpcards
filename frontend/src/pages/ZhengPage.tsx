import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useZhengGame } from '../hooks/useZhengGame';
import { gameTheme } from '../styles/gameTheme';
import type { ZhengResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import {
  formatZhengState,
  parseZhengCommand,
  ZHENG_HELP,
  type ZhengCliArgs,
} from '../utils/cli/commands/zhengCommands';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isValidZhengCombo, zhengInvalidReason } from '../utils/zhengComboValidator';

/** Tutorial steps for Zheng Shangyou. */
const ZHENG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="zheng-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="zheng-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="zheng-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="zheng-play-pass"]',
    messageKey: 'tutorial.playPass',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="zheng-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Zheng Shangyou (争上游) game page. */
export const ZhengPage = withTutorial(ZhengPageContent, 'zheng', ZHENG_TUTORIAL_STEPS);
function ZhengPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('zheng');
  const {
    state,
    loading,
    error,
    callApi,
    selectedIndices,
    toggleCardSelection,
    configInput,
    handleConfigChange,
    handlePlay,
    handlePass,
    handleResetWithConfig,
    retry,
  } = useZhengGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('zheng', state);

  // Announce turn arrival, the lead-can't-pass rule, and the human's own finish —
  // CPU turns advance silently otherwise, leaving SR users unaware their turn came.
  const [liveMsg, setLiveMsg] = useState('');
  const prevAnnounceRef = useRef<{ turn: boolean; finished: boolean } | null>(null);
  useEffect(() => {
    if (!state || state.players.length < 4) return;
    const turn = state.currentTurn === 0 && !state.gameEndFlag;
    const finished = state.players[0]?.isFinished ?? false;
    const lead = state.tableCards.length === 0;
    const prev = prevAnnounceRef.current;
    prevAnnounceRef.current = { turn, finished };
    if (!prev) return;
    if (finished && !prev.finished) {
      setLiveMsg(t('announce.finished', { rank: t(`rank.${state.players[0]?.rank}`) }));
    } else if (turn && !prev.turn) {
      setLiveMsg(lead ? t('announce.yourTurnLead') : t('announce.yourTurn'));
    }
  }, [state, t]);

  // Values match the Go domain constants: 0=Normal, 1=Easy, 2=Hard (ZhengConfig.go).
  const difficultyOptions = useMemo(
    () => [
      { value: '0', label: t('settings.difficultyNormal') },
      { value: '1', label: t('settings.difficultyEasy') },
      { value: '2', label: t('settings.difficultyHard') },
    ],
    [t],
  );

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('zheng');
  const cliConfig: CliGameConfig<ZhengResponse, ZhengCliArgs> = useMemo(
    () => ({
      gameName: 'zheng',
      parseCommand: parseZhengCommand,
      formatResponse: formatZhengState,
      helpText: [...ZHENG_HELP],
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(frontendHint) : null),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  if (!state || state.players.length < 4) {
    return (
      <GameSkeleton
        gameKey="zheng"
        layout={{
          kind: 'trick-taking',
          titleBar: false,
          opponents: 3,
          opponentStyle: 'hand',
          opponentHandSize: 4,
          trickArea: true,
          footerHandSize: 5,
          footerButton: 'wide',
        }}
      />
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanWon = isGameEnd && state.players[0]?.rank === 1;
  const isHumanTurn = state.currentTurn === 0 && !isGameEnd;
  const isLead = state.tableCards.length === 0;
  const human = state.players[0];
  const selectedCards = selectedIndices.map((i) => human.cards[i]).filter((c): c is NonNullable<typeof c> => c != null);
  const hasValidCombo = isValidZhengCombo(selectedCards);
  const canPlay = isHumanTurn && selectedIndices.length > 0 && hasValidCombo;
  // Warn-only: explain WHY the selection can't be played (invalid type, wrong
  // count, doesn't beat the table, …). The play button gating is unchanged; the
  // backend remains the authority, this only guides the human toward a fix.
  const invalidReason =
    isHumanTurn && selectedIndices.length > 0
      ? zhengInvalidReason(selectedCards, state.tableCards, state.tablePlayType)
      : null;
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');

  // Standings: finished players first (by confirmed rank), then still-playing
  // players ordered by fewest cards left. Shown once anyone has gone out so the
  // remaining race (avoiding last place) is visible without waiting for game end.
  const anyFinished = state.players.some((p) => p.isFinished);
  const rankedPlayers = [...state.players].sort((a, b) => {
    if (a.isFinished && b.isFinished) return a.rank - b.rank;
    if (a.isFinished !== b.isFinished) return a.isFinished ? -1 : 1;
    return a.cardCount - b.cardCount;
  });

  return (
    <GamePageShell
      title={tc('nav.zheng')}
      gameThemeBg={gameTheme.zheng.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/zheng"
      gameEndFlag={!!isGameEnd}
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="zheng-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                      <span>{tc('player.cpu', { id: p.id })}</span>
                      {p.isFinished ? (
                        <span className="font-bold">{t(`rank.${p.rank}`)}</span>
                      ) : (
                        <span>— {p.cardCount}</span>
                      )}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 14) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Standings — finished players' confirmed ranks + remaining card counts */}
            {anyFinished && (
              <div className="mx-auto max-w-xs rounded-lg bg-black/20 px-3 py-2" data-testid="zheng-rank-table">
                <div className="mb-1 text-center text-xs text-ds-text-muted">{t('rankTable.title')}</div>
                <ul className="space-y-0.5">
                  {rankedPlayers.map((p) => (
                    <li
                      key={p.id}
                      className="flex items-center justify-between text-xs"
                      data-testid={`zheng-rank-row-${p.id}`}
                    >
                      <span>{p.isHuman ? tc('player.you') : tc('player.cpu', { id: p.id })}</span>
                      {p.isFinished ? (
                        <span className="font-bold text-ds-text">{t(`rank.${p.rank}`)}</span>
                      ) : (
                        <span className="text-ds-text-muted">{t('cardCount', { count: p.cardCount })}</span>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="zheng-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('tableCards')}</div>
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.9} />)
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="zheng-player-hand">
              <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                <span>{tc('player.you')}</span>
                {human.isFinished && <span className="font-bold">{t(`rank.${human.rank}`)}</span>}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => {
                  const selected = selectedIndices.includes(i);
                  const cardClass = selected
                    ? isHumanTurn
                      ? 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-pointer hover:opacity-90'
                      : 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-default'
                    : isHumanTurn
                      ? 'rounded transition-all cursor-pointer hover:opacity-90'
                      : 'rounded transition-all cursor-default';
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => isHumanTurn && toggleCardSelection(i)}
                      disabled={!isHumanTurn}
                      className={cardClass}
                      data-testid={`hand-card-${i}`}
                    >
                      <AnimatedCard card={c} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(configInput.cpuDifficulty ?? 0),
                    options: difficultyOptions,
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />
          <ReplaySpeedSettingsPanel />

          <GameFooter className={`${gameTheme.zheng.footer} px-4 py-2.5`}>
            {/* Polite live region for turn/pass/finish transitions. */}
            <span className="sr-only" role="status" aria-live="polite" data-testid="zheng-turn-announce">
              {liveMsg}
            </span>
            {invalidReason && (
              <p
                role="status"
                data-testid="zheng-invalid-combo"
                className="mb-1 text-center font-medium text-ds-warning text-xs"
              >
                {t(`invalidReason.${invalidReason}`)}
              </p>
            )}
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="zheng-play-pass">
              <button
                type="button"
                onClick={handlePlay}
                disabled={loading || !canPlay}
                className="px-4 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="play-button"
              >
                {t('playButton')}
              </button>
              <button
                type="button"
                onClick={handlePass}
                disabled={loading || !isHumanTurn || isLead}
                className="px-4 py-2 rounded-lg bg-ds-warning hover:bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="pass-button"
              >
                {t('passButton')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="zheng-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
