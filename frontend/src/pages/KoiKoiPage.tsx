import { useEffect, useMemo } from 'react';
import type { koikoiApi } from '../api/gameApi';
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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useKoiKoiGame } from '../hooks/useKoiKoiGame';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, KoiKoiResponse, KoiKoiYaku } from '../types/card';
import { KoiKoiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KOIKOI_HELP, parseKoiKoiCommand } from '../utils/cli/commands/koikoiCommands';
import { formatKoiKoiState } from '../utils/cli/formatters/koikoiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { groupCapturedByCategory, KOIKOI_CATEGORY_ORDER, type KoiKoiCategory } from '../utils/koikoiCategory';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Koi-Koi (こいこい) tutorial step definitions. */
const KOIKOI_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="koikoi-field"]', messageKey: 'tutorial.field', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="koikoi-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="koikoi-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="koikoi-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Props for {@link CapturedCategories}. */
interface CapturedCategoriesProps {
  /** The player's captured hanafuda cards. */
  captured: Card[];
  /** Rendered width of each captured card, in pixels. */
  cardWidth: number;
  /** Localizes a category key (e.g. `bright`) to its display label. */
  categoryLabel: (cat: KoiKoiCategory) => string;
  /** Text shown when the pile is empty. */
  emptyText: string;
  /** Prefix for each group's `data-testid` (e.g. `koikoi-cpu`). */
  testidPrefix: string;
}

/**
 * Renders a player's captured cards grouped by yaku category (光 / 種 / 短冊 / カス),
 * each group headed by its localized label and count. Empty categories are hidden.
 * The row scrolls horizontally so it stays within narrow mobile widths.
 */
function CapturedCategories({ captured, cardWidth, categoryLabel, emptyText, testidPrefix }: CapturedCategoriesProps) {
  if (captured.length === 0) {
    return <div className="text-xs text-ds-text-muted min-h-[24px]">{emptyText}</div>;
  }
  const groups = groupCapturedByCategory(captured);
  return (
    <div className="flex gap-4 overflow-x-auto justify-center px-2 min-h-[24px]">
      {KOIKOI_CATEGORY_ORDER.filter((cat) => groups[cat].length > 0).map((cat) => (
        <div key={cat} className="flex flex-col items-center shrink-0" data-testid={`${testidPrefix}-group-${cat}`}>
          <div className="text-[10px] text-ds-text-muted mb-0.5 whitespace-nowrap">
            {categoryLabel(cat)} · {groups[cat].length}
          </div>
          <div className="flex gap-0.5">
            {groups[cat].map((c, i) => (
              <CardImage key={i} card={c} width={cardWidth * 0.42} />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

/** Renders the Koi-Koi (こいこい) game page: a 2-player hanafuda capture game with yaku scoring. */
export const KoiKoiPage = withTutorial(KoiKoiPageContent, 'koikoi', KOIKOI_TUTORIAL_STEPS);

/** Inner content of the Koi-Koi page, wrapped by TutorialWrapper. */
function KoiKoiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('koikoi');
  const {
    state,
    loading,
    error,
    callApi,
    retry,
    handIndex,
    selectHand,
    configInput,
    handleConfigChange,
    playCard,
    callKoiKoi,
    callStop,
    handleNextRound,
    handleResetWithConfig,
  } = useKoiKoiGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('koikoi', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('koikoi');
  const cliConfig: CliGameConfig<KoiKoiResponse, Parameters<typeof koikoiApi.exec>> = useMemo(
    () => ({
      gameName: 'koikoi',
      parseCommand: parseKoiKoiCommand,
      formatResponse: formatKoiKoiState,
      helpText: KOIKOI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.koikoi.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const cpu = state.players.find((p) => !p.isHuman) ?? null;

  const isPlayPhase = state.phase === KoiKoiPhase.PLAY;
  const isDecisionPhase = state.phase === KoiKoiPhase.KOIKOI_DECISION;
  // **こいこいを一度でも宣言していれば確定点は 2 倍** (KoiKoi.endRound と同じ)。
  const decisionMultiplier = state.koikoiCount >= 1 ? 2 : 1;
  const isRoundEnd = state.phase === KoiKoiPhase.ROUND_END;
  const isGameEnd = state.phase === KoiKoiPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winner === (human?.id ?? 0);
  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDecisionPhase
        ? t('phase.decision')
        : t('phase.play');

  // Field indices the currently-selected hand card can capture (backend hint).
  const candidates = handIndex !== null && isPlayPhase && isHumanTurn ? (state.captureOptions[handIndex] ?? []) : [];
  const needsFieldPick = candidates.length > 1;
  const candidateSet = new Set(candidates);

  /** Localizes a yaku key, falling back to the raw key. */
  const yakuName = (key: string): string => t(`yaku.${key}`, { defaultValue: key });

  /** Localizes a captured-card category key (光 / 種 / 短冊 / カス). */
  const categoryLabel = (cat: KoiKoiCategory): string => t(`category.${cat}`);

  /** Renders a compact "name (points)" list of yaku. */
  const yakuList = (yaku: KoiKoiYaku[]): string => yaku.map((y) => `${yakuName(y.key)} (${y.points})`).join('  ·  ');

  /** Handles clicking a hand card: selects it, or plays immediately when no field choice is needed. */
  const onHandClick = (idx: number) => {
    if (!isPlayPhase || !isHumanTurn) return;
    const opts = state.captureOptions[idx] ?? [];
    if (opts.length > 1) {
      // Two-way match: require the player to pick a field card next.
      selectHand(idx);
    } else {
      // Zero or single match: play right away (backend resolves the capture).
      playCard(idx);
    }
  };

  /** Handles clicking a field card during a two-way-match pick. */
  const onFieldClick = (fieldIdx: number) => {
    if (!needsFieldPick || handIndex === null) return;
    if (!candidateSet.has(fieldIdx)) return;
    playCard(handIndex, fieldIdx);
  };

  const playerLine = (label: string, p: (typeof state.players)[number]) =>
    `${label} — ${t('captured', { count: p.capturedCount })} · ${t('score', { score: p.score })}${
      p.yaku.length > 0 ? ` · ${yakuList(p.yaku)}` : ''
    }`;

  const winnerName = state.winner < 0 ? '' : state.winner === (human?.id ?? 0) ? t('you') : t('cpu');

  return (
    <GamePageShell
      title={tc('nav.koikoi')}
      gameThemeBg={gameTheme.koikoi.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/koikoi"
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
            <div className="text-center text-xs text-ds-text-muted" data-tutorial="koikoi-info">
              <span className="mr-3">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-3">{t('deck', { count: state.remainingDeck })}</span>
              <span>{t('koikoiCount', { count: state.koikoiCount })}</span>
            </div>

            {/* CPU captured + yaku */}
            {cpu && (
              <div className="text-center" data-testid="koikoi-cpu">
                <div className="text-xs text-ds-text-muted mb-1">{playerLine(t('cpu'), cpu)}</div>
                <CapturedCategories
                  captured={cpu.captured}
                  cardWidth={cardWidth}
                  categoryLabel={categoryLabel}
                  emptyText={t('capturedEmpty')}
                  testidPrefix="koikoi-cpu"
                />
              </div>
            )}

            {/* Field cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="koikoi-field">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('field')}</div>
              <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                {state.fieldCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('fieldEmpty')}</span>
                ) : (
                  state.fieldCards.map((c, i) => {
                    const isCandidate = candidateSet.has(i);
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => onFieldClick(i)}
                        disabled={!needsFieldPick || !isCandidate}
                        className={`rounded transition-all ${
                          isCandidate ? 'ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                        } ${needsFieldPick && isCandidate ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                        data-testid={`field-card-${i}`}
                        data-capture-candidate={isCandidate || undefined}
                      >
                        <CardImage card={c} width={cardWidth * 0.9} />
                      </button>
                    );
                  })
                )}
              </div>
              {needsFieldPick && (
                <div className="text-center text-sm text-ds-accent mt-2" data-testid="koikoi-field-pick">
                  {t('pickField')}
                </div>
              )}
            </div>

            {/* Human captured, grouped by yaku category */}
            {human && (
              <div className="text-center" data-testid="koikoi-human-captured">
                <CapturedCategories
                  captured={human.captured}
                  cardWidth={cardWidth}
                  categoryLabel={categoryLabel}
                  emptyText={t('capturedEmpty')}
                  testidPrefix="koikoi-human"
                />
              </div>
            )}

            {/* Human hand */}
            <div className="text-center" data-tutorial="koikoi-hand">
              <div className="text-xs text-ds-text-muted mb-1">{human ? playerLine(t('you'), human) : ''}</div>
              <div className="flex flex-wrap justify-center gap-2">
                {human?.cards.map((c, i) => {
                  const playable = state.playableIndices.includes(i);
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => onHandClick(i)}
                      disabled={!isPlayPhase || !isHumanTurn}
                      className={`rounded transition-all ${
                        handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                      } ${isPlayPhase && isHumanTurn && playable ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                      data-testid={`hand-card-${i}`}
                      data-playable={playable || undefined}
                    >
                      <CardImage card={c} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Turn prompt */}
            {isPlayPhase && (
              <div className="text-center text-sm text-ds-accent" data-testid="koikoi-prompt">
                {isHumanTurn ? t('turnYours') : t('turnCpu')}
              </div>
            )}

            {/* Koi-Koi / Shobu decision */}
            {isDecisionPhase && !isGameEnd && (
              <div
                className="my-2 p-3 rounded-lg bg-black/40 text-center border border-ds-warning/60"
                data-testid="koikoi-decision"
              >
                <div className="text-ds-text-primary font-semibold mb-1">{t('decision.title')}</div>
                <div className="text-ds-warning text-sm mb-2" data-testid="koikoi-decision-points">
                  {/* **こいこい 1 回以降は倍。**生の pendingPoints を出すと、
                      実際より低い「今止めた場合の点数」を見せることになる
                      (#4929)。CUI は koikoiDecisionInfoStr で倍率を掛けている。 */}
                  {yakuList(state.pendingYaku)} ={' '}
                  {t('decision.points', { points: state.pendingPoints * decisionMultiplier })}
                  {decisionMultiplier > 1 && ` ${t('decision.multiplier', { mult: decisionMultiplier })}`}
                </div>
                <div className="flex gap-3 justify-center">
                  <button type="button" className={btnWarning} onClick={callKoiKoi} disabled={loading}>
                    {t('decision.koikoi')}
                  </button>
                  <button type="button" className={btnSuccess} onClick={callStop} disabled={loading}>
                    {t('decision.shobu')}
                  </button>
                </div>
              </div>
            )}

            {/* Round-end result */}
            {isRoundEnd && state.lastRoundResult && (
              <div
                className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                data-testid="koikoi-round-result"
              >
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div className="text-ds-success mb-1">
                  {state.lastRoundResult.winner < 0
                    ? t('roundResult.draw')
                    : t('roundResult.winner', {
                        name: state.lastRoundResult.winner === (human?.id ?? 0) ? t('you') : t('cpu'),
                      })}
                </div>
                {state.lastRoundResult.yaku.length > 0 && <div>{yakuList(state.lastRoundResult.yaku)}</div>}
                <div>
                  {t('roundResult.total', {
                    base: state.lastRoundResult.basePoints,
                    mult: state.lastRoundResult.multiplier,
                    total: state.lastRoundResult.total,
                  })}
                </div>
              </div>
            )}

            {/* Game-end result */}
            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="koikoi-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                {state.winner >= 0 && (
                  <div className="text-ds-success mb-1">{t('result.winner', { name: winnerName })}</div>
                )}
                {state.players.map((p) => (
                  <div key={p.id}>{t('result.score', { name: p.isHuman ? t('you') : t('cpu'), score: p.score })}</div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(configInput.cpuDifficulty ?? 1),
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: String(o.value),
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'select' as const,
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: String(configInput.targetScore ?? 50),
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                    onSelect: (v: string) => handleConfigChange('targetScore', Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.koikoi.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="koikoi-actions">
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnPrimary} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleResetWithConfig} disabled={loading}>
                  {t('newGame')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleResetWithConfig}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="koikoi-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
