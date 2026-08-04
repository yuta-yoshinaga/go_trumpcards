import { useCallback, useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { useCassinoGame } from '../hooks/useCassinoGame';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { gameTheme } from '../styles/gameTheme';
import type { CassinoResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { cassinoTakeCandidates } from '../utils/cassinoTakeCandidates';
import { suggestCassinoAction } from '../utils/cassinoUtils';
import {
  CASSINO_HELP,
  type CassinoCliArgs,
  formatCassinoState,
  parseCassinoCommand,
} from '../utils/cli/commands/cassinoCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const BUILD_VALUE_OPTIONS = Array.from({ length: 9 }, (_, i) => ({
  value: String(i + 2),
  label: String(i + 2),
}));

/** Tutorial steps for Cassino. */
const CS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cs-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="cs-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cs-builds"]',
    messageKey: 'tutorial.builds',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cs-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cs-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cs-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Cassino (カッシーノ) game page. */
export const CassinoPage = withTutorial(CassinoPageContent, 'cassino', CS_TUTORIAL_STEPS);
function CassinoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cassino');
  const {
    state,
    loading,
    error,
    callApi,
    handIndex,
    setHandIndex,
    tableIndices,
    toggleTable,
    buildIndices,
    toggleBuild,
    declaredValue,
    setDeclaredValue,
    configInput,
    handleConfigChange,
    playTake,
    playBuild,
    playTrail,
    handleResetWithConfig,
    retry,
  } = useCassinoGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cassino', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cassino');
  const cliConfig: CliGameConfig<CassinoResponse, CassinoCliArgs> = useMemo(
    () => ({
      gameName: 'cassino',
      parseCommand: parseCassinoCommand,
      formatResponse: formatCassinoState,
      helpText: [...CASSINO_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  // Hooks below must run unconditionally — they're computed before the early-return
  // skeleton guard so the hook order stays stable on the first render when `state`
  // is still null.
  const human = state && state.players.length >= 4 ? state.players[0] : null;
  const isHumanTurn = !!state && state.currentTurn === 0 && !state.gameEndFlag;

  const suggestion = useMemo(
    () =>
      state && human && isHumanTurn && handIndex !== null
        ? suggestCassinoAction({
            handCard: human.cards[handIndex],
            hand: human.cards,
            handIndex,
            selectedTableCards: tableIndices.map((i) => state.tableCards[i]).filter(Boolean),
            selectedBuilds: buildIndices.map((i) => state.builds[i]).filter(Boolean),
          })
        : null,
    [state, human, isHumanTurn, handIndex, tableIndices, buildIndices],
  );

  const onSuggest = useCallback(() => {
    if (!suggestion || handIndex === null) return;
    if (suggestion.type === 'take') {
      playTake();
      return;
    }
    setDeclaredValue(suggestion.declaredValue);
    callApi('build', {
      handIndex,
      tableIndices: [...tableIndices].sort((a, b) => a - b),
      declaredValue: suggestion.declaredValue,
    });
  }, [suggestion, handIndex, playTake, setDeclaredValue, callApi, tableIndices]);

  if (!state || !human) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.cassino.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanWon = isGameEnd && state.roundWinners.includes(0);
  const takeCandidateIndices =
    handIndex !== null && isHumanTurn
      ? cassinoTakeCandidates(state.tableCards, human.cards[handIndex]?.value ?? 0).indices
      : new Set<number>();
  const canTake = isHumanTurn && handIndex !== null && (tableIndices.length > 0 || buildIndices.length > 0);
  const canBuild = isHumanTurn && handIndex !== null && tableIndices.length > 0;
  const canTrail = isHumanTurn && handIndex !== null;
  const phaseName = isGameEnd ? t('phase.end') : t(`phase.${state.phase}`, t('phase.play'));
  const suggestionLabel =
    suggestion?.type === 'take'
      ? t('button.suggestTake', { value: suggestion.value })
      : suggestion?.type === 'build'
        ? t('button.suggestBuild', { value: suggestion.declaredValue })
        : '';

  return (
    <GamePageShell
      title={tc('nav.cassino')}
      gameThemeBg={gameTheme.cassino.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/cassino"
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="cs-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} —{' '}
                      {t('label.cpuStats', { cards: p.cardCount, score: p.totalScore })}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 8) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="cs-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('label.tableCards')}</div>
              <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('label.tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => {
                    const isCandidate = takeCandidateIndices.has(i);
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => isHumanTurn && toggleTable(i)}
                        disabled={!isHumanTurn}
                        className={`rounded transition-all ${
                          tableIndices.includes(i)
                            ? 'ring-2 ring-ds-warning -translate-y-1'
                            : isCandidate
                              ? 'ring-2 ring-ds-success motion-safe:animate-pulse'
                              : ''
                        } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                        data-testid={`table-card-${i}`}
                        data-take-candidate={isCandidate || undefined}
                        aria-label={`${cardAlt(c)}${
                          tableIndices.includes(i)
                            ? ` ${t('label.selected')}`
                            : isCandidate
                              ? ` ${t('label.takeCandidate')}`
                              : ''
                        }`}
                      >
                        <AnimatedCard card={c} width={cardWidth * 0.9} />
                      </button>
                    );
                  })
                )}
              </div>
            </div>

            {/* Builds */}
            <div className="py-2 bg-black/10 rounded-lg" data-tutorial="cs-builds">
              <div className="text-center text-xs text-ds-text-muted mb-1">{t('label.builds')}</div>
              {state.builds.length === 0 ? (
                <div className="text-center text-ds-text-muted text-sm">—</div>
              ) : (
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.builds.map((b, i) => {
                    const kind = b.isMulti ? t('build.multi') : t('build.single');
                    const owner = b.ownerIdx === 0 ? tc('player.you') : tc('player.cpu', { id: b.ownerIdx });
                    const buildLabel = t('build.label', { value: b.value, kind, owner });
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => isHumanTurn && toggleBuild(i)}
                        disabled={!isHumanTurn}
                        className={`px-3 py-1 rounded border text-sm ${
                          buildIndices.includes(i) ? 'ring-2 ring-ds-info bg-ds-info/20' : 'border-white/20 bg-black/20'
                        } ${isHumanTurn ? 'cursor-pointer' : ''}`}
                        data-testid={`build-${i}`}
                        aria-label={`${buildLabel}${buildIndices.includes(i) ? ` ${t('label.selected')}` : ''}`}
                      >
                        {buildLabel}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="cs-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} —{' '}
                {t('label.humanStats', {
                  hand: human.cardCount,
                  captured: human.capturedCount,
                  sweep: human.sweepCount,
                  score: human.totalScore,
                })}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && setHandIndex(handIndex === i ? null : i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${i}`}
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            {suggestion && (
              <div className="flex justify-center" data-testid="cs-suggest-area">
                <button
                  type="button"
                  onClick={onSuggest}
                  disabled={loading}
                  className="px-5 py-2.5 rounded-full bg-ds-accent text-ds-text-on-accent font-semibold shadow-lg motion-safe:animate-pulse disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                  data-testid="cs-suggest-button"
                  aria-label={suggestionLabel}
                >
                  {suggestionLabel}
                </button>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {/*
             * Single hint surface: the registry-driven advisory hint (`getCassinoHint`,
             * mirroring the domain heuristic) and the selection-derived one-click
             * suggestion are mutually exclusive. When the player's current selection
             * already forms a concrete take/build, that actionable suggestion supersedes
             * the generic advisory so only ONE piece of guidance is ever shown and the
             * two can never contradict each other.
             */}
            {!suggestion && <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />}
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
                    options: DIFFICULTY_OPTIONS,
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'multiBuildEnabled',
                    label: t('settings.multiBuild'),
                    checked: configInput.multiBuildEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('multiBuildEnabled', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'sweepBonusEnabled',
                    label: t('settings.sweepBonus'),
                    checked: configInput.sweepBonusEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('sweepBonusEnabled', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.cassino.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="cs-actions">
              <button
                type="button"
                onClick={playTake}
                disabled={loading || !canTake}
                className="px-4 py-2 rounded-lg bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="take-button"
              >
                {t('button.take')}
              </button>
              <div className="flex items-center gap-1">
                <select
                  aria-label={t('label.buildValue')}
                  value={String(declaredValue)}
                  onChange={(e) => setDeclaredValue(Number.parseInt(e.target.value, 10))}
                  className="px-2 py-2 rounded bg-black/30 text-ds-text-primary text-sm"
                  data-testid="build-value-select"
                >
                  {BUILD_VALUE_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>
                      {o.label}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  onClick={playBuild}
                  disabled={loading || !canBuild}
                  className="px-4 py-2 rounded-lg bg-ds-success text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                  data-testid="build-button"
                >
                  {t('button.build')}
                </button>
              </div>
              <button
                type="button"
                onClick={playTrail}
                disabled={loading || !canTrail}
                className="px-4 py-2 rounded-lg bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="trail-button"
              >
                {t('button.trail')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cs-reset-button"
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
