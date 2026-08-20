import { motion } from 'framer-motion';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ExchangeLog } from '../components/common/ExchangeLog';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePresidentGame } from '../hooks/usePresidentGame';
import { gameTheme } from '../styles/gameTheme';
import type { PresidentAction, PresidentResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { cardLabel } from '../utils/cardUtils';
import {
  formatPresidentState,
  PRESIDENT_HELP,
  type PresidentCliArgs,
  parsePresidentCommand,
} from '../utils/cli/commands/presidentCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tutorial steps for President. */
const PR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pr-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="pr-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pr-play-pass"]',
    messageKey: 'tutorial.playPass',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PRESIDENT_RANK_KEYS: Readonly<Record<number, string>> = {
  1: 'rank.president',
  2: 'rank.vicePresident',
  3: 'rank.viceScum',
  4: 'rank.scum',
};

const PRESIDENT_RANK_ICON: Readonly<Record<number, string>> = {
  1: '👑',
  2: '🥈',
  3: '🥉',
  4: '🗑️',
};

const PRESIDENT_RANK_BG: Readonly<Record<number, string>> = {
  1: 'bg-ds-warning text-ds-text-on-accent',
  2: 'bg-ds-info text-ds-text-on-accent',
  3: 'bg-ds-surface text-ds-text-muted',
  4: 'bg-ds-error text-ds-text-on-accent',
};

/** Stamp-style rank badge that springs in when a player finishes. */
function RankStamp({ rank, label }: { rank: number; label: string }) {
  return (
    <motion.span
      data-testid={`rank-stamp-${rank}`}
      initial={{ scale: 2, opacity: 0, rotate: -20 }}
      animate={{ scale: 1, opacity: 1, rotate: -8 }}
      transition={{ type: 'spring', stiffness: 260, damping: 14 }}
      className={`inline-block px-2 py-0.5 rounded-md font-bold text-xs border-2 border-current ${PRESIDENT_RANK_BG[rank] ?? 'bg-ds-surface text-ds-text-primary'}`}
    >
      {PRESIDENT_RANK_ICON[rank] ?? ''} {label}
    </motion.span>
  );
}

/** Renders the President (プレジデント) game page. */
export const PresidentPage = withTutorial(PresidentPageContent, 'president', PR_TUTORIAL_STEPS);
function PresidentPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('president');
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
  } = usePresidentGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('president', state);

  // One-shot full-screen flash when revolution toggles on (false → true), to
  // make the strength inversion impossible to miss. flashKey re-triggers the
  // CSS animation without AnimatePresence; the timer hides the scrim after 400ms.
  // The reverse transition (true → false, i.e. revolution reverting to normal)
  // fires a distinct info-coloured status banner so it is never mistaken for the
  // activation flash and is announced to screen readers via role="status".
  const [revolutionFlash, setRevolutionFlash] = useState(0);
  const [revolutionEndFlash, setRevolutionEndFlash] = useState(0);
  const prevRevolution = useRef(false);
  const revolutionFlashTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const revolutionEndFlashTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(
    () => () => {
      clearTimeout(revolutionFlashTimer.current ?? undefined);
      clearTimeout(revolutionEndFlashTimer.current ?? undefined);
    },
    [],
  );
  useEffect(() => {
    const active = state?.revolutionActive ?? false;
    if (active && !prevRevolution.current) {
      setRevolutionFlash((k) => k + 1);
      clearTimeout(revolutionFlashTimer.current ?? undefined);
      revolutionFlashTimer.current = setTimeout(() => setRevolutionFlash(0), 400);
    } else if (!active && prevRevolution.current) {
      setRevolutionEndFlash((k) => k + 1);
      clearTimeout(revolutionEndFlashTimer.current ?? undefined);
      revolutionEndFlashTimer.current = setTimeout(() => setRevolutionEndFlash(0), 1800);
    }
    prevRevolution.current = active;
  }, [state?.revolutionActive]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('president');
  const cliConfig: CliGameConfig<PresidentResponse, PresidentCliArgs> = useMemo(
    () => ({
      gameName: 'president',
      parseCommand: parsePresidentCommand,
      formatResponse: formatPresidentState,
      helpText: [...PRESIDENT_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  if (!state || state.players.length < 4) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.president.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanWon = isGameEnd && state.players[0]?.rank === 1;
  const isHumanTurn = state.currentTurn === 0 && !isGameEnd;
  const human = state.players[0];
  // 自分の手と CPU の手を 1 本の時系列にする。CUI と同じ書き分け
  // (出した / パスした) を使う。
  const describeAction = (action: PresidentAction): string => {
    const name = action.playerIdx === 0 ? tc('player.you') : tc('player.cpu', { id: action.playerIdx });
    if (!action.playedCards || action.playedCards.length === 0) return t('actionPassed', { name });
    return t('actionPlayed', { name, cards: action.playedCards.map(cardLabel).join(', ') });
  };
  const actionHistory = [...(state.humanAction ? [state.humanAction] : []), ...(state.cpuActions ?? [])].map(
    describeAction,
  );
  const canPlay = isHumanTurn && selectedIndices.length > 0;
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.president')}
      gameThemeBg={gameTheme.president.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/president"
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
          {revolutionFlash > 0 && (
            <div
              key={revolutionFlash}
              data-testid="president-revolution-flash"
              aria-hidden="true"
              className="pointer-events-none fixed inset-0 z-40 bg-ds-warning/40 motion-safe:animate-[fadeIn_0.2s_ease-out] motion-reduce:hidden"
            />
          )}
          {revolutionEndFlash > 0 && (
            <div
              key={revolutionEndFlash}
              data-testid="president-revolution-end-flash"
              role="status"
              className="pointer-events-none fixed inset-x-0 top-16 z-40 flex justify-center px-4 motion-safe:animate-[fadeIn_0.2s_ease-out]"
            >
              <span className="rounded-full bg-ds-info px-4 py-2 font-semibold text-ds-text-on-accent shadow-lg">
                {t('flash.revolutionEnd')}
              </span>
            </div>
          )}
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {state.revolutionActive && (
              <div className="flex items-center justify-center gap-1.5 text-center text-ds-warning font-semibold">
                <span className="inline-block motion-safe:rotate-180 transition-transform" aria-hidden="true">
                  ⇅
                </span>
                {t('badge.revolution')}
              </div>
            )}

            {/* ラウンド開始時のカード交換ログ。CUI (PresidentCuiPresenter) と Daifugo は
                以前から出していたが、President の Web だけ state.exchangeActions を
                描画していなかった (#4745)。 */}
            {state.exchangeActions && state.exchangeActions.length > 0 && (
              <ExchangeLog ns="president" players={state.players} actions={state.exchangeActions} />
            )}

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="pr-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                      <span>{tc('player.cpu', { id: p.id })}</span>
                      {p.isFinished ? (
                        <RankStamp rank={p.rank} label={t(PRESIDENT_RANK_KEYS[p.rank] ?? 'rank.unknown')} />
                      ) : (
                        <span>— {p.cardCount}</span>
                      )}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 13) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="pr-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('label.tableCards')}</div>
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('label.tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => (
                    // Wrap each display-only table card in a role="img" group so
                    // screen readers announce it as a single named unit (the card),
                    // independent of the inner <img alt>. Visuals are unchanged.
                    <span key={i} role="img" aria-label={cardAlt(c)}>
                      <AnimatedCard card={c} width={cardWidth * 0.9} />
                    </span>
                  ))
                )}
              </div>
            </div>

            {/* Action history: CUI は毎ターン「誰が何を出したか / パスしたか」を
                出しているのに、Web は場札しか見えず、CPU 3 人のうち誰が場をこの形に
                したのか追えなかった (#5548)。 */}
            {actionHistory.length > 0 && (
              <div
                role="log"
                aria-live="polite"
                aria-label={t('label.actionLog')}
                className="bg-black/40 rounded-lg text-ds-text-primary py-2 px-3.5 my-2 whitespace-pre-line text-xs"
                data-testid="pr-action-log"
              >
                {actionHistory.join('\n')}
              </div>
            )}

            {/* Human hand */}
            <div className="text-center" data-tutorial="pr-player-hand">
              <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                <span>{tc('player.you')}</span>
                {human.isFinished && (
                  <RankStamp rank={human.rank} label={t(PRESIDENT_RANK_KEYS[human.rank] ?? 'rank.unknown')} />
                )}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && toggleCardSelection(i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      selectedIndices.includes(i) ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${i}`}
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
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
                    value: String(configInput.cpuDifficulty ?? 1),
                    options: [
                      { value: '0', label: t('settings.difficulty.easy') },
                      { value: '1', label: t('settings.difficulty.normal') },
                      { value: '2', label: t('settings.difficulty.hard') },
                    ],
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'revolutionEnabled',
                    label: t('settings.revolution'),
                    checked: configInput.revolutionEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('revolutionEnabled', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'cardExchangeEnabled',
                    label: t('settings.cardExchange'),
                    checked: configInput.cardExchangeEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('cardExchangeEnabled', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'passFieldFlushEnabled',
                    label: t('settings.passFieldFlush'),
                    checked: configInput.passFieldFlushEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('passFieldFlushEnabled', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />
          <ReplaySpeedSettingsPanel />

          <GameFooter className={`${gameTheme.president.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="pr-play-pass">
              <button
                type="button"
                onClick={handlePlay}
                disabled={loading || !canPlay}
                className="px-4 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="play-button"
              >
                {t('button.play')}
              </button>
              <button
                type="button"
                onClick={handlePass}
                disabled={loading || !isHumanTurn}
                className="px-4 py-2 rounded-lg bg-ds-warning hover:bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="pass-button"
              >
                {t('button.pass')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="pr-reset-button"
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
