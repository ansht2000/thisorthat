import { useEffect, useRef, useState } from "react";
import { ApiError, fetchMatchup, postMatch } from "../data/api";
import type { Character } from "../types/character";
import type { Match } from "../types/match";
import { toError } from "./use-query";

// how long the rating change stays on the cards before the next pair replaces them
const RESULT_MS = 900;

function wait(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

function voteErrorMessage(error: unknown): string {
    if (error instanceof ApiError && error.status === 429) {
        return `Slow down a little, you can vote again in ${error.retryAfter ?? 1}s.`;
    }
    return `Your vote didn't go through: ${toError(error).message}`;
}

// runs voting for one franchise: loads a pair, sends the vote, shows the result,
// then loads the next pair while skipping the one just seen.
// mount a fresh one per franchise (key it on the franchise id) so nothing carries over
export function useMatchup(franchiseId: string) {
    const [pair, setPair] = useState<[Character, Character] | null>(null);
    // stops voting entirely: no pair could be loaded
    const [error, setError] = useState<Error | null>(null);
    // the vote just cast, shown on the cards until the next pair arrives
    const [result, setResult] = useState<Match | null>(null);
    // a problem with the last vote that doesn't stop the next one, like being rate limited
    const [notice, setNotice] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);
    // bumped by retry to load again after an error
    const [attempt, setAttempt] = useState(0);

    // the busy state only updates on the next render, so a double click
    // or a held down arrow key could get two votes in before it does
    const busyRef = useRef(false);
    const lastPair = useRef<[string, string] | undefined>(undefined);

    useEffect(() => {
        let cancelled = false;
        fetchMatchup(franchiseId, lastPair.current)
            .then((next) => {
                if (!cancelled) setPair(next);
            })
            .catch((error: unknown) => {
                if (!cancelled) setError(toError(error));
            });
        return () => {
            cancelled = true;
        };
    }, [franchiseId, attempt]);

    async function vote(winner: Character) {
        if (!pair || busyRef.current) return;
        const loser = winner.id === pair[0].id ? pair[1] : pair[0];
        busyRef.current = true;
        setBusy(true);
        setNotice(null);

        try {
            let match: Match;
            try {
                match = await postMatch(winner.id, loser.id);
            } catch (error) {
                // the vote didn't count, so leave the same pair up to try again
                setNotice(voteErrorMessage(error));
                return;
            }

            setResult(match);
            lastPair.current = [pair[0].id, pair[1].id];
            try {
                // load the next pair while the result is on screen, but keep it up for at least RESULT_MS
                const [next] = await Promise.all([fetchMatchup(franchiseId, lastPair.current), wait(RESULT_MS)]);
                setPair(next);
            } catch (error) {
                setPair(null);
                setError(toError(error));
            } finally {
                setResult(null);
            }
        } finally {
            busyRef.current = false;
            setBusy(false);
        }
    }

    function retry() {
        setError(null);
        setAttempt((a) => a + 1);
    }

    return {
        pair,
        loading: pair === null && error === null,
        error,
        busy,
        result,
        notice,
        vote,
        retry,
    };
}
