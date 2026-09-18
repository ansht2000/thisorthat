import { fetchLeaderboard } from "../data/api";
import type { LeaderboardEntry } from "../types/character";
import { useQuery, type QueryResult } from "./use-query";

export function useQueryLeaderboard(franchiseId: string): QueryResult<LeaderboardEntry[]> {
    return useQuery(`leaderboard:${franchiseId}`, () => fetchLeaderboard(franchiseId));
}
