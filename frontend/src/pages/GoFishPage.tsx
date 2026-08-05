import { useCallback, useMemo, useState } from 'react';
import type { goFishApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GoFishBooksDisplay } from '../components/gofish/GoFishBooksDisplay';
import { GoFishPlayerArea } from '../components/gofish/GoFishPlayerArea';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { type ActionBinding, useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGoFishGame } from '../hooks/useGoFishGame';
import { useGoFishKnownRanks } from '../hooks/useGoFishKnownRanks';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GoFishResponse } from '../types/card';
import { GoFishPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { GOFISH_HELP, parseGofishCommand } from '../utils/cli/commands/gofishCommands';
import { formatGofishState } from '../utils/cli/formatters/gofishFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** CPU difficulty options for Go Fish. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Go Fish tutorial step definitions. */
const GF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gf-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gf-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gf-ask-button"]',
    messageKey: 'tutorial.askButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gf-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GOFISH_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GoFishPhase.PLAY]: 'play',
  [GoFishPhase.GAME_END]: 'end',
};

/** Renders the Go Fish game page. */
export const GoFishPage = withTutorial(GoFishPageContent, 'gofish', GF_TUTORIAL_STEPS);
/** Inner content of the Go Fish page, wrapped by TutorialProvider. */
function GoFishPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('gofish');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    goFishConfig,
    handleConfigChange,
    selectedTarget,
    selectedRank,
    handleSelectTarget,
    handleSelectRank,
    handleAsk,
  } = useGoFishGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gofish', state);

  const phaseNames = usePhaseNames('gofish', GOFISH_PHASE_KEYS);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('gofish');
  const cliConfig: CliGameConfig<GoFishResponse, Parameters<typeof goFishApi.exec>> = useMemo(
    () => ({
      gameName: 'gofish',
      parseCommand: parseGofishCommand,
      formatResponse: formatGofishState,
      helpText: GOFISH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, { cpuDifficulty: goFishConfig.cpuDifficulty });
  }, [exec, hideActionLog, goFishConfig.cpuDifficulty]);

  const knownRanks = useGoFishKnownRanks(state);

  // Keyboard: number keys pick the opponent, arrows cycle the rank, Enter/a asks.
  // Selection changes are announced via an aria-live region (below), on top of
  // the per-button aria-pressed state.
  const [kbdAnnounce, setKbdAnnounce] = useState('');
  const kbdCpuPlayers = useMemo(() => state?.players.filter((p) => !p.isHuman) ?? [], [state?.players]);
  const kbdHumanRanks = useMemo(() => {
    const human = state?.players.find((p) => p.isHuman);
    return human ? [...new Set(human.cards.map((c) => c.value))].sort((a, b) => a - b) : [];
  }, [state?.players]);
  const kbdIsHumanTurn =
    state?.phase === GoFishPhase.PLAY && state.players[state.currentTurn]?.isHuman === true && !loading;
  const kbdCanAsk = kbdIsHumanTurn && selectedTarget !== null && selectedRank !== null;
  const cycleRank = useCallback(
    (dir: 1 | -1) => {
      if (kbdHumanRanks.length === 0) return;
      const cur = selectedRank === null ? -1 : kbdHumanRanks.indexOf(selectedRank);
      const next =
        cur === -1
          ? dir === 1
            ? 0
            : kbdHumanRanks.length - 1
          : (cur + dir + kbdHumanRanks.length) % kbdHumanRanks.length;
      const rank = kbdHumanRanks[next];
      handleSelectRank(rank);
      setKbdAnnounce(t('a11y.rankSelected', { rank: valueName(rank) }));
    },
    [kbdHumanRanks, selectedRank, handleSelectRank, t],
  );
  const askBindings = useMemo(() => {
    // Annotated so the pushes below can carry `label`; inference from the map
    // alone would fix the element type to exactly the properties it sets.
    const bindings: ActionBinding[] = kbdCpuPlayers.map((p, i) => ({
      key: String(i + 1),
      action: () => {
        handleSelectTarget(p.id);
        setKbdAnnounce(t('a11y.targetSelected', { name: playerName(p.id, false) }));
      },
      enabled: kbdIsHumanTurn,
      // 相手ごとに別のキーなので、一覧も相手ごとに別の文言にする。全部
      // 「対象のプレイヤーを選ぶ」だと、どのキーが誰なのか読み取れない (#4862)。
      label: 'selectTargetNamed',
      labelParams: { name: playerName(p.id, false) },
    }));
    bindings.push({ key: 'ArrowRight', action: () => cycleRank(1), enabled: kbdIsHumanTurn, label: 'nextRank' });
    bindings.push({ key: 'ArrowLeft', action: () => cycleRank(-1), enabled: kbdIsHumanTurn, label: 'prevRank' });
    // Ask key is the letter "a" only. Enter is deliberately not bound: the hook
    // listens at the document level and does not exclude BUTTON, so an Enter
    // binding would double-fire (native button activation + this handler) and
    // send a duplicate ask when a button is focused.
    bindings.push({ key: 'a', action: handleAsk, enabled: kbdCanAsk, label: 'ask' });
    return bindings;
  }, [kbdCpuPlayers, handleSelectTarget, cycleRank, handleAsk, kbdIsHumanTurn, kbdCanAsk, t]);
  useActionKeyboardNav({ bindings: askBindings, enabled: !!kbdIsHumanTurn });

  if (!state) return <GameSkeleton gameKey="gofish" layout={{ kind: 'trick-taking', footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);

  const isPlayPhase = state.phase === GoFishPhase.PLAY;
  const isGameEnd = state.phase === GoFishPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentTurn]?.isHuman === true;
  const canAsk = isHumanTurn && selectedTarget !== null && selectedRank !== null && !loading;

  // Get unique ranks from human hand for rank selection
  const humanRanks = humanPlayer ? [...new Set(humanPlayer.cards.map((c) => c.value))].sort((a, b) => a - b) : [];
  const humanRankSet = new Set(humanRanks);

  return (
    <GamePageShell
      title={tc('nav.gofish')}
      gameThemeBg={gameTheme.gofish.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/gofish"
      gameEndFlag={!!isGameEnd}
      winShow={!!state.gameEndFlag}
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
            title={t('setup.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('setup.cpuDifficulty'),
                    value: goFishConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`setup.${o.label}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Turn & deck info */}
            <div className="text-ds-text-primary text-center mb-2 flex items-center justify-center gap-4">
              <span>{t('deck', { count: state.deckRemaining })}</span>
            </div>

            <div className="sr-only" role="status" aria-live="polite" data-testid="gf-kbd-announce">
              {kbdAnnounce}
            </div>

            {/* CPU player areas */}
            <div data-tutorial="gf-cpu-area">
              {cpuPlayers.map((p) => {
                // Surface a transient bubble on whichever player was just
                // targeted by the most recent ask. `turnNumber` guarantees a
                // fresh triggerKey even when two consecutive asks have the
                // same rank and target. See issue #1490.
                //
                // `cardsReceived` is serialized by the Go backend with
                // `omitempty`, so on a miss the field is absent entirely —
                // using `?.length ?? 0` instead of `.length` avoids the
                // TypeError that otherwise crashes the page before CPU turns
                // finish.
                const lastAsk = state.lastAsk;
                const askAnnotation =
                  lastAsk && lastAsk.targetIdx === p.id
                    ? {
                        rank: lastAsk.rank,
                        receivedCount: lastAsk.cardsReceived?.length ?? 0,
                        triggerKey: `${state.turnNumber}-${lastAsk.playerIdx}-${lastAsk.targetIdx}-${lastAsk.rank}`,
                      }
                    : undefined;
                const playerKnown = knownRanks[p.id] ?? [];
                const matched = playerKnown.filter((r) => humanRankSet.has(r));
                return (
                  <GoFishPlayerArea
                    key={p.id}
                    player={p}
                    isSelected={selectedTarget === p.id}
                    onSelect={handleSelectTarget}
                    disabled={!isHumanTurn || loading}
                    askAnnotation={askAnnotation}
                    knownRanks={playerKnown}
                    matchedRanks={matched}
                  />
                );
              })}
            </div>

            {/* Rank selector */}
            {isHumanTurn && humanRanks.length > 0 && (
              <div className="my-3">
                <div className="text-ds-text-muted text-sm mb-1">{t('selectRank')}</div>
                <div className="flex flex-wrap gap-2">
                  {humanRanks.map((rank) => (
                    <button
                      key={rank}
                      type="button"
                      onClick={() => handleSelectRank(rank)}
                      className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                        selectedRank === rank
                          ? 'bg-ds-warning text-ds-text-on-accent'
                          : 'bg-white/10 text-ds-text-primary hover:bg-white/20'
                      }`}
                      aria-pressed={selectedRank === rank}
                    >
                      {valueName(rank)}
                    </button>
                  ))}
                </div>
              </div>
            )}

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Human books display */}
            {humanPlayer && <GoFishBooksDisplay books={humanPlayer.books} />}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.gofish.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="gf-player-hand">
                {humanPlayer.cards.map((card) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}`}
                    onClick={() => handleSelectRank(card.value)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedRank === card.value}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...selectedCardStyle(selectedRank === card.value),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex gap-2 items-center" data-tutorial="gf-ask-button">
              {isHumanTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleAsk} disabled={!canAsk}>
                    {t('button.ask')}
                  </button>
                  <span className="text-ds-text-muted text-xs hidden sm:inline">{t('a11y.kbdHint')}</span>
                </>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="gf-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={askBindings} data-testid="go-fish-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
