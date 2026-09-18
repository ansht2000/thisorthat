import { useEffect, useState } from "react";
import { EmptyState, ErrorState, LoadingState } from "../../../components/status";
import { ApiError } from "../../../data/api";
import { useMatchup } from "../../../hooks/use-matchup";
import type { Character } from "../../../types/character";
import type { Match } from "../../../types/match";
import CharacterCard, { type Side } from "./character_card";
import VSBadge from "./vs_badge";
import "./arena.css";

// how the vote went for this character, if it was part of the vote just cast
function outcomeFor(character: Character, result: Match | null) {
    if (result?.winner_id === character.id) {
        return { won: true, delta: result.winner_elo_after - result.winner_elo_before };
    }
    if (result?.loser_id === character.id) {
        return { won: false, delta: result.loser_elo_after - result.loser_elo_before };
    }
    return undefined;
}

function Arena({ franchiseId }: { franchiseId: string }) {
    const { pair, loading, error, busy, result, notice, vote, retry } = useMatchup(franchiseId);
    const [flash, setFlash] = useState<Side | null>(null);

    function pick(side: Side) {
        if (!pair || busy) return;
        setFlash(side);
        vote(side === "left" ? pair[0] : pair[1]);
    }

    // arrow keys pick a side. no dependency list so the listener always calls the current pick
    useEffect(() => {
        function onKeyDown(event: KeyboardEvent) {
            if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
            if (event.key === "ArrowLeft") pick("left");
            if (event.key === "ArrowRight") pick("right");
        }
        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    });

    if (loading) return <LoadingState label="Finding a matchup" />;
    if (error instanceof ApiError && error.status === 409) {
        return <EmptyState title="Not enough characters" message="This franchise needs at least two characters before it can be ranked." />;
    }
    if (error || !pair) return <ErrorState message={error?.message ?? "No matchup to show."} onRetry={retry} />;

    const [left, right] = pair;
    const winner = result && (result.winner_id === left.id ? left : right);
    return (
        <>
            <div className="arena">
                {flash && (
                    <div
                        className={`flashOverlay ${flash === "left" ? "flashOverlayLeft" : "flashOverlayRight"}`}
                        onAnimationEnd={() => setFlash(null)}
                    />
                )}
                <CharacterCard character={left} side="left" onPick={() => pick("left")} disabled={busy} outcome={outcomeFor(left, result)} />
                <VSBadge />
                <CharacterCard character={right} side="right" onPick={() => pick("right")} disabled={busy} outcome={outcomeFor(right, result)} />
            </div>
            <p className="arenaNotice" role="status">
                {notice}
                {winner && result && (
                    <span className="visuallyHidden">
                        Vote counted. {winner.name} gained {result.winner_elo_after - result.winner_elo_before} points.
                    </span>
                )}
            </p>
        </>
    );
}

export default Arena;
