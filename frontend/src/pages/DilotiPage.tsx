import { useEffect, useMemo } from 'react';
import type { dilotiApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { DILOTI_CPU_DIFFICULTY_OPTIONS, DILOTI_TARGET_OPTIONS, useDilotiGame } from '../hooks/useDilotiGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DilotiResponse } from '../types/card';
import { DilotiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DILOTI_HELP, parseDilotiCommand } from '../utils/cli/commands/dilotiCommands';
import { formatDilotiState } from '../utils/cli/formatters/dilotiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Diloti tutorial step definitions. */
const DILOTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="diloti-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="diloti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
const DILOTI_PHASE_KEYS: Readonly<Record<string, string>> = {
  [DilotiPhase.PLAY]: 'play',
  [DilotiPhase.ROUND_END]: 'roundEnd',
  [DilotiPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Diloti page: the Greek fishing game where captures sum to the
 * played card's own rank, declarations build piles, and a one-card sweep of
 * the table is a xeri worth ten.
 */
export const DilotiPage = withTutorial(DilotiPageContent, 'diloti', DILOTI_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function DilotiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('diloti');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    dilotiConfig,
    handleConfigChange,
    selectedHandIdx,
    selectHand,
    take,
    declare,
    trail,
    handleNextRound,
    reset,
  } = useDilotiGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('diloti');
  const cliConfig: CliGameConfig<DilotiResponse, Parameters<typeof dilotiApi.exec>> = useMemo(
    () => ({
      gameName: 'diloti',
      parseCommand: parseDilotiCommand,
      formatResponse: formatDilotiState,
      helpText: DILOTI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('diloti', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return (
      <GameSkeleton
        gameKey="diloti"
        layout={{ kind: 'casino-table', sections: [4], footerStyle: 'hand', footerHandSize: 6 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === DilotiPhase.PLAY;
  const isRoundEnd = state.phase === DilotiPhase.ROUND_END;
  const isGameEnd = state.phase === DilotiPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // **打てる手はサーバが数えたものだけ。** 同ランク・合計一致・宣言が絡むので、
  // 画面側で組み直すと必ずずれる。
  const takes = selectedHandIdx !== null ? (state.takeOptions[selectedHandIdx] ?? []) : [];
  const declareCands = selectedHandIdx !== null ? (state.declareOptions[selectedHandIdx] ?? []) : [];
  const mayTrail = selectedHandIdx !== null && (state.canTrail[selectedHandIdx] ?? false);
  const handValidIndices = canPlay ? humanPlayer?.cards.map((_, i) => i) : undefined;

  /** Renders a capture target list as "1, 2, decl 0". */
  const takeLabel = (tableIdxs: number[], declIdxs: number[]) =>
    [...tableIdxs.map(String), ...declIdxs.map((d) => t('declShort', { idx: d }))].join(', ');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.diloti')}
      gameThemeBg={gameTheme.diloti.bg}
      phaseName={t(`phase.${DILOTI_PHASE_KEYS[state.phase] ?? 'play'}`)}
      isHumanTurn={state.isHumanTurn}
      gamePath="/diloti"
      gameEndFlag={isGameEnd}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: dilotiConfig.cpuDifficulty,
                    options: DILOTI_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: dilotiConfig.targetScore,
                    options: DILOTI_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, target: state.config.targetScore })}</span>
              <span data-testid="diloti-deck">{t('deck', { n: state.deckRemaining })}</span>
            </div>

            <div className={lgTwoColGrid}>
              <div data-tutorial="diloti-table">
                <div className="mb-1 text-ds-text-muted text-sm">{t('tableLabel')}</div>
                <div className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1" data-testid="diloti-table">
                  {state.table.length === 0 ? (
                    <span className="text-ds-text-muted text-sm">{t('tableEmpty')}</span>
                  ) : (
                    state.table.map((c, i) => (
                      <div key={`${c.design}-${c.value}-${i}`} className="flex flex-col items-center">
                        <CardImage card={c} width={cardWidth} />
                        <span className="text-xs text-ds-text-muted">{i}</span>
                      </div>
                    ))
                  )}
                </div>

                {/* **宣言も番号付きで見せる。** 見えないと `d0` で取る対象を
                    指せず、グループ宣言かどうかも分からない。 */}
                {state.declarations.length > 0 && (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="diloti-declarations">
                    <div className="mb-1 text-ds-text-muted text-sm">{t('declarationsLabel')}</div>
                    {state.declarations.map((d, i) => (
                      <div
                        key={`${d.ownerIdx}-${d.value}-${i}`}
                        className="text-ds-text-primary text-sm py-0.5"
                        data-testid={`diloti-declaration-${i}`}
                      >
                        {t('declarationLine', {
                          idx: i,
                          value: d.value,
                          kind: d.isGroup ? t('declGroup') : t('declPlain'),
                          owner: playerName(d.ownerIdx, d.ownerIdx === 0),
                        })}
                        <div className="flex flex-wrap gap-1 mt-0.5">
                          {d.groups.flat().map((c, j) => (
                            <CardImage key={`${c.design}-${c.value}-${j}`} card={c} width={cardWidth} />
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* **打てる手はサーバが数えたものだけ。** */}
                {canPlay && selectedHandIdx !== null && (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="diloti-move-options">
                    <div className="text-ds-text-primary text-sm mb-1">
                      {takes.length > 0 || declareCands.length > 0 ? t('chooseMove') : t('noCapture')}
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {takes.map((o) => (
                        <button
                          key={`t-${o.tableIdxs.join('-')}-${o.declIdxs.join('-')}`}
                          type="button"
                          className={btnPrimary}
                          onClick={() => take(o.tableIdxs, o.declIdxs)}
                          disabled={loading}
                          data-testid={`diloti-take-${[...o.tableIdxs, ...o.declIdxs.map((d) => `d${d}`)].join('-')}`}
                        >
                          {t('takeGroup', { cards: takeLabel(o.tableIdxs, o.declIdxs) })}
                        </button>
                      ))}
                      {declareCands.map((c) => (
                        <button
                          key={`d-${c.value}-${c.tableIdxs.join('-')}`}
                          type="button"
                          className={btnSecondary}
                          onClick={() => declare(c.value, c.tableIdxs)}
                          disabled={loading}
                          data-testid={`diloti-declare-${c.value}-${c.tableIdxs.join('-')}`}
                        >
                          {t('declareGroup', { value: c.value, cards: c.tableIdxs.join(', ') })}
                        </button>
                      ))}
                      {mayTrail && (
                        <button
                          type="button"
                          className={btnSuccess}
                          onClick={trail}
                          disabled={loading}
                          data-testid="diloti-lay-off"
                        >
                          {t('layOffButton')}
                        </button>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div data-tutorial="diloti-scores">
                <div className="mb-2 p-2 rounded bg-black/30" data-testid="diloti-scores">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      <span className="flex items-center gap-2">
                        <span>
                          {playerName(p.id, p.isHuman)}: {t('score', { n: p.score })} /{' '}
                          {t('taken', { n: p.capturedCount })} / {t('xeri', { n: p.xeri })}
                        </span>
                        {p.isDealer && (
                          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                            {t('dealerBadge')}
                          </span>
                        )}
                      </span>
                    </div>
                  ))}
                </div>

                {state.lastResult && (isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="diloti-round-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult')}</div>
                    {state.lastResult.lines
                      .filter((l) => l.points[0] !== 0 || l.points[1] !== 0)
                      .map((l) => (
                        <div key={l.key}>
                          {t(`score.${l.key}`)}: {l.points[0]} - {l.points[1]}
                        </div>
                      ))}
                    <div className="text-ds-text-primary">
                      {t('roundTotal', { a: state.lastResult.totals[0], b: state.lastResult.totals[1] })}
                    </div>
                  </div>
                )}

                {isGameEnd && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="diloti-winner">
                    {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
                  </div>
                )}
              </div>
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

          <GameFooter className={`${gameTheme.diloti.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedHandIdx === null ? [] : [selectedHandIdx]}
                toggleCard={selectHand}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="diloti"
                validIndices={handValidIndices}
                legalIndices={handValidIndices}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="diloti-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && state.hintHandIdx >= 0 && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hintReason}`)} ([{state.hintHandIdx}]
                  {` → ${t(`action.${state.hintAction}`)}`})
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="diloti-action-buttons">
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextRound}
                  disabled={loading}
                  data-testid="diloti-next-round"
                >
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="diloti-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
