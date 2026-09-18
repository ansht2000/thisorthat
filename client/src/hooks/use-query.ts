import { useEffect, useRef, useState } from "react";

export type QueryResult<T> = {
    // the latest response, or the cached one from last time while a fresh one loads
    data: T | undefined;
    // waiting on a response for the current key
    loading: boolean;
    error: Error | null;
    refetch: () => void;
};

// the last response for each key, so going back to a page shows what it had
// right away instead of a spinner, then swaps in the fresh response
const cache = new Map<string, unknown>();

// fetches whenever key changes, and key should cover everything fetcher depends on.
// a null key means there's nothing to fetch yet, like characters before a franchise loads
export function useQuery<T>(key: string | null, fetcher: () => Promise<T>): QueryResult<T> {
    const [version, setVersion] = useState(0);
    // tagged with the key and version it answers, so a response for a franchise
    // you've already clicked away from never shows up under the new one
    const [result, setResult] = useState<{ key: string; version: number; data?: T; error?: Error } | null>(null);

    // the latest fetcher, so the effect below only reruns when key does
    const fetcherRef = useRef(fetcher);
    useEffect(() => {
        fetcherRef.current = fetcher;
    });

    useEffect(() => {
        if (key === null) return;
        let cancelled = false;
        fetcherRef.current()
            .then((data) => {
                cache.set(key, data);
                if (!cancelled) setResult({ key, version, data });
            })
            .catch((error: unknown) => {
                if (!cancelled) setResult({ key, version, error: toError(error) });
            });
        return () => {
            cancelled = true;
        };
    }, [key, version]);

    const settled = result?.key === key && result.version === version ? result : null;
    return {
        data: settled?.data ?? (key === null ? undefined : (cache.get(key) as T | undefined)),
        loading: key !== null && settled === null,
        error: settled?.error ?? null,
        refetch: () => setVersion((v) => v + 1),
    };
}

export function toError(error: unknown): Error {
    return error instanceof Error ? error : new Error(String(error));
}
