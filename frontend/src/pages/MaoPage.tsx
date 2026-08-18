import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { maoApi } from '../api/gameApi';
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
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useMaoGame } from '../hooks/useMaoGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, MaoResponse } from '../types/card';
import { CrazyEightsSuit, MaoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MAO_HELP, parseMaoCommand } from '../utils/cli/commands/maoCommands';
import { formatMaoState } from '../utils/cli/formatters/maoFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { ruleHintText } from '../utils/maoRuleHint';
import { appendSayWordAttempt, type MaoSayWordAttempt } from '../utils/maoSayWordHistory';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const MAO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MaoPhase.PLAY]: 'play',
  [MaoPhase.CHOOSE_SUIT]: 'chooseSuit',
  [MaoPhase.MUST_DECLARE]: 'mustDeclare',
  [MaoPhase.ROUND_END]: 'roundEnd',
  [MaoPhase.GAME_END]: 'gameEnd',
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

/** Red suits (hearts, diamonds) — rendered in the error/red token; spades and clubs use the ivory primary token. */
const RED_SUITS: ReadonlySet<number> = new Set([CrazyEightsSuit.HEART, CrazyEightsSuit.DIAMOND]);

/** Number of correct compliances required before a rule hint is unlocked. */
const HINT_THRESHOLD = 3;

/** Mao tutorial step definitions. */
const MAO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mao-discard-pile"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mao-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mao-play-draw"]', messageKey: 'tutorial.playDraw', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="mao-magic"]',
    messageKey: 'tutorial.magicCards',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mao-rule-panel"]',
    messageKey: 'tutorial.hiddenRule',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mao-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Mao game page with magic-card play, suit selection, declarations, and the hidden-rule panel. */
export const MaoPage = withTutorial(MaoPageContent, 'mao', MAO_TUTORIAL_STEPS);
/** Inner content of the Mao page, wrapped by TutorialProvider. */
function MaoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mao');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    maoConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleChooseSuit,
    handleDeclare,
    handleSkipDeclare,
    handleNextRound,
    handleDeclareWord,
  } = useMaoGame();
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const [wordInput, setWordInput] = useState('');
  // Local log of say-word attempts and their outcome; the server never returns
  // this history, but the player types the word and the response's rulePenalty
  // flag reveals whether it broke the hidden rule.
  const [sayWordHistory, setSayWordHistory] = useState<MaoSayWordAttempt[]>([]);
  // **最新は末尾** (appendSayWordAttempt は後ろに足す)。先頭を読むと、2 回目以降は
  // 1 回目の結果を読み上げ続ける。
  const latestSayWord = sayWordHistory[sayWordHistory.length - 1] ?? null;
  // Holds a submitted word until its response arrives; prevState pins the state
  // object present at submit time so the outcome is read from the *response*,
  // not from an interim re-render.
  const pendingWordRef = useRef<{ word: string; board: Card | null; prevState: MaoResponse | null } | null>(null);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('mao');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('mao', state);
  const cliConfig: CliGameConfig<MaoResponse, Parameters<typeof maoApi.exec>> = useMemo(
    () => ({
      gameName: 'mao',
      parseCommand: parseMaoCommand,
      formatResponse: formatMaoState,
      helpText: MAO_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === MaoPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players?.[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players?.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  const submitWord = useCallback(() => {
    const trimmed = wordInput.trim();
    if (trimmed.length === 0) return;
    pendingWordRef.current = { word: trimmed, board: state?.discardTop ?? null, prevState: state };
    handleDeclareWord(trimmed);
    setWordInput('');
  }, [handleDeclareWord, wordInput, state]);

  // When the say-word response lands (a new state object), record the attempt
  // together with the outcome the server reported for it.
  useEffect(() => {
    const pending = pendingWordRef.current;
    if (!state || !pending || state === pending.prevState) return;
    pendingWordRef.current = null;
    setSayWordHistory((h) =>
      appendSayWordAttempt(h, { word: pending.word, board: pending.board, penalty: state.rulePenalty }),
    );
  }, [state]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    setWordInput('');
    setSayWordHistory([]);
    pendingWordRef.current = null;
    void gameExec('reset', undefined, undefined, {
      cpuDifficulty: maoConfig.cpuDifficulty,
      pointLimit: maoConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, maoConfig.cpuDifficulty, maoConfig.pointLimit]);

  const handleNextRoundWithHistory = useCallback(() => {
    setSayWordHistory([]);
    pendingWordRef.current = null;
    handleNextRound();
  }, [handleNextRound]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('mao', MAO_PHASE_KEYS);

  // Buzz once on the false→true edge; fast CPU turns would otherwise swallow the feedback.
  const prevRulePenaltyRef = useRef(false);
  useEffect(() => {
    const penalty = state?.rulePenalty ?? false;
    if (penalty && !prevRulePenaltyRef.current) {
      playSound('errorBuzz');
    }
    prevRulePenaltyRef.current = penalty;
  }, [state?.rulePenalty, playSound]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="mao"
        layout={{ kind: 'trick-taking', centerCard: true, trickArea: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === MaoPhase.PLAY;
  const isChooseSuit = state.phase === MaoPhase.CHOOSE_SUIT;
  const isMustDeclare = state.phase === MaoPhase.MUST_DECLARE;
  const isRoundEnd = state.phase === MaoPhase.ROUND_END;
  const isGameEnd = state.phase === MaoPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const hasPenalty = state.penaltyDrawCount > 0;
  const directionLabel = state.direction < 0 ? '←' : '→';

  return (
    <GamePageShell
      title={tc('nav.mao')}
      gameThemeBg={gameTheme.mao.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isChooseSuit || (isMustDeclare && state.players[state.currentPlayerIdx]?.isHuman)}
      gamePath="/mao"
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
                    value: maoConfig.cpuDifficulty,
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
                    value: maoConfig.pointLimit,
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
              <span className="mr-4">{t('drawPile', { count: state.drawPileCount })}</span>
              <span>{t('direction', { dir: directionLabel })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top */}
                {state.discardTop && (
                  <div
                    className="my-3 p-3 rounded bg-black/40 flex items-center gap-3 relative overflow-hidden"
                    data-tutorial="mao-discard-pile"
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

                {hasPenalty && (
                  <div className={`my-2 p-2 rounded text-sm font-semibold ${badgeWarningColors}`} role="status">
                    {t('penaltyBanner', { count: state.penaltyDrawCount })}
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
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          <GameFooter className={`${gameTheme.mao.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="mao-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`transition-transform ${focusRingCard}`}
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
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* Hidden-rule panel: the player must guess when to speak — the game never reveals the rule. */}
            <div
              className={`my-2 p-2 rounded bg-black/30 flex flex-col gap-2 text-sm ${
                state.rulePenalty ? 'ring-2 ring-ds-error motion-safe:animate-pulse' : ''
              }`}
              data-tutorial="mao-rule-panel"
              data-testid="mao-rule-panel"
            >
              <div className="flex items-center gap-3 flex-wrap">
                <span className="text-ds-text-muted">
                  {t('compliance', { count: state.correctCount, total: HINT_THRESHOLD })}
                </span>
                {state.rulePenalty && (
                  <span className="text-ds-error font-semibold" role="status" data-testid="rule-penalty">
                    {t('rulePenalty')}
                  </span>
                )}
                {state.awaitingWord && (
                  <span className="text-ds-warning" role="status" data-testid="awaiting-word">
                    {t('awaitingWord')}
                  </span>
                )}
              </div>
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
              {state.hintUnlocked && state.ruleHint && (
                <div className="text-ds-accent" data-testid="rule-hint">
                  {t('ruleHintLabel')}: {ruleHintText(state, t)}
                </div>
              )}
              <div className="flex gap-2 items-center">
                <input
                  type="text"
                  value={wordInput}
                  onChange={(e) => setWordInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') submitWord();
                  }}
                  placeholder={t('sayWordPlaceholder')}
                  aria-label={t('sayWordPlaceholder')}
                  className="flex-1 min-w-0 rounded bg-black/40 px-2 py-1 text-ds-text-primary border border-white/20"
                  disabled={loading}
                />
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={submitWord}
                  disabled={loading || wordInput.trim().length === 0}
                >
                  {t('sayWordButton')}
                </button>
              </div>

              {/* Attempt log: track which words were tried and how the hidden rule reacted. */}
              {/* **違反はブザーと rule-penalty で強く伝わるのに、成功は履歴に静かに
                  積まれるだけだった** (#5668)。隠しルールを試行錯誤で learn する
                  ゲームなので、耳に届く情報が違反だけでは材料が片側しかない。
                  最新の 1 件を、違反と同じ polite で読み上げる。 */}
              {latestSayWord && (
                <span className="sr-only" role="status" aria-live="polite" data-testid="sayword-live">
                  {t('sayWordAnnounce', {
                    word: latestSayWord.word,
                    outcome: latestSayWord.penalty ? t('sayWordHistory.penalty') : t('sayWordHistory.correct'),
                  })}
                </span>
              )}
              {sayWordHistory.length > 0 && (
                <details className="rounded bg-black/20 px-2 py-1" data-testid="mao-sayword-history">
                  <summary className="cursor-pointer select-none text-ds-text-muted">
                    {t('sayWordHistory.title', { count: sayWordHistory.length })}
                  </summary>
                  <ul className="mt-1 flex flex-col gap-1">
                    {sayWordHistory.map((attempt, idx) => (
                      <li
                        key={`${attempt.word}-${idx}`}
                        className="flex items-center gap-2 flex-wrap"
                        data-testid="mao-sayword-history-row"
                      >
                        <span className="text-ds-text-primary font-medium">“{attempt.word}”</span>
                        {attempt.board && <span className="text-ds-text-muted">{cardAlt(attempt.board)}</span>}
                        <span
                          className={`font-semibold ${attempt.penalty ? 'text-ds-error' : 'text-ds-success'}`}
                          data-testid={attempt.penalty ? 'sayword-outcome-penalty' : 'sayword-outcome-correct'}
                        >
                          {attempt.penalty ? t('sayWordHistory.penalty') : t('sayWordHistory.correct')}
                        </span>
                      </li>
                    ))}
                  </ul>
                </details>
              )}
            </div>

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="mao-magic">
              {isHumanTurn && (
                <div className="flex gap-2" data-tutorial="mao-play-draw">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  <button
                    type="button"
                    className={`${hasPenalty ? btnDanger : btnPrimary} relative`}
                    onClick={handleDraw}
                    disabled={loading}
                  >
                    {hasPenalty ? t('takePenaltyButton', { count: state.penaltyDrawCount }) : t('drawButton')}
                    {hasPenalty && (
                      <span
                        data-testid="penalty-badge"
                        aria-hidden="true"
                        className="pointer-events-none absolute -top-2 -right-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-ds-error px-1 font-bold text-white text-xs"
                      >
                        {state.penaltyDrawCount}
                      </span>
                    )}
                  </button>
                </div>
              )}
              {isChooseSuit && state.players[state.currentPlayerIdx]?.isHuman && (
                <div className="flex gap-1" data-tutorial="mao-suit-choice">
                  {SUIT_BUTTONS.map(({ suit, key }) => (
                    <button
                      key={suit}
                      type="button"
                      className={`${btnSecondary} inline-flex items-center gap-1.5`}
                      onClick={() => handleChooseSuit(suit)}
                      disabled={loading}
                      aria-label={t(key)}
                    >
                      <span
                        aria-hidden="true"
                        data-testid={`suit-symbol-${suit}`}
                        className={`text-lg leading-none ${
                          RED_SUITS.has(suit) ? 'text-ds-error' : 'text-ds-text-primary'
                        }`}
                      >
                        {SUIT_SYMBOLS[suit]}
                      </span>
                      {t(key)}
                    </button>
                  ))}
                </div>
              )}
              {isMustDeclare && state.players[state.currentPlayerIdx]?.isHuman && (
                <div className="flex gap-2">
                  <button type="button" className={btnSuccess} onClick={handleDeclare} disabled={loading}>
                    {t('declareButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleSkipDeclare} disabled={loading}>
                    {t('skipDeclareButton')}
                  </button>
                </div>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRoundWithHistory} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="mao-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="mao-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
