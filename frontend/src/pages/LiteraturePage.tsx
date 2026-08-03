import { useEffect, useMemo, useState } from 'react';
import { literatureApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { LiteratureResponse } from '../types/card';
import { LiteraturePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { LITERATURE_HELP, parseLiteratureCommand } from '../utils/cli/commands/literatureCommands';
import { formatLiteratureState } from '../utils/cli/formatters/literatureFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Half-suit ownership codes (sync: `LiteratureHalfSuitState`). */
const STATE_OPEN = 0;
const STATE_CANCELLED = 3;

/**
 * The wire carries suits by name, but `ask` takes the numeric suit the domain
 * uses (1 = Spade … 4 = Diamond), so the two have to be bridged here.
 */
const SUIT_NUMBER: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/** Literature tutorial step definitions. */
const LITERATURE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="literature-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="literature-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="literature-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="literature-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="literature-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const LITERATURE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [LiteraturePhase.PLAY]: 'play',
  [LiteraturePhase.GAME_END]: 'gameEnd',
};

/** Renders the Literature game page: a deduction fishing game over eight half-suits. */
export const LiteraturePage = withTutorial(LiteraturePageContent, 'literature', LITERATURE_TUTORIAL_STEPS);

/** Inner content of the Literature page, wrapped by TutorialProvider. */
function LiteraturePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('literature');
  const { state, loading, error, exec, retry } = useGameApi(literatureApi.exec);

  const [askTarget, setAskTarget] = useState(1);
  const [askCard, setAskCard] = useState('');
  // **所在の申告は組とセットで持つ。**別々に持つと、組を切り替えたときに前の組の
  // ために選んだ席がそのまま残り、選んでいない配置で宣言してしまう。
  const [claim, setClaim] = useState<{ half: number; holders: number[] }>({
    half: 0,
    holders: [0, 0, 0, 0, 0, 0],
  });

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('literature');
  const cliConfig: CliGameConfig<LiteratureResponse, Parameters<typeof literatureApi.exec>> = useMemo(
    () => ({
      gameName: 'literature',
      parseCommand: parseLiteratureCommand,
      formatResponse: formatLiteratureState,
      helpText: LITERATURE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('literature', LITERATURE_PHASE_KEYS);

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('literature', state);

  if (!state)
    return <GameSkeleton gameKey="literature" layout={{ kind: 'trick-taking', trickArea: false, footerHandSize: 8 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === LiteraturePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.currentPlayerIdx === 0;
  const humanWon = isGameEnd && state.winnerTeam === 0;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitLabel = (s: number): string => t(`suitName.${s}`);
  const suitOf = (design: string): number => SUIT_NUMBER[design] ?? 1;
  const halfSuitLabel = (half: number): string =>
    `${suitLabel(1 + Math.floor(half / 2))} ${half % 2 === 0 ? t('lowHalf') : t('highHalf')}`;
  const stateLabel = (st: number): string => {
    if (st === STATE_OPEN) return t('stateOpen');
    if (st === STATE_CANCELLED) return t('stateCancelled');
    return t('stateTeam', { n: st - 1 });
  };

  // **要求できるのは相手チームのみ。**人間は席 0 = チーム 0。
  const opponents = state.players.filter((p) => p.team !== 0 && p.cardCount > 0).map((p) => p.id);

  // 未決の組の札だけを選択肢にする。
  const askableCards = state.halfSuits.flatMap((st, half) =>
    st === STATE_OPEN ? (state.halfSuitCards[half] ?? []).map((c, i) => ({ key: `${half}-${i}`, half, card: c })) : [],
  );

  const openHalfSuits = state.halfSuits.flatMap((st, half) => (st === STATE_OPEN ? [half] : []));
  // 選んでいた組が決着したら、開いている組へ寄せる。
  const selectedHalf = openHalfSuits.includes(claim.half) ? claim.half : (openHalfSuits[0] ?? 0);
  // **寄せた先の組には、まだ何も申告していない。**前の組の配置を持ち越さない。
  const claimHolders = claim.half === selectedHalf ? claim.holders : [0, 0, 0, 0, 0, 0];
  const selectedTarget = opponents.includes(askTarget) ? askTarget : (opponents[0] ?? 1);
  const selectedCardKey = askableCards.some((a) => a.key === askCard) ? askCard : (askableCards[0]?.key ?? '');
  const selectedCard = askableCards.find((a) => a.key === selectedCardKey);

  // 自チームの席だけが所在の候補になる。
  const ownTeamSeats = state.players.filter((p) => p.team === 0).map((p) => p.id);

  const handleAsk = () => {
    if (!selectedCard) return;
    exec('ask', { target: selectedTarget, suit: suitOf(selectedCard.card.design), value: selectedCard.card.value });
  };

  const handleClaim = () => {
    exec('claim', { halfSuit: selectedHalf, holders: claimHolders });
  };

  const setHolder = (i: number, seat: number) => {
    setClaim({ half: selectedHalf, holders: claimHolders.map((v, j) => (j === i ? seat : v)) });
  };

  const handleManualReset = () => {
    hideActionLog();
    exec('reset');
  };

  return (
    <GamePageShell
      title={tc('nav.literature')}
      gameThemeBg={gameTheme.literature.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/literature"
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
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Five of eight wins — four is exactly half and decides nothing. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="literature-scores">
              <div data-tutorial="literature-info">
                {t('scoreLine', {
                  t0: state.teamHalfSuits[0],
                  t1: state.teamHalfSuits[1],
                  cancelled: state.cancelledCount,
                  open: state.openCount,
                })}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="literature-threshold-note">
                {t('thresholdNote')}
              </div>
            </div>

            {/* Half-suit ownership, cancelled shown distinctly. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="literature-halfsuits">
              <div className="mb-1 text-ds-text-primary">{t('halfSuitsTitle')}</div>
              <div className="flex flex-wrap gap-x-4">
                {state.halfSuits.map((st, half) => (
                  <span
                    key={`half-${halfSuitLabel(half)}`}
                    className={st === STATE_CANCELLED ? 'text-ds-warning' : 'text-ds-text-primary'}
                  >
                    [{half}] {halfSuitLabel(half)}: {stateLabel(st)}
                  </span>
                ))}
              </div>
            </div>

            {/* Players — every other hand, teammates included, stays hidden. */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="literature-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="literature-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{t('seat', { n: p.id })}</span>
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>({t('team', { n: p.team })})</span>
                  <span>{t('hiddenHand', { count: p.cardCount })}</span>
                  {p.isCurrentTurn && !isGameEnd && <span className="text-ds-accent">[{t('turnTag')}]</span>}
                </div>
              ))}
              <div className="mt-1 text-xs text-ds-text-muted" data-testid="literature-hidden-note">
                {t('hiddenNote')}
              </div>
            </div>

            {/* Ask history — public information, and the raw material for deduction. */}
            {state.asks.length > 0 && (
              <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="literature-history">
                <div className="mb-1 text-ds-text-primary">{t('historyTitle')}</div>
                {state.asks.slice(-5).map((a, i) => (
                  <div key={`ask-${a.from}-${a.to}-${a.card?.design}-${a.card?.value}-${i}`}>
                    {t(a.success ? 'askHitLine' : 'askMissLine', {
                      from: a.from,
                      to: a.to,
                      card: a.card ? `${suitLabel(suitOf(a.card.design))}${a.card.value}` : '?',
                    })}
                  </div>
                ))}
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

          {/* Footer */}
          <GameFooter className={`${gameTheme.literature.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="literature-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => (
                  <CardImage key={`hand-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            <label className="flex items-center gap-1 text-ds-text-primary text-xs cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {isHumanTurn && (
              <div className="flex flex-col gap-2" data-tutorial="literature-actions">
                {/* Ask — opponents only. */}
                <div className="text-ds-text-muted text-xs" data-testid="literature-ask-rules">
                  {t('askRules')}
                </div>
                <div className="flex flex-wrap gap-2 items-center">
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="literature-target">
                    {t('askTargetLabel')}
                    <select
                      id="literature-target"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={selectedTarget}
                      onChange={(e) => setAskTarget(Number(e.target.value))}
                    >
                      {opponents.map((seat) => (
                        <option key={seat} value={seat}>
                          {t('seat', { n: seat })}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="literature-card">
                    {t('askCardLabel')}
                    <select
                      id="literature-card"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={selectedCardKey}
                      onChange={(e) => setAskCard(e.target.value)}
                    >
                      {askableCards.map((a) => (
                        <option key={a.key} value={a.key}>
                          {suitLabel(suitOf(a.card.design))}
                          {a.card.value}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button type="button" className={btnPrimary} onClick={handleAsk} disabled={loading || !selectedCard}>
                    {t('askButton')}
                  </button>
                </div>

                {/* Claim — placing all six, and misplacing within your own team cancels it. */}
                <div className="text-ds-text-muted text-xs" data-testid="literature-claim-rules">
                  {t('claimRules')}
                </div>
                <div className="flex flex-wrap gap-2 items-center">
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="literature-half">
                    {t('claimHalfLabel')}
                    <select
                      id="literature-half"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={selectedHalf}
                      onChange={(e) => setClaim({ half: Number(e.target.value), holders: [0, 0, 0, 0, 0, 0] })}
                    >
                      {openHalfSuits.map((half) => (
                        <option key={half} value={half}>
                          {halfSuitLabel(half)}
                        </option>
                      ))}
                    </select>
                  </label>
                  {(state.halfSuitCards[selectedHalf] ?? []).map((c, i) => (
                    <label
                      key={`holder-${c.design}-${c.value}`}
                      className="text-ds-text-muted text-xs flex items-center gap-1"
                      htmlFor={`literature-holder-${i}`}
                    >
                      {suitLabel(suitOf(c.design))}
                      {c.value}
                      <select
                        id={`literature-holder-${i}`}
                        className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                        value={claimHolders[i] ?? 0}
                        onChange={(e) => setHolder(i, Number(e.target.value))}
                      >
                        {ownTeamSeats.map((seat) => (
                          <option key={seat} value={seat}>
                            {seat}
                          </option>
                        ))}
                      </select>
                    </label>
                  ))}
                  <button type="button" className={btnSuccess} onClick={handleClaim} disabled={loading}>
                    {t('claimButton')}
                  </button>
                </div>
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center mt-2">
              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {state.winnerTeam < 0 ? t('draw') : humanWon ? t('win') : t('lose')}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="literature-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
