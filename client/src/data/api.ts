import { API_URL } from "../env";
import type { Character, LeaderboardEntry } from "../types/character";
import type { FranchiseType } from "../types/franchise";
import type { Match } from "../types/match";

export class ApiError extends Error {
    // 0 when the request never got a response at all
    status: number;
    // seconds the server asked us to wait, sent along with a 429
    retryAfter: number | null;

    constructor(message: string, status: number, retryAfter: number | null = null) {
        super(message);
        this.name = "ApiError";
        this.status = status;
        this.retryAfter = retryAfter;
    }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
    let response: Response;
    try {
        response = await fetch(`${API_URL}${path}`, init);
    } catch {
        // fetch only throws when there's no response, like the server being down or the network dropping
        throw new ApiError("Couldn't reach the server. Check your connection and try again.", 0);
    }

    if (!response.ok) {
        const body = await response.json().catch(() => null);
        const retryAfter = Number(response.headers.get("Retry-After")) || null;
        // the api always sends { "error": "..." }. anything else means something other
        // than the api answered, like a different program on the same port, so say where we asked
        const message = body?.error ?? `${response.status} from ${API_URL}${path}, which doesn't look like the thisorthat API.`;
        throw new ApiError(message, response.status, retryAfter);
    }
    return response.json();
}

export function fetchFranchises(): Promise<FranchiseType[]> {
    return request("/lists");
}

export function fetchCharacters(franchiseId: string): Promise<Character[]> {
    return request(`/lists/${franchiseId}/characters`);
}

export function fetchLeaderboard(franchiseId: string): Promise<LeaderboardEntry[]> {
    return request(`/lists/${franchiseId}/leaderboard`);
}

// newest first
export function fetchMatches(franchiseId: string, limit: number): Promise<Match[]> {
    return request(`/lists/${franchiseId}/matches?limit=${limit}`);
}

// exclude is the pair the voter just saw, so the server doesn't hand it straight back
export function fetchMatchup(franchiseId: string, exclude?: [string, string]): Promise<[Character, Character]> {
    const query = exclude ? `?${new URLSearchParams({ exclude: exclude.join(",") })}` : "";
    return request(`/lists/${franchiseId}/matchup${query}`);
}

export function postMatch(winnerId: string, loserId: string): Promise<Match> {
    return request("/matches", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ winner_id: winnerId, loser_id: loserId }),
    });
}
