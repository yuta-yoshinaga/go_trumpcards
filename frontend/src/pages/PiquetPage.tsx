import type { TFunction } from 'i18next';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePiquetGame } from '../hooks/usePiquetGame';
import { badgeErrorColors, badgeSuccessColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PiquetDeclaration, PiquetPlayerData, PiquetResponse } from '../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { isRequestedHint } from '../utils/hintRequest';
import { type DeclarationHighlight, declarationHighlight } from '../utils/piquetDeclarationHighlight';

const HIGHLIGHT_MS = 1500;

const PIQUET_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="piquet-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piquet-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** i18n keys for each declaration kind, mirroring `piquetKindLabel` in PiquetCuiPresenter.go. */
const DECL_KIND_KEYS: Record<number, string> = {
  [PiquetDeclarationKind.POINT]: 'declKindPoint',
  [PiquetDeclarationKind.SEQUENCE]: 'declKindSequence',
  [PiquetDeclarationKind.SET]: 'declKindSet',
};

function declKindLabel(kind: number, t: TFunction): string {
  const key = DECL_KIND_KEYS[kind];
  return key ? t(key) : '?';
}

/** Renders the Piquet game page. */
export const PiquetPage = withTutorial(PiquetPageContent, 'piquet', PIQUET_TUTORIAL_STEPS);

function PiquetPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('piquet');
  const game = usePiquetGame();
  const { state, loading, error, retry } = game;
  const { cardWidth } = useCardDimensions();
  const [selectedDiscards, setSelectedDiscards] = useState<number[]>([]);
  const [activeHighlight, setActiveHighlight] = useState<DeclarationHighlight | null>(null);
  const prevDeclLenRef = useRef(0);

  // **フックは早期 return より上。**`if (!state) return <GameSkeleton>` の
  // 下に置くと、初回レンダーだけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('piquet', state);

  const human = state?.players.find((p) => p.isHuman);
  const elderIdx = state?.elderIdx ?? 0;
  const declResultsLen = state?.declResults.length ?? 0;

  useEffect(() => {
    const prev = prevDeclLenRef.current;
    prevDeclLenRef.current = declResultsLen;
    if (declResultsLen <= prev || !state || !human) return;
    const latest = state.declResults[declResultsLen - 1];
    if (!latest) return;
    const next = declarationHighlight(latest, human.id, elderIdx, human.cards);
    if (!next) return;
    setActiveHighlight(next);
    const timer = window.setTimeout(() => setActiveHighlight(null), HIGHLIGHT_MS);
    return () => window.clearTimeout(timer);
  }, [declResultsLen, state, human, elderIdx]);

  const toggleDiscard = useCallback((idx: number) => {
    setSelectedDiscards((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  }, []);

  const handleElderExchange = useCallback(() => {
    game.handleExchangeElder(selectedDiscards);
    setSelectedDiscards([]);
  }, [game, selectedDiscards]);

  const handleYoungerExchange = useCallback(() => {
    game.handleExchangeYounger(selectedDiscards);
    setSelectedDiscards([]);
  }, [game, selectedDiscards]);

  if (!state) {
    return (
      <GameSkeleton
        gameKey="piquet"
        layout={{ kind: 'trick-taking', opponents: 1, trickArea: true, footerHandSize: 12 }}
      />
    );
  }

  const isHumanElder = human?.id === elderIdx;
  const inExchangePhase = state.phase === PiquetPhase.EXCHANGE;
  const inDeclarationPhase = state.phase === PiquetPhase.DECLARATION;
  const inPlayPhase = state.phase === PiquetPhase.PLAY;
  const inScorePhase = state.phase === PiquetPhase.SCORE;
  const inGameEndPhase = state.phase === PiquetPhase.GAME_END;

  const humanCanExchange =
    inExchangePhase &&
    human != null &&
    ((state.exchangeTurn === PiquetExchangeTurn.ELDER && isHumanElder) ||
      (state.exchangeTurn === PiquetExchangeTurn.YOUNGER && !isHumanElder));
  const humanCanPlay = inPlayPhase && human?.id === state.currentPlayerIdx;
  // presenter がマストフォローの合法手を毎回返している (PiquetWebPresenter.go)。
  // **空/未設定は「どれも出せない」ではなく「制限なし」。**交換フェーズなど、
  // そもそも載らない状態で手札を全部殺さないため (#5603)。
  // 未設定は「制限なし」。presenter 側が `omitempty` なので、プレイフェーズの
  // 人間手番以外では**キーごと来ない** ── それを「どれも出せない」と読むと
  // 交換フェーズで手札が全部死ぬ。
  const legalPlays = state.legalPlayIndices;
  const isPlayable = (i: number): boolean => legalPlays === undefined || legalPlays.includes(i);

  return (
    <GamePageShell
      title={tc('nav.piquet')}
      gameThemeBg={gameTheme.piquet.bg}
      phaseName={piquetPhaseLabel(state.phase, t)}
      isHumanTurn={humanCanPlay || humanCanExchange}
      gamePath="/piquet"
      gameEndFlag={inGameEndPhase}
      winShow={inGameEndPhase && state.winnerIdx === human?.id}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <ErrorAlert message={error} onRetry={retry} />
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {/* Scrollable state display. Piquet fits a 375x667 viewport today with
          room to spare, but it had no play area at all, so any growth would have
          gone straight into the document height. See issue #4373. */}
      <div className="flex-1 overflow-y-auto min-h-0">
        <div className="text-sm opacity-80 px-2">
          {t('dealHeader', { deal: state.dealNumber, total: state.dealsPerPartie })}
        </div>

        <PlayerCard
          label={t('roleElder')}
          player={state.players[elderIdx]}
          carteBlanche={state.carteBlanche[elderIdx]}
        />
        <PlayerCard
          label={t('roleYounger')}
          player={state.players[state.youngerIdx]}
          carteBlanche={state.carteBlanche[state.youngerIdx]}
        />

        {human ? (
          <div data-tutorial="piquet-hand" className="rounded border border-white/20 p-2 mx-2">
            <div className="mb-1 flex items-center justify-between text-xs">
              <span className="opacity-70">{t('yourHand')}</span>
              {activeHighlight ? (
                <span
                  data-testid="piquet-meld-badge"
                  className={`rounded px-2 py-0.5 text-xs font-bold ${
                    activeHighlight.won ? badgeSuccessColors : badgeErrorColors
                  }`}
                >
                  {t(activeHighlight.won ? 'meldWonBadge' : 'meldLostBadge', {
                    label: t(activeHighlight.labelKey, { count: activeHighlight.count }),
                  })}
                </span>
              ) : null}
            </div>
            <div className="flex flex-wrap gap-1">
              {human.cards.map((c, i) => {
                const selected = selectedDiscards.includes(i);
                const highlighted = activeHighlight?.cardIndices.includes(i) ?? false;
                // Highlight ring lives on the AnimatedCard wrapper so it tracks the lift; the
                // selection lift/glow is driven by AnimatedCard's isSelected (Framer Motion).
                const ringClass = highlighted
                  ? activeHighlight?.won
                    ? 'ring-2 ring-ds-success'
                    : 'ring-2 ring-ds-error'
                  : '';
                const handlePlay = () => game.handlePlay(i);
                const handleClick = humanCanExchange ? () => toggleDiscard(i) : humanCanPlay ? handlePlay : undefined;
                // 出せない札は PlayerHandSection と同じ扱い ── `aria-disabled` で
                // フォーカスは残したまま押しても何も起きない形にする。HTML の
                // `disabled` にすると読み上げ側から札そのものが消える (#5603)。
                const restricted = humanCanPlay && !isPlayable(i);
                return (
                  <button
                    key={`hand-${i}-${c.design}-${c.value}`}
                    type="button"
                    aria-pressed={humanCanExchange ? selected : undefined}
                    data-hint-action={humanCanExchange ? 'discard' : 'play'}
                    className={`rounded ${restricted ? 'opacity-50 cursor-not-allowed' : ''}`}
                    onClick={restricted ? undefined : handleClick}
                    aria-disabled={restricted || undefined}
                    disabled={handleClick == null}
                  >
                    <AnimatedCard card={c} width={cardWidth} isSelected={selected} wrapperClassName={ringClass} />
                  </button>
                );
              })}
            </div>
          </div>
        ) : null}
      </div>

      <div data-tutorial="piquet-controls" className="flex flex-wrap gap-2 px-2">
        {humanCanExchange && state.exchangeTurn === PiquetExchangeTurn.ELDER ? (
          <button
            type="button"
            onClick={handleElderExchange}
            data-hint-action="discard"
            disabled={selectedDiscards.length < 1 || selectedDiscards.length > 5}
            className={btnPrimary}
          >
            {t('exchangeElder', { count: selectedDiscards.length })}
          </button>
        ) : null}
        {humanCanExchange && state.exchangeTurn === PiquetExchangeTurn.YOUNGER ? (
          <button
            type="button"
            onClick={handleYoungerExchange}
            data-hint-action="discard"
            disabled={selectedDiscards.length > 3}
            className={btnPrimary}
          >
            {t('exchangeYounger', { count: selectedDiscards.length })}
          </button>
        ) : null}
        {inDeclarationPhase ? (
          <button type="button" onClick={game.handleResolveDeclaration} className={btnPrimary}>
            {t('advanceDeclaration', { kind: declKindLabel(state.declStage, t) })}
          </button>
        ) : null}
        {inScorePhase ? (
          <button type="button" onClick={game.handleNextDeal} className={btnSuccess}>
            {t('nextDeal')}
          </button>
        ) : null}
        {humanCanPlay || humanCanExchange ? (
          <button type="button" onClick={game.handleHint} className={btnSuccess} disabled={loading}>
            {t('hintButton')}
          </button>
        ) : null}
        <GameResetButton
          isGameEnd={inGameEndPhase}
          onReset={game.handleReset}
          requestConfirm={requestConfirm}
          loading={loading}
        />
      </div>

      <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mt-2 cursor-pointer min-h-[44px]">
        <input
          type="checkbox"
          checked={frontendHintEnabled}
          onChange={(e) => setFrontendHintEnabled(e.target.checked)}
        />
        {tc('hint.toggle', { ns: 'tutorial' })}
      </label>

      <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

      {state.hint && isRequestedHint(state) ? (
        <p className="mt-2 text-sm text-ds-accent" data-testid="piquet-hint">
          {state.hint.cardIndex !== undefined
            ? t('hintPlay', { index: state.hint.cardIndex })
            : t('hintDiscard', { indices: (state.hint.discardIndices ?? []).join(', ') })}
        </p>
      ) : null}

      {state.declResults.length > 0 ? <DeclarationList results={state.declResults} elderIdx={elderIdx} /> : null}

      {inPlayPhase && state.currentTrick.length > 0 ? <TrickView trick={state.currentTrick} /> : null}

      {inGameEndPhase ? (
        <div className="text-lg font-bold px-2">
          {state.winnerIdx === -1 ? t('partieDraw') : t('partieWinner', { idx: state.winnerIdx })}
        </div>
      ) : null}

      <ActionLogSection
        isEndPhase={inGameEndPhase}
        actionLog={actionLog}
        showActionLog={showActionLog}
        hideActionLog={hideActionLog}
      />
    </GamePageShell>
  );
}

function piquetPhaseLabel(phase: number, t: (key: string) => string): string {
  switch (phase) {
    case PiquetPhase.EXCHANGE:
      return t('exchangeHeader');
    case PiquetPhase.DECLARATION:
      return t('declarationHeader');
    case PiquetPhase.PLAY:
      return t('playHeader');
    case PiquetPhase.SCORE:
      return t('roundEnd');
    case PiquetPhase.GAME_END:
      return t('partieEnd');
    default:
      return '';
  }
}

interface PlayerCardProps {
  label: string;
  player: PiquetPlayerData | undefined;
  carteBlanche: boolean;
}

function PlayerCard({ label, player, carteBlanche }: PlayerCardProps) {
  const { t } = useTranslation('piquet');
  if (!player) return null;
  return (
    <div className="rounded border border-white/20 p-2 mx-2 text-sm">
      <div className="flex items-center justify-between">
        <span className="font-bold">{label}</span>
        {carteBlanche ? <span className="text-ds-warning">{t('carteBlanche')}</span> : null}
      </div>
      <div className="text-xs opacity-80">
        {t('playerStats', {
          hand: player.cardCount,
          tricks: player.trickCount,
          round: player.roundScore,
          match: player.matchScore,
        })}
      </div>
    </div>
  );
}

function DeclarationList({ results, elderIdx }: { results: PiquetDeclaration[]; elderIdx: number }) {
  const { t } = useTranslation('piquet');
  const youngerIdx = elderIdx === 0 ? 1 : 0;
  return (
    // role="log" defaults to aria-atomic="false", so a screen reader announces
    // only each newly-appended declaration, not the whole list again. We rely on
    // the default aria-relevant ("additions text") rather than forcing "additions"
    // alone, which can make some readers (JAWS/NVDA) announce the change silently.
    <div
      className="rounded border border-white/20 p-2 mx-2 text-sm"
      role="log"
      aria-live="polite"
      data-testid="piquet-declaration-list"
    >
      <div className="mb-1 font-bold">{t('declarationsList')}</div>
      {results.map((r) => {
        const playerLabel =
          r.scoredBy === elderIdx ? t('roleElder') : r.scoredBy === youngerIdx ? t('roleYounger') : '?';
        return (
          <div key={`decl-${r.kind}-${r.winner}-${r.score}`} className="text-xs">
            {declKindLabel(r.kind, t)}:{' '}
            {r.score === 0 ? t('declTied') : t('declScored', { player: playerLabel, score: r.score })}
          </div>
        );
      })}
    </div>
  );
}

function TrickView({ trick }: { trick: PiquetResponse['currentTrick'] }) {
  const { t } = useTranslation('piquet');
  const { cardWidth } = useCardDimensions();
  return (
    <div className="rounded border border-white/20 p-2 mx-2 text-sm">
      <div className="mb-1 font-bold">{t('trickHeader')}</div>
      <div className="flex gap-2">
        {trick.map((tc, i) => (
          <div
            key={`trick-${i}-${tc.playerIdx}-${tc.card.design}-${tc.card.value}`}
            className="flex flex-col items-center gap-0.5"
          >
            <span className="text-xs text-ds-text-muted">P{tc.playerIdx}</span>
            <AnimatedCard card={tc.card} width={cardWidth} />
          </div>
        ))}
      </div>
    </div>
  );
}
