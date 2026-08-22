import { useEffect, useMemo, useRef, useState } from 'react';
import { ristikontraApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, RistikontraPlayer, RistikontraResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseRistikontraCommand, RISTIKONTRA_HELP } from '../utils/cli/commands/ristikontraCommands';
import { formatRistikontraState } from '../utils/cli/formatters/ristikontraFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ristikontra is always a fixed 2-vs-2 partnership, so the table is always 4. */
const RISTIKONTRA_PLAYER_COUNT = 4;

/** CPU difficulty options for the Ristikontra settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
];

/** Ristikontra tutorial step definitions. */
const RISTIKONTRA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ristikontra-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ristikontra-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ristikontra-pile"]',
    messageKey: 'tutorial.pile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ristikontra-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ristikontra-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps the backend Ristikontra phase strings to i18n phase-label keys. */
const RISTIKONTRA_PHASE_KEYS: Readonly<Record<string, string>> = {
  play: 'phase.play',
  roundEnd: 'phase.roundEnd',
  gameEnd: 'phase.gameEnd',
};

/** Renders the Ristikontra page: a Finnish 2-vs-2 capture (fishing) game. */
export const RistikontraPage = withTutorial(RistikontraPageContent, 'ristikontra', RISTIKONTRA_TUTORIAL_STEPS);

/** Inner content of the Ristikontra page, wrapped by TutorialProvider. */
function RistikontraPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ristikontra');
  const { state, loading, error, exec, retry } = useGameApi(ristikontraApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  // **席数は 4 固定。** 2 対 2 の固定パートナーシップなので、選ばせる余地は無い。
  // クローン元のピシュティは 2〜4 人を選べたが、その選択肢を残すとチームを
  // 組めない卓を頼めてしまう (バックエンドは弾くので、押しても何も起きない)。
  const playerCnt = RISTIKONTRA_PLAYER_COUNT;

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleDifficultyChange = (value: string) => {
    const level = Number(value);
    setCpuDifficulty(level);
    exec('reset', { config: { cpuDifficulty: level, playerCnt } });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ristikontra');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ristikontra', state);
  const cliConfig: CliGameConfig<RistikontraResponse, Parameters<typeof ristikontraApi.exec>> = useMemo(
    () => ({
      gameName: 'ristikontra',
      parseCommand: parseRistikontraCommand,
      formatResponse: formatRistikontraState,
      helpText: RISTIKONTRA_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  // Celebrate a **counter** the instant it lands. Ristikontra's highlight is the
  // steal — laying the rank that just captured takes the whole bundle away — and
  // it is easy to miss amid fast CPU turns.
  //
  // **The signal is a captured count going DOWN.** Cards only ever leave a
  // player's pile by being countered, so a decrease is an unambiguous steal.
  // (The clone source celebrated a Pişti bonus rising; Ristikontra has no bonus,
  // so that badge could never fire here.)
  const [counterCelebration, setCounterCelebration] = useState<{ key: number } | null>(null);
  const prevCapturedRef = useRef<number[] | null>(null);
  useEffect(() => {
    if (!state) return;
    const current = state.players.map((p) => p.capturedCount);
    const prev = prevCapturedRef.current;
    prevCapturedRef.current = current;
    if (prev === null || prev.length !== current.length) {
      setCounterCelebration(null);
      return;
    }
    const total = current.reduce((a, b) => a + b, 0);
    const prevTotal = prev.reduce((a, b) => a + b, 0);
    // A reset or a new deal drops everyone's pile; that is not a steal.
    if (total < prevTotal) {
      setCounterCelebration(null);
      return;
    }
    if (current.some((v, i) => v < prev[i])) {
      setCounterCelebration((c) => ({ key: (c?.key ?? 0) + 1 }));
      playSound('chipClick', { pitchVariation: 0.1 });
    }
  }, [state, playSound]);

  if (!state)
    return <GameSkeleton gameKey="ristikontra" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;

  const phaseName = t(RISTIKONTRA_PHASE_KEYS[state.phase] ?? '');

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
  const isHumanTurn = state.phase === 'play' && state.currentTurn === 0 && !isGameEnd;
  const humanWon = isGameEnd && state.winners.includes(0);

  // On the human's turn, hint which cards can capture the pile: a Jack takes the whole
  // pile (accent), and a card matching the pile-top rank captures it (success). Pure
  // client-side derivation from pileTop + the hand.
  const JACK_VALUE = 11;
  const captureRing = (c: Card): string => {
    if (!isHumanTurn) return '';
    if (c.value === JACK_VALUE) return 'ring-2 ring-ds-accent motion-safe:animate-pulse';
    if (state.pileTop && c.value === state.pileTop.value) return 'ring-2 ring-ds-success motion-safe:animate-pulse';
    return '';
  };

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  // **途中経過はチームの獲得枚数そのもの。** 札ごとの点数もボーナスも無いので
  // 近似ではなく確定値。席には自分のチームの合計を出す —— 席ごとの枚数だと
  // 「自分は取っているのに負けている」が読めない。
  const teamCaptured = [0, 0];
  for (const p of state.players) teamCaptured[p.id % 2] += p.capturedCount;
  const provisionalScore = (p: RistikontraPlayer): number => teamCaptured[p.id % 2];

  const handleManualReset = () => {
    hideActionLog();
    exec('reset', { config: { cpuDifficulty, playerCnt } });
  };

  return (
    <GamePageShell
      title={tc('nav.ristikontra')}
      gameThemeBg={gameTheme.ristikontra.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/ristikontra"
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    onSelect: handleDifficultyChange,
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="ristikontra-info">
              <span>{t('deck', { count: state.remainingDeck })}</span>
            </div>

            {/* Players (hand / captured / Pişti bonus / current turn) */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="ristikontra-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {!isGameEnd && (
                <div className="mb-1 text-ds-text-muted text-xs" data-testid="ristikontra-provisional-note">
                  {t('provisionalNote')}
                </div>
              )}
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  className={`text-sm py-0.5 flex items-center gap-3 ${
                    p.id === state.currentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>{t('captured', { count: p.capturedCount })}</span>
                  <span className="text-ds-text-muted">{t('teamLabel', { team: (p.id % 2) + 1 })}</span>
                  {!isGameEnd && (
                    <span
                      className="text-ds-text-primary"
                      data-testid={`ristikontra-provisional-${p.id}`}
                      title={t('provisionalNote')}
                    >
                      {t('provisional', { score: provisionalScore(p) })}
                      {teamCaptured[p.id % 2] > teamCaptured[(p.id + 1) % 2] && (
                        <span className="ml-1 text-ds-accent">★</span>
                      )}
                    </span>
                  )}
                  {isGameEnd && (
                    <span className="text-ds-text-primary">{t('finalScore', { score: p.finalScore })}</span>
                  )}
                </div>
              ))}
            </div>

            {/* Center pile */}
            <div className="relative mb-2 p-3 rounded bg-black/20 text-center" data-tutorial="ristikontra-pile">
              {counterCelebration && (
                <div
                  key={counterCelebration.key}
                  className="absolute inset-x-0 -top-2 z-10 flex justify-center motion-safe:animate-bounce pointer-events-none"
                  role="status"
                  data-testid="ristikontra-counter-celebration"
                >
                  <span className="rounded-full bg-ds-accent px-3 py-0.5 text-sm font-bold text-ds-text-on-accent shadow-lg">
                    {t('counterCelebration')}
                  </span>
                </div>
              )}
              <div className="text-ds-text-muted text-xs mb-1">
                {t('pile')} — {t('pileCount', { count: state.pileCount })}
              </div>
              {state.pileTop ? (
                <div className="flex justify-center">
                  <CardImage card={state.pileTop} width={cardWidth} />
                </div>
              ) : (
                <div className="text-ds-text-muted text-sm">{t('pileEmpty')}</div>
              )}
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

          {/* Footer */}
          <GameFooter className={`${gameTheme.ristikontra.footer} px-4 py-2.5`}>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="mb-2" data-tutorial="ristikontra-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-2">
                {humanPlayer?.cards.map((c, i) => (
                  <button
                    key={`hand-${i}`}
                    type="button"
                    onClick={() => isHumanTurn && exec('play', { handIndex: i })}
                    disabled={!isHumanTurn || loading}
                    className={`rounded transition-all ${
                      isHumanTurn ? 'cursor-pointer hover:opacity-90 hover:-translate-y-1' : 'cursor-default'
                    } ${captureRing(c)}`}
                    data-testid={`hand-card-${i}`}
                    aria-label={
                      captureRing(c)
                        ? `${t('playCardAria', { card: cardAlt(c) })} — ${t('captureHint')}`
                        : t('playCardAria', { card: cardAlt(c) })
                    }
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanTurn && (
              <div className="text-ds-text-muted text-xs mb-2" role="status" data-testid="ristikontra-turn-notice">
                {t('turnNotice')}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center">
              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon
                    ? t('win')
                    : t('lose', { name: playerLabel(state.winners[0] ?? -1, state.winners[0] === 0) })}
                </span>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextGame')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ristikontra-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
