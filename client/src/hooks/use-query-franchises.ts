import { fetchFranchises } from "../data/api";
import type { FranchiseType } from "../types/franchise";
import { useQuery, type QueryResult } from "./use-query";

export function useQueryFranchises(): QueryResult<FranchiseType[]> {
    return useQuery("franchises", fetchFranchises);
}
