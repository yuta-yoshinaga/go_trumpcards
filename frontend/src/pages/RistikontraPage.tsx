import type { TFunction } from 'i18next';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ristikontraApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { Card, RistikontraPlayer, RistikontraResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseRistikontraCommand, RISTIKONTRA_HELP } from '../utils/cli/commands/ristikontraCommands';
import { formatRistikontraState } from '../utils/cli/formatters/ristikontraFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
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

type RistikontraGame = ReturnType<typeof useGame>;
type RistikontraView = ReturnType<typeof useView>;
type RistikontraContext = GamePageContext<RistikontraResponse, RistikontraGame, RistikontraView>;

function useGame() {
  const { state, loading, error, exec, retry } = useGameApi(ristikontraApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  // **席数は 4 固定。** 2 対 2 の固定パートナーシップなので、選ばせる余地は無い。
  // クローン元のピシュティは 2〜4 人を選べたが、その選択肢を残すとチームを
  // 組めない卓を頼めてしまう (バックエンドは弾くので、押しても何も起きない)。
  const playerCnt = RISTIKONTRA_PLAYER_COUNT;

  // Fetch a fresh game on mount.
  useMountReset(exec);

  const handleDifficultyChange = useCallback(
    (value: string) => {
      const level = Number(value);
      setCpuDifficulty(level);
      void exec('reset', { config: { cpuDifficulty: level, playerCnt } });
    },
    [exec],
  );

  const handleReset = useCallback(() => {
    void exec('reset', { config: { cpuDifficulty, playerCnt } });
  }, [cpuDifficulty, exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    cpuDifficulty,
    handleDifficultyChange,
    handleReset,
  };
}

function useView({ game, t }: { game: RistikontraGame; t: TFunction }) {
  const state = game.state;
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  const phaseNames: Readonly<Record<string, string>> = useMemo(
    () => ({
      play: t(RISTIKONTRA_PHASE_KEYS.play),
      roundEnd: t(RISTIKONTRA_PHASE_KEYS.roundEnd),
      gameEnd: t(RISTIKONTRA_PHASE_KEYS.gameEnd),
    }),
    [t],
  );

  const [selectedIdx, setSelectedIdx] = useState<number | null>(null);
  const toggleSelect = useCallback((idx: number) => {
    setSelectedIdx((prev) => (prev === idx ? null : idx));
  }, []);
  const clearSelect = useCallback(() => setSelectedIdx(null), []);
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset selection when the turn or phase changes.
  useEffect(() => {
    setSelectedIdx(null);
  }, [state?.currentTurn, state?.phase]);
  const confirmPlay = useCallback(() => {
    if (selectedIdx !== null) {
      void game.exec('play', { handIndex: selectedIdx });
      clearSelect();
    }
  }, [clearSelect, game.exec, selectedIdx]);

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

  return {
    cardWidth,
    phaseNames,
    selectedIdx,
    toggleSelect,
    clearSelect,
    confirmPlay,
    counterCelebration,
  };
}

function Board({ state, t, view }: RistikontraContext) {
  const { cardWidth, counterCelebration } = view;
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  // **途中経過はチームの獲得枚数そのもの。** 札ごとの点数もボーナスも無いので
  // 近似ではなく確定値。席には自分のチームの合計を出す —— 席ごとの枚数だと
  // 「自分は取っているのに負けている」が読めない。
  const teamCaptured = [0, 0];
  for (const p of state.players) teamCaptured[p.id % 2] += p.capturedCount;
  const provisionalScore = (p: RistikontraPlayer): number => teamCaptured[p.id % 2];

  return (
    <>
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
            <span>{t('handCount', { count: p.cardCount })}</span>
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
            {isGameEnd && <span className="text-ds-text-primary">{t('finalScore', { score: p.finalScore })}</span>}
          </div>
        ))}
      </div>

      {/* Center pile */}
      <div className="relative mb-2 p-3 rounded bg-black/20 text-center" data-tutorial="ristikontra-pile">
        <div
          key={counterCelebration?.key}
          className="sr-only"
          role="status"
          aria-live="polite"
          data-testid="ristikontra-counter-announcement"
        >
          {counterCelebration ? t('counterCelebration') : ''}
        </div>
        {counterCelebration && (
          <div
            key={`celebration-${counterCelebration.key}`}
            className="absolute inset-x-0 -top-2 z-10 flex justify-center motion-safe:animate-bounce pointer-events-none"
            aria-hidden="true"
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

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, game, t, view, busy, frontendHint }: RistikontraContext) {
  const { cardWidth, selectedIdx } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
  const isHumanTurn = state.phase === 'play' && state.currentTurn === 0 && !isGameEnd;

  // On the human's turn, ring the cards that actually do something:
  //
  //   - a card matching `counterRank` steals the bundle the last capture took
  //     (accent — the biggest swing available), and
  //   - a card matching the pile top captures the pile (success).
  //
  // **No Jack ring.** The clone source (Pişti) treats a Jack as a wild sweep and
  // ringed every one of them; here a Jack that does not match is an ordinary
  // card, and ringing it tells the player "this captures" when it does not.
  const captureRing = (c: Card): string => {
    if (!isHumanTurn) return '';
    if (state.counterRank > 0 && c.value === state.counterRank)
      return 'ring-2 ring-ds-accent motion-safe:animate-pulse';
    if (state.pileTop && c.value === state.pileTop.value) return 'ring-2 ring-ds-success motion-safe:animate-pulse';
    return '';
  };

  const captureAriaHint = (c: Card): string | null => {
    if (!isHumanTurn) return null;
    const canCounter = state.counterRank > 0 && c.value === state.counterRank;
    const canCapture = !!state.pileTop && c.value === state.pileTop.value;
    if (canCounter && canCapture) return t('counterAndCaptureHint');
    if (canCounter) return t('counterHint');
    if (canCapture) return t('captureHint');
    return null;
  };

  return (
    <>
      {/* Above the hand, as on the hand-wired page: the tooltip is a block in the footer flow, not an overlay. */}
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      <div className="mb-2" data-tutorial="ristikontra-hand">
        <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
        <div className="flex flex-wrap gap-2">
          {humanPlayer?.cards.map((c, i) => (
            <button
              key={`hand-${i}`}
              type="button"
              onClick={() => isHumanTurn && void game.exec('play', { handIndex: i })}
              disabled={!isHumanTurn || busy}
              className={`rounded transition-all ${
                isHumanTurn ? 'cursor-pointer hover:opacity-90 hover:-translate-y-1' : 'cursor-default'
              } ${captureRing(c)} ${selectedIdx === i ? 'ring-2 ring-ds-warning' : ''}`}
              aria-pressed={selectedIdx === i}
              data-testid={`hand-card-${i}`}
              aria-label={[t('playCardAria', { card: cardAlt(c) }), captureAriaHint(c)].filter(Boolean).join(' — ')}
            >
              <CardImage card={c} width={cardWidth} />
            </button>
          ))}
        </div>
      </div>
    </>
  );
}

function FooterStatus({ state, t }: RistikontraContext) {
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
  const isHumanTurn = state.phase === 'play' && state.currentTurn === 0 && !isGameEnd;

  return (
    <>
      {isHumanTurn && (
        <div className="text-ds-text-muted text-xs mb-2" role="status" data-testid="ristikontra-turn-notice">
          {t('turnNotice')}
        </div>
      )}
    </>
  );
}

function FooterActions({ state, game, t, busy }: RistikontraContext) {
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
  const humanWon = isGameEnd && state.winners.some((winner) => winner % 2 === 0);
  const isDraw = isGameEnd && state.winners.length === 0;
  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  return (
    <>
      {isGameEnd && (
        <span className="text-ds-text-primary text-sm font-semibold mr-1">
          {isDraw ? t('draw') : humanWon ? t('win') : t('lose', { name: playerLabel(state.winners[0], false) })}
        </span>
      )}
      {isGameEnd && (
        <button type="button" className={btnSuccess} onClick={() => void game.exec('next')} disabled={busy}>
          {t('nextGame')}
        </button>
      )}
    </>
  );
}

const ristikontraDefinition: GameDefinition<
  RistikontraResponse,
  RistikontraGame,
  Parameters<typeof ristikontraApi.exec>,
  RistikontraView
> = {
  key: 'ristikontra',
  useGame,
  useView,
  tutorial: RISTIKONTRA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 4 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseRistikontraCommand,
    formatResponse: formatRistikontraState,
    helpText: RISTIKONTRA_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
    const humanWon = isGameEnd && state.winners.some((winner) => winner % 2 === 0);
    return {
      name: view.phaseNames[state.phase] ?? '',
      ended: isGameEnd,
      won: humanWon,
    };
  },
  shell: ({ state }) => {
    const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
    const isHumanTurn = state.phase === 'play' && state.currentTurn === 0 && !isGameEnd;
    return {
      isHumanTurn,
    };
  },
  reset: ({ game }) => {
    game.handleReset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label}`),
          })),
          onSelect: game.handleDifficultyChange,
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ state, view }) => ({
    cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
    onToggle: view.toggleSelect,
    onConfirm: view.confirmPlay,
    onClear: view.clearSelect,
    enabled: state.phase === 'play' && state.currentTurn === 0 && !state.gameEndFlag,
  }),
  shortcutsTestId: 'ristikontra-kbd-shortcuts',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  resetTutorialId: 'ristikontra-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Ristikontra page: a Finnish 2-vs-2 capture (fishing) game. */
export const RistikontraPage = createGamePage(ristikontraDefinition);
