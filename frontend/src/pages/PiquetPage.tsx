import { useCallback, useState } from 'react';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePiquetGame } from '../hooks/usePiquetGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, PiquetDeclaration, PiquetPlayerData, PiquetResponse } from '../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';

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

const SUIT_GLYPH: Record<string, string> = {
  SPADE: '♠',
  CLOVER: '♣',
  HEART: '♥',
  DIAMOND: '♦',
  JOKER: 'Jk',
};

/** Format a single card into a short string like "K♠". */
function formatCard(card: Card): string {
  const v = card.value;
  let rank: string;
  switch (v) {
    case 1:
      rank = 'A';
      break;
    case 11:
      rank = 'J';
      break;
    case 12:
      rank = 'Q';
      break;
    case 13:
      rank = 'K';
      break;
    default:
      rank = String(v);
  }
  return `${rank}${SUIT_GLYPH[card.design] ?? '?'}`;
}

function declKindLabel(kind: number): string {
  switch (kind) {
    case PiquetDeclarationKind.POINT:
      return 'Point';
    case PiquetDeclarationKind.SEQUENCE:
      return 'Sequence';
    case PiquetDeclarationKind.SET:
      return 'Set';
    default:
      return '?';
  }
}

/** Renders the Piquet game page. */
export const PiquetPage = withTutorial(PiquetPageContent, 'piquet', PIQUET_TUTORIAL_STEPS);

function PiquetPageContent() {
  const { t, tc, confirmOpen, requestConfirm, confirmReset, cancelReset } = useGamePageSetup('piquet');
  const game = usePiquetGame();
  const { state, loading, error, retry } = game;
  const [selectedDiscards, setSelectedDiscards] = useState<number[]>([]);

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

  const elderIdx = state.elderIdx;
  const human = state.players.find((p) => p.isHuman);
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

      <div className="text-sm opacity-80 px-2">
        {t('dealHeader', { deal: state.dealNumber, total: state.dealsPerPartie })}
      </div>

      <PlayerCard label={t('roleElder')} player={state.players[elderIdx]} carteBlanche={state.carteBlanche[elderIdx]} />
      <PlayerCard
        label={t('roleYounger')}
        player={state.players[state.youngerIdx]}
        carteBlanche={state.carteBlanche[state.youngerIdx]}
      />

      {human ? (
        <div data-tutorial="piquet-hand" className="rounded border border-white/20 p-2 mx-2">
          <div className="mb-1 text-xs opacity-70">{t('yourHand')}</div>
          <div className="flex flex-wrap gap-1">
            {human.cards.map((c, i) => {
              const selected = selectedDiscards.includes(i);
              const handlePlay = () => game.handlePlay(i);
              const handleClick = humanCanExchange ? () => toggleDiscard(i) : humanCanPlay ? handlePlay : undefined;
              return (
                <button
                  key={`hand-${i}-${c.design}-${c.value}`}
                  type="button"
                  className={`rounded border px-2 py-1 text-sm font-mono min-h-[44px] min-w-[44px] ${
                    selected ? 'bg-ds-warning/40 border-ds-warning' : 'bg-white/10 border-white/30'
                  }`}
                  onClick={handleClick}
                  disabled={handleClick == null}
                >
                  {formatCard(c)}
                </button>
              );
            })}
          </div>
        </div>
      ) : null}

      <div data-tutorial="piquet-controls" className="flex flex-wrap gap-2 px-2">
        {humanCanExchange && state.exchangeTurn === PiquetExchangeTurn.ELDER ? (
          <button
            type="button"
            onClick={handleElderExchange}
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
            disabled={selectedDiscards.length > 3}
            className={btnPrimary}
          >
            {t('exchangeYounger', { count: selectedDiscards.length })}
          </button>
        ) : null}
        {inDeclarationPhase ? (
          <button type="button" onClick={game.handleResolveDeclaration} className={btnPrimary}>
            {t('advanceDeclaration', { kind: declKindLabel(state.declStage) })}
          </button>
        ) : null}
        {inScorePhase ? (
          <button type="button" onClick={game.handleNextDeal} className={btnSuccess}>
            {t('nextDeal')}
          </button>
        ) : null}
        <GameResetButton
          isGameEnd={inGameEndPhase}
          onReset={game.handleReset}
          requestConfirm={requestConfirm}
          loading={loading}
        />
      </div>

      {state.declResults.length > 0 ? <DeclarationList results={state.declResults} elderIdx={elderIdx} /> : null}

      {inPlayPhase && state.currentTrick.length > 0 ? <TrickView trick={state.currentTrick} /> : null}

      {inGameEndPhase ? (
        <div className="text-lg font-bold px-2">
          {state.winnerIdx === -1 ? t('partieDraw') : t('partieWinner', { idx: state.winnerIdx })}
        </div>
      ) : null}

      <GameFooter />
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
  if (!player) return null;
  return (
    <div className="rounded border border-white/20 p-2 mx-2 text-sm">
      <div className="flex items-center justify-between">
        <span className="font-bold">{label}</span>
        {carteBlanche ? <span className="text-ds-warning">★ carte blanche</span> : null}
      </div>
      <div className="text-xs opacity-80">
        hand: {player.cardCount} | tricks: {player.trickCount} | round: {player.roundScore} | match: {player.matchScore}
      </div>
    </div>
  );
}

function DeclarationList({ results, elderIdx }: { results: PiquetDeclaration[]; elderIdx: number }) {
  const youngerIdx = elderIdx === 0 ? 1 : 0;
  return (
    <div className="rounded border border-white/20 p-2 mx-2 text-sm">
      <div className="mb-1 font-bold">Declarations</div>
      {results.map((r) => (
        <div key={`decl-${r.kind}-${r.winner}-${r.score}`} className="text-xs">
          {declKindLabel(r.kind)}:{' '}
          {r.score === 0
            ? 'tied'
            : `player ${r.scoredBy === elderIdx ? 'E' : r.scoredBy === youngerIdx ? 'Y' : '?'} +${r.score}`}
        </div>
      ))}
    </div>
  );
}

function TrickView({ trick }: { trick: PiquetResponse['currentTrick'] }) {
  return (
    <div className="rounded border border-white/20 p-2 mx-2 text-sm">
      <div className="mb-1 font-bold">Trick</div>
      <div className="flex gap-2">
        {trick.map((tc, i) => (
          <div
            key={`trick-${i}-${tc.playerIdx}-${tc.card.design}-${tc.card.value}`}
            className="rounded bg-white/10 px-2 py-1 font-mono"
          >
            P{tc.playerIdx}: {formatCard(tc.card)}
          </div>
        ))}
      </div>
    </div>
  );
}
