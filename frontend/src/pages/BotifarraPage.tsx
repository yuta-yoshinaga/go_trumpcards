import { useCallback, useMemo } from 'react';
import { botifarraApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BotifarraResponse } from '../types/card';
import { BOTIFARRA_NO_TRUMP, BOTIFARRA_TOTAL_POINTS } from '../types/games/botifarra';
import { BotifarraPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BOTIFARRA_CLI_HELP, parseBotifarraCommand } from '../utils/cli/commands/botifarraCommands';
import { formatBotifarraState } from '../utils/cli/formatters/botifarraFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const BF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bf-declare"]',
    messageKey: 'tutorial.declare',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="bf-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bf-score"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
];

/** The four trump choices plus no-trump, in declaration order. */
const TRUMP_CHOICES = [
  { suit: 1, key: 'trump.spade' },
  { suit: 2, key: 'trump.clover' },
  { suit: 3, key: 'trump.heart' },
  { suit: 4, key: 'trump.diamond' },
  { suit: BOTIFARRA_NO_TRUMP, key: 'trump.none' },
] as const;

/** Renders the Botifarra game page (#5229). */
export const BotifarraPage = withTutorial(BotifarraPageContent, 'botifarra', BF_TUTORIAL_STEPS);

function BotifarraPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('botifarra');

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(botifarraApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('botifarra');
  const cliConfig: CliGameConfig<BotifarraResponse, Parameters<typeof botifarraApi.exec>> = useMemo(
    () => ({
      gameName: 'botifarra',
      parseCommand: parseBotifarraCommand,
      formatResponse: formatBotifarraState,
      helpText: BOTIFARRA_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isDeclarePhase = phase === BotifarraPhase.DECLARE || phase === BotifarraPhase.DELEGATED;
  const isDoublePhase = phase === BotifarraPhase.DOUBLE;
  const isPlayPhase = phase === BotifarraPhase.PLAY;
  const isRoundEnd = phase === BotifarraPhase.ROUND_END;

  // **切り札は -1 も有効な値。** 値を必ず送ります。
  const handleDeclare = useCallback((suit: number) => execApi('declare', undefined, suit), [execApi]);
  const handlePlay = useCallback((cardIndex: number) => execApi('play', cardIndex), [execApi]);

  const actionBindings = useMemo(
    () => [
      { key: 'n', action: () => execApi('next'), enabled: isRoundEnd },
      { key: 'g', action: () => execApi('giveup'), enabled: !!state && !state.gameEndFlag },
    ],
    [execApi, isRoundEnd, state],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** 下に置くと初回レンダーだけフック数が変わり、
  // ページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('botifarra', state);

  if (!state) {
    return <GameSkeleton gameKey="botifarra" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const legal = new Set(state.validPlays);
  const phaseName = t(
    isDeclarePhase
      ? phase === BotifarraPhase.DELEGATED
        ? 'phase.delegated'
        : 'phase.declare'
      : isDoublePhase
        ? 'phase.double'
        : isPlayPhase
          ? 'phase.play'
          : isRoundEnd
            ? 'phase.roundEnd'
            : 'phase.gameEnd',
  );
  const trumpLabel =
    state.trumpSuit === BOTIFARRA_NO_TRUMP
      ? t('trump.none')
      : t(TRUMP_CHOICES.find((c) => c.suit === state.trumpSuit)?.key ?? 'trump.none');

  return (
    <GamePageShell
      title={tc('nav.botifarra')}
      gameThemeBg={gameTheme.botifarra.bg}
      phaseName={phaseName}
      gamePath="/botifarra"
      gameEndFlag={state.gameEndFlag}
      winShow={state.gameEndFlag && state.winnerTeam === 0}
      lossShow={state.gameEndFlag && state.winnerTeam === 1}
      loading={loading}
      isHumanTurn={state.isHumanTurn}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="bf-score">
            {t('label.score')}: {state.scores[0] ?? 0} / {state.scores[1] ?? 0}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div data-testid="card-area" className={`overflow-y-auto pt-3 px-4 lg:px-8 flex-1 ${lgCardAreaConstraint}`}>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="text-ds-text-primary text-center text-sm mb-2 space-y-1">
              <div data-testid="botifarra-trump">
                {t('label.trump')}: {trumpLabel}
                {state.multiplier > 1 && ` (x${state.multiplier})`}
              </div>
              <div data-testid="botifarra-round-points">
                {t('label.roundPoints')}: {state.roundPoints[0] ?? 0} / {state.roundPoints[1] ?? 0} (
                {BOTIFARRA_TOTAL_POINTS})
              </div>
            </div>

            {state.currentTrick.length > 0 && (
              <div className="mb-4" data-testid="botifarra-trick">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.trick')}</div>
                <div className="flex justify-center gap-3 flex-wrap">
                  {state.currentTrick.map((tc) => (
                    <div key={`trick-${tc.playerIdx}`} className="flex flex-col items-center">
                      <span className="text-ds-text-muted text-xs mb-1">#{tc.playerIdx}</span>
                      <AnimatedCard card={tc.card} width={cardWidth} />
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="flex justify-center gap-4 flex-wrap mb-4" data-testid="botifarra-seats">
              {state.players.map((p) => (
                <div key={`seat-${p.id}`} className="text-center text-xs text-ds-text-muted">
                  <div>
                    #{p.id} {p.isHuman ? t('label.you') : p.id === 2 ? t('label.partner') : t('label.opponent')}
                  </div>
                  <div>
                    {p.cardCount} {t('label.cards')} / {p.trickCount} {t('label.tricks')}
                  </div>
                </div>
              ))}
            </div>

            {human && human.cards.length > 0 && (
              <div data-tutorial="bf-hand">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.hand')}</div>
                <div className="flex justify-center gap-1 flex-wrap" data-testid="botifarra-hand">
                  {human.cards.map((card, i) => {
                    const playable = isPlayPhase && state.isHumanTurn && legal.has(i);
                    return (
                      <button
                        key={`hand-${card.design}-${card.value}-${i}`}
                        type="button"
                        onClick={() => handlePlay(i)}
                        disabled={loading || !playable}
                        aria-disabled={!playable}
                        aria-label={`${card.design} ${card.value}`}
                        className={playable ? 'ring-2 ring-ds-success rounded' : 'opacity-60'}
                      >
                        <AnimatedCard card={card} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
                {isPlayPhase && <p className="text-ds-text-muted text-center text-xs mt-1">{t('playGuide')}</p>}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.botifarra.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {isDeclarePhase && state.isHumanTurn && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="bf-declare">
                <p className="text-ds-text-muted text-sm">{t('declareGuide')}</p>
                <div className="flex justify-center gap-2 flex-wrap">
                  {TRUMP_CHOICES.map((c) => (
                    <button
                      key={`trump-${c.suit}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleDeclare(c.suit)}
                      disabled={loading}
                    >
                      {t('button.declare', { trump: t(c.key) })}
                    </button>
                  ))}
                  {phase === BotifarraPhase.DECLARE && (
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => execApi('delegate')}
                      disabled={loading}
                    >
                      {t('button.delegate')}
                    </button>
                  )}
                </div>
              </div>
            )}

            {isDoublePhase && state.isHumanTurn && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <p className="text-ds-text-muted text-sm">{t('doubleGuide')}</p>
                <div className="flex justify-center gap-2 flex-wrap">
                  <button type="button" className={btnWarning} onClick={() => execApi('double')} disabled={loading}>
                    {t('button.double')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => execApi('passdouble')}
                    disabled={loading}
                  >
                    {t('button.passDouble')}
                  </button>
                </div>
              </div>
            )}

            <div className="flex justify-center gap-2 pb-2 flex-wrap">
              {isRoundEnd && !state.gameEndFlag && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => execApi('next')}
                  disabled={loading}
                  data-testid="bf-next-button"
                  aria-keyshortcuts="n"
                >
                  {t('button.next')}
                  <KbdBadge label={t('kbd.next')} />
                </button>
              )}
              <GameResetButton
                isGameEnd={state.gameEndFlag}
                onReset={() => execApi('reset')}
                requestConfirm={requestConfirm}
                loading={loading}
              />
              {!state.gameEndFlag && (
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={() => execApi('giveup')}
                  disabled={loading}
                  aria-keyshortcuts="g"
                >
                  {t('button.giveUp')}
                  <KbdBadge label={t('kbd.giveUp')} />
                </button>
              )}
              <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
