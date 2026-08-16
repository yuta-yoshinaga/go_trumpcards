import { useEffect, useMemo, useState } from 'react';
import { klaberjassApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KlaberjassResponse } from '../types/card';
import { KlaberjassPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KLABERJASS_HELP, parseKlaberjassCommand } from '../utils/cli/commands/klaberjassCommands';
import { formatKlaberjassState } from '../utils/cli/formatters/klaberjassFormatter';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Target-score options for the Klaberjass settings panel. */
const TARGET_OPTIONS = [301, 501, 1000];

/** Suit glyphs by design value (1=Spade … 4=Diamond). */
const SUIT_GLYPHS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/**
 * The trump order with its point values.
 *
 * Shown permanently because it is the one thing that reliably trips people up:
 * in trumps the jack and nine jump above the ace, so a hand that looks strong
 * by ordinary reckoning may not be.
 */
const TRUMP_LADDER = [
  { label: 'J', points: 20 },
  { label: '9', points: 14 },
  { label: 'A', points: 11 },
  { label: '10', points: 10 },
  { label: 'K', points: 4 },
  { label: 'Q', points: 3 },
  { label: '8', points: 0 },
  { label: '7', points: 0 },
];

/** Klaberjass tutorial step definitions. */
const KLABERJASS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="klaberjass-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KLABERJASS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KlaberjassPhase.BID_TURN_UP]: 'bidTurnUp',
  [KlaberjassPhase.BID_FREE]: 'bidFree',
  [KlaberjassPhase.SCHMEISS]: 'schmeiss',
  [KlaberjassPhase.PLAY]: 'play',
  [KlaberjassPhase.HAND_END]: 'handEnd',
  [KlaberjassPhase.GAME_END]: 'gameEnd',
};

/** Renders the Klaberjass game page: the two-player ancestor of the Jass family. */
export const KlaberjassPage = withTutorial(KlaberjassPageContent, 'klaberjass', KLABERJASS_TUTORIAL_STEPS);

/** Inner content of the Klaberjass page, wrapped by TutorialProvider. */
function KlaberjassPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('klaberjass');
  const { state, loading, error, exec, retry } = useGameApi(klaberjassApi.exec);

  const [targetScore, setTargetScore] = useState(501);
  const [selected, setSelected] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleTargetChange = (value: string) => {
    const next = Number(value);
    setTargetScore(next);
    exec('reset', { config: { targetScore: next } });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('klaberjass');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('klaberjass', state);
  const cliConfig: CliGameConfig<KlaberjassResponse, Parameters<typeof klaberjassApi.exec>> = useMemo(
    () => ({
      gameName: 'klaberjass',
      parseCommand: parseKlaberjassCommand,
      formatResponse: formatKlaberjassState,
      helpText: KLABERJASS_HELP,
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(frontendHint) : null),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('klaberjass', KLABERJASS_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="klaberjass" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isBidTurnUp = state.phase === KlaberjassPhase.BID_TURN_UP;
  const isBidFree = state.phase === KlaberjassPhase.BID_FREE;
  const isSchmeiss = state.phase === KlaberjassPhase.SCHMEISS;
  const isPlay = state.phase === KlaberjassPhase.PLAY;
  const isHandEnd = state.phase === KlaberjassPhase.HAND_END;
  const isGameEnd = state.phase === KlaberjassPhase.GAME_END || state.gameEndFlag;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isHumanBid = (isBidTurnUp || isBidFree) && state.bidPlayerIdx === 0 && !isGameEnd;
  // 投げの提案に答えるのは、提案した「相手」ではないほう。
  const isHumanSchmeissAnswer = isSchmeiss && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitGlyph = (suit: number): string => SUIT_GLYPHS[suit] ?? t('trumpNone');

  const canPlay = (i: number) => state.validPlays.includes(i);

  const handlePlay = () => {
    if (selected === null) return;
    exec('play', { cardIndex: selected });
    setSelected(null);
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelected(null);
    exec('reset', { config: { targetScore } });
  };

  return (
    <GamePageShell
      title={tc('nav.klaberjass')}
      gameThemeBg={gameTheme.klaberjass.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBid || isHumanSchmeissAnswer || isHumanPlay}
      gamePath="/klaberjass"
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
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: targetScore,
                    options: TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: handleTargetChange,
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="klaberjass-info">
              <span className="mr-4">{t('deal', { n: state.dealNumber })}</span>
              <span className="mr-4">{t('target', { n: state.targetScore })}</span>
              <span className="mr-4">
                {t('trump')}: {suitGlyph(state.trumpSuit)}
              </span>
              {state.turnUpCard && state.trumpSuit === 0 && (
                <span data-testid="klaberjass-turnup">
                  {t('turnUp')}: {suitGlyph(suitOf(state.turnUpCard.design))}
                </span>
              )}
            </div>

            {/* The trump ladder — the one rule people reliably get wrong. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="klaberjass-ladder">
              <div className="mb-1 text-ds-text-primary">{t('ladderTitle')}</div>
              <div className="flex flex-wrap gap-x-3 gap-y-1">
                {TRUMP_LADDER.map((r) => (
                  <span
                    key={r.label}
                    className={r.points >= 14 ? 'text-ds-warning font-semibold' : 'text-ds-text-muted'}
                  >
                    {r.label} ({r.points})
                  </span>
                ))}
              </div>
              <div className="mt-1 text-ds-text-muted">{t('ladderNote')}</div>
              <div className="text-ds-text-muted">{t('sequenceNote')}</div>
            </div>

            {/* Turn-up card while bidding */}
            {state.turnUpCard && state.trumpSuit === 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="klaberjass-turnup-card">
                <span className="text-ds-text-muted text-sm">{t('turnUp')}</span>
                <CardImage card={state.turnUpCard} width={cardWidth} />
              </div>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="klaberjass-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="klaberjass-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  {p.isMaker && <span className="text-ds-success">[{t('maker')}]</span>}
                  <span>{t('score', { n: p.score })}</span>
                  <span>{t('handPoints', { n: p.handPoints })}</span>
                  {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
                </div>
              ))}
            </div>

            {/* Trick */}
            {state.trick.length > 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="klaberjass-trick">
                <span className="text-ds-text-muted text-sm">{t('trick')}</span>
                {state.trick.map((c, i) => (
                  <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            )}

            {/* Settlement */}
            {(isHandEnd || isGameEnd) && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="klaberjass-settlement"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('settlementTitle')}</div>
                <div>{state.bete ? t('beteLine') : t('madeLine')}</div>
                <div>
                  {state.sequenceWinner >= 0
                    ? t('sequenceWinner', { name: playerLabel(state.sequenceWinner, state.sequenceWinner === 0) })
                    : t('sequenceNobody')}
                </div>
                {state.belaScored && state.belaHolder >= 0 && (
                  <div>{t('belaLine', { name: playerLabel(state.belaHolder, state.belaHolder === 0) })}</div>
                )}
                {state.dixUsed && <div>{t('dixLine')}</div>}
                {/* **最終トリックには 10 点が付く。**書かないと、ベラや宣言点を
                    足しても handPoints と合わない理由が説明できない (#4937)。 */}
                {state.lastTrickWinner >= 0 && (
                  <div data-testid="klaberjass-last-trick-bonus">
                    {t('lastTrickBonus', {
                      name: playerLabel(state.lastTrickWinner, state.lastTrickWinner === 0),
                      points: state.lastTrickBonus,
                    })}
                  </div>
                )}
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
          <GameFooter className={`${gameTheme.klaberjass.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="klaberjass-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => {
                  const playable = !isPlay || canPlay(i);
                  return (
                    <button
                      key={`hand-${c.design}-${c.value}-${i}`}
                      type="button"
                      data-hint-action="play"
                      onClick={() => setSelected(i)}
                      disabled={loading || (isPlay && !canPlay(i))}
                      className={`rounded ${selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                        playable ? '' : 'opacity-40'
                      }`}
                    >
                      <CardImage card={c} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanBid && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="klaberjass-bid-notice">
                {isBidTurnUp ? t('bidNotice') : t('bidFreeNotice')}
              </div>
            )}
            {isHumanSchmeissAnswer && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="klaberjass-schmeiss-notice">
                {t('schmeissNotice')}
              </div>
            )}
            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="klaberjass-play-notice">
                {t('playNotice')}
              </div>
            )}

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="klaberjass-actions">
              {isHumanBid && isBidTurnUp && (
                <button type="button" className={btnSuccess} onClick={() => exec('accept')} disabled={loading}>
                  {t('acceptButton')}
                </button>
              )}

              {isHumanBid &&
                isBidFree &&
                [1, 2, 3, 4]
                  // **断られた表向きのスートは選べない。**第1ラウンドで流れている。
                  .filter((s) => s !== (state.turnUpCard ? suitOf(state.turnUpCard.design) : 0))
                  .map((s) => (
                    <button
                      key={`call-${s}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => exec('call', { suit: s })}
                      disabled={loading}
                    >
                      {t('callButton', { suit: SUIT_GLYPHS[s] })}
                    </button>
                  ))}

              {isHumanBid && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => exec('pass')} disabled={loading}>
                    {t('passButton')}
                  </button>
                  <button type="button" className={btnWarning} onClick={() => exec('schmeiss')} disabled={loading}>
                    {t('schmeissButton')}
                  </button>
                </>
              )}

              {isHumanSchmeissAnswer && (
                <>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => exec('answerschmeiss', { accept: true })}
                    disabled={loading}
                  >
                    {t('schmeissAcceptButton')}
                  </button>
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={() => exec('answerschmeiss', { accept: false })}
                    disabled={loading}
                  >
                    {t('schmeissRefuseButton')}
                  </button>
                </>
              )}

              {isHumanPlay && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={loading || selected === null}
                >
                  {t('playButton')}
                </button>
              )}

              {isHandEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextDeal')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="klaberjass-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Maps a wire design string back to the domain's 1-based suit value. */
function suitOf(design: string): number {
  switch (design) {
    case 'SPADE':
      return 1;
    case 'CLOVER':
      return 2;
    case 'HEART':
      return 3;
    case 'DIAMOND':
      return 4;
    default:
      return 0;
  }
}
