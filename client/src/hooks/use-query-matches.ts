import { fetchMatches } from "../data/api";
import type { Match } from "../types/match";
import { useQuery, type QueryResult } from "./use-query";

export function useQueryMatches(franchiseId: string, limit: number): QueryResult<Match[]> {
    return useQuery(`matches:${franchiseId}:${limit}`, () => fetchMatches(franchiseId, limit));
}
