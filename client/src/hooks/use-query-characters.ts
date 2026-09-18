import { fetchCharacters } from "../data/api";
import type { Character } from "../types/character";
import { useQuery, type QueryResult } from "./use-query";

export function useQueryCharacters(franchiseId: string | undefined): QueryResult<Character[]> {
    // the key is null until there's an id, so the fetcher never runs without one
    return useQuery(franchiseId ? `characters:${franchiseId}` : null, () => fetchCharacters(franchiseId!));
}
