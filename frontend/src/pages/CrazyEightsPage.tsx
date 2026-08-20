import { useCallback, useMemo, useState } from 'react';
import type { crazyeightsApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCrazyEightsGame } from '../hooks/useCrazyEightsGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CrazyEightsResponse } from '../types/card';
import { CrazyEightsPhase, CrazyEightsSuit } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CRAZYEIGHTS_HELP, parseCrazyeightsCommand } from '../utils/cli/commands/crazyeightsCommands';
import { formatCrazyeightsState } from '../utils/cli/formatters/crazyeightsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isCrazyEightsLegalPlay } from '../utils/crazyEightsLegal';
import {
  type CrazyEightsSortMode,
  loadCrazyEightsSortMode,
  saveCrazyEightsSortMode,
  sortedCrazyEightsHand,
} from '../utils/crazyEightsSort';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const CRAZYEIGHTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CrazyEightsPhase.PLAY]: 'play',
  [CrazyEightsPhase.CHOOSE_SUIT]: 'chooseSuit',
  [CrazyEightsPhase.ROUND_END]: 'roundEnd',
  [CrazyEightsPhase.GAME_END]: 'gameEnd',
};

const SUIT_BUTTONS = [
  { suit: CrazyEightsSuit.SPADE, key: 'suitSpade' },
  { suit: CrazyEightsSuit.CLOVER, key: 'suitClover' },
  { suit: CrazyEightsSuit.HEART, key: 'suitHeart' },
  { suit: CrazyEightsSuit.DIAMOND, key: 'suitDiamond' },
] as const;

const SUIT_SYMBOLS: Record<number, string> = {
  [CrazyEightsSuit.SPADE]: '♠',
  [CrazyEightsSuit.CLOVER]: '♣',
  [CrazyEightsSuit.HEART]: '♥',
  [CrazyEightsSuit.DIAMOND]: '♦',
};

/** Chosen-suit number → i18n key for the spoken suit name (used in the live-region announcement). */
const SUIT_NAME_KEYS: Record<number, string> = {
  [CrazyEightsSuit.SPADE]: 'suitName.spade',
  [CrazyEightsSuit.CLOVER]: 'suitName.clover',
  [CrazyEightsSuit.HEART]: 'suitName.heart',
  [CrazyEightsSuit.DIAMOND]: 'suitName.diamond',
};

/** DOM id linking each illegal card button to the shared screen-reader reason text. */
const ILLEGAL_REASON_ID = 'ce-illegal-reason';

/** Hand sort options for the Crazy Eights footer. */
const CRAZYEIGHTS_SORT_MODES: { mode: CrazyEightsSortMode; labelKey: string }[] = [
  { mode: 'original', labelKey: 'sort.original' },
  { mode: 'rank', labelKey: 'sort.rank' },
  { mode: 'suit', labelKey: 'sort.suit' },
];

/** Crazy Eights tutorial step definitions. */
const CE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ce-discard-pile"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ce-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ce-play-draw"]', messageKey: 'tutorial.playDraw', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ce-suit-choice"]',
    messageKey: 'tutorial.suitChoice',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ce-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Crazy Eights game page with card play and suit selection. */
export const CrazyEightsPage = withTutorial(CrazyEightsPageContent, 'crazyeights', CE_TUTORIAL_STEPS);
/** Inner content of the Crazy Eights page, wrapped by TutorialProvider. */
function CrazyEightsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('crazyeights');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    crazyEightsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleChooseSuit,
    handleNextRound,
    hint: serverHint,
    hintError,
    handleHint,
  } = useCrazyEightsGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('crazyeights', state);
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('crazyeights');
  const cliConfig: CliGameConfig<CrazyEightsResponse, Parameters<typeof crazyeightsApi.exec>> = useMemo(
    () => ({
      gameName: 'crazyeights',
      parseCommand: parseCrazyeightsCommand,
      formatResponse: formatCrazyeightsState,
      helpText: CRAZYEIGHTS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Display-only hand sort: reorders the rendered hand while every click maps
  // back to the card's ORIGINAL server index, so selection/play stay correct.
  const [sortMode, setSortMode] = useState<CrazyEightsSortMode>(loadCrazyEightsSortMode);
  const handleSortMode = useCallback((mode: CrazyEightsSortMode) => {
    setSortMode(mode);
    saveCrazyEightsSortMode(mode);
  }, []);

  const isPlayPhaseForKbd = state?.phase === CrazyEightsPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardsForKbd = state?.players.find((p) => p.isHuman)?.cards;
  const humanCardCountForKbd = humanCardsForKbd?.length ?? 0;

  // Original server indices in current display order, so the digit keys select
  // the visually Nth card (following the sort) while still toggling its real index.
  const displayIndexOrder = useMemo(
    () => (humanCardsForKbd ? sortedCrazyEightsHand(humanCardsForKbd, sortMode).map((e) => e.index) : []),
    [humanCardsForKbd, sortMode],
  );
  const toggleByDisplayIndex = useCallback(
    (displayIdx: number) => {
      const original = displayIndexOrder[displayIdx];
      if (original !== undefined) toggleCard(original);
    },
    [displayIndexOrder, toggleCard],
  );

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, undefined, {
      cpuDifficulty: crazyEightsConfig.cpuDifficulty,
      pointLimit: crazyEightsConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, crazyEightsConfig.cpuDifficulty, crazyEightsConfig.pointLimit]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleByDisplayIndex,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('crazyeights', CRAZYEIGHTS_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton
        gameKey="crazyeights"
        layout={{ kind: 'trick-taking', centerCard: true, trickArea: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === CrazyEightsPhase.PLAY;
  const isChooseSuit = state.phase === CrazyEightsPhase.CHOOSE_SUIT;
  const isRoundEnd = state.phase === CrazyEightsPhase.ROUND_END;
  const isGameEnd = state.phase === CrazyEightsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  // Announce the active suit whenever an 8 has changed it. The watermark and the
  // sidebar readout are aria-hidden / static, so this sr-only live region is the
  // only channel that tells a screen-reader user a suit was chosen (e.g. by a CPU).
  const suitNameKey = SUIT_NAME_KEYS[state.chosenSuit];
  const suitAnnouncement = state.chosenSuit > 0 && suitNameKey ? t('suitChanged', { suit: t(suitNameKey) }) : '';

  return (
    <GamePageShell
      title={tc('nav.crazyeights')}
      gameThemeBg={gameTheme.crazyeights.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isChooseSuit}
      gamePath="/crazyeights"
      gameEndFlag={!!state.gameEndFlag}
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
                    value: crazyEightsConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: crazyEightsConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            {/* Always-rendered polite live region announcing the active suit after an 8.
                Empty when no suit is chosen so the announcement fires on the transition. */}
            <div className="sr-only" role="status" aria-live="polite" data-testid="ce-suit-live-region">
              {suitAnnouncement}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top */}
                {state.discardTop && (
                  <div
                    className="my-3 p-3 rounded bg-black/40 flex items-center gap-3 relative overflow-hidden"
                    data-tutorial="ce-discard-pile"
                  >
                    {state.chosenSuit > 0 && (
                      <span
                        aria-hidden="true"
                        data-testid="chosen-suit-watermark"
                        className="pointer-events-none absolute inset-0 flex items-center justify-end pr-4 text-[6rem] leading-none opacity-15 text-ds-warning motion-safe:animate-suit-watermark"
                      >
                        {SUIT_SYMBOLS[state.chosenSuit] ?? '?'}
                      </span>
                    )}
                    {/* Wrap in a positioned div so DOM order — not "positioned beats static" —
                        decides stacking: the card must paint on top of the absolute watermark span above. */}
                    <div className="relative">
                      <AnimatedCard card={state.discardTop} width={cardWidth} />
                    </div>
                    <div className="text-ds-text-muted text-sm relative">
                      <div>{t('discardTop')}</div>
                      {state.chosenSuit > 0 && (
                        <div className="text-ds-warning">
                          {t('chosenSuit')}: {SUIT_SYMBOLS[state.chosenSuit] ?? '?'}
                        </div>
                      )}
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

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => {
                        // **勝者は「自分かどうか」とは別の軸。** 行の色分けは
                        // isHuman だけで、誰が勝ったかは表に出ていなかった
                        // (winnerIdx は届いていて文章にだけ反映されていた)。CUI は
                        // 勝者名を緑のバナーで出しており、そちらのほうが明確だった (#5499)。
                        // 決着前は出さない -- 途中経過の首位を勝者と読ませない。
                        const isWinner = isGameEnd && p.id === state.winnerIdx;
                        return (
                          <tr
                            key={p.id}
                            data-testid={`ce-score-row-${p.id}`}
                            data-winner={isWinner ? 'true' : undefined}
                            className={`${p.isHuman ? 'text-ds-accent' : ''}${
                              isWinner ? ' bg-ds-success/15 font-bold' : ''
                            }`}
                          >
                            <td>
                              {playerName(p.id, p.isHuman)}
                              {/* 色だけでは伝わらないので、勝者はテキストでも示す。 */}
                              {isWinner && (
                                <span className="ml-1 text-ds-success text-xs font-bold">{t('winnerBadge')}</span>
                              )}
                            </td>
                            <td className="text-center">{p.roundScore}</td>
                            <td className="text-center">{p.cumulativeScore}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          <GameFooter className={`${gameTheme.crazyeights.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 && (
              <fieldset className="flex flex-wrap justify-center gap-1.5 mb-2 border-0 p-0 m-0">
                <legend className="sr-only">{t('sort.label')}</legend>
                {CRAZYEIGHTS_SORT_MODES.map(({ mode, labelKey }) => (
                  <button
                    key={mode}
                    type="button"
                    onClick={() => handleSortMode(mode)}
                    className={sortMode === mode ? `${btnPrimary} min-w-[64px]` : `${btnSecondary} min-w-[64px]`}
                    aria-pressed={sortMode === mode}
                    data-testid={`ce-sort-${mode}`}
                  >
                    {t(labelKey)}
                  </button>
                ))}
              </fieldset>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ce-player-hand">
                {/* Shared screen-reader reason, referenced by every illegal card via
                    aria-describedby so the "why" is spoken (title alone is skipped by SRs). */}
                <span id={ILLEGAL_REASON_ID} className="sr-only">
                  {t('illegalHint')}
                </span>
                {sortedCrazyEightsHand(humanPlayer.cards, sortMode).map(({ card, index: idx }) => {
                  // On the human's turn, highlight playable cards (matching suit/rank or an 8)
                  // and dim the rest with a reason tooltip, so the rule is visible at a glance.
                  const legal = !isHumanTurn || isCrazyEightsLegalPlay(card, state.discardTop, state.chosenSuit);
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      title={isHumanTurn && !legal ? t('illegalHint') : undefined}
                      aria-describedby={isHumanTurn && !legal ? ILLEGAL_REASON_ID : undefined}
                      data-legal={isHumanTurn ? legal : undefined}
                      className={`transition-transform ${focusRingCard} ${
                        isHumanTurn && legal ? 'rounded-lg ring-2 ring-ds-success' : ''
                      } ${isHumanTurn && !legal ? 'opacity-50' : ''}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        ...selectedCardStyle(selectedCardIndices.includes(idx)),
                        boxSizing: 'border-box',
                      }}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
                <div className="flex gap-2" data-tutorial="ce-play-draw">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                    {t('drawButton')}
                  </button>
                  {/* **Hearts / Spades はサーバー計算の理由付きヒントを返すのに、
                      CrazyEights には無く、全ゲーム共通の簡易ヒューリスティック
                      しか支援が無かった (#4737)。** */}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleHint}
                    disabled={loading}
                    data-testid="ce-hint-button"
                  >
                    {tc('button.hint')}
                  </button>
                </div>
              )}
              {serverHint && (
                <p className="mt-2 text-sm text-ds-accent" data-testid="ce-server-hint">
                  {serverHint.suit !== undefined
                    ? t('hintSuit', { suit: SUIT_SYMBOLS[serverHint.suit] ?? '?' })
                    : t('hintCard', { idx: serverHint.cardIndex })}{' '}
                  ({t(`hintReason.${serverHint.reason}`)})
                </p>
              )}
              <ErrorAlert message={hintError} onRetry={undefined} />

              {isChooseSuit && (
                <div className="flex gap-1" data-tutorial="ce-suit-choice">
                  {SUIT_BUTTONS.map(({ suit, key }) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleChooseSuit(suit)}
                      disabled={loading}
                    >
                      {t(key)}
                    </button>
                  ))}
                </div>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ce-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="crazy-eights-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
