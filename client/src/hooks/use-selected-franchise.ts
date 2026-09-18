import { useStore } from "@nanostores/react";
import { $selectedFranchiseId, setSelectedFranchiseId } from "../lib/store";
import { useQueryFranchises } from "./use-query-franchises";

export function useSelectedFranchise() {
    const franchises = useQueryFranchises();
    const selectedId = useStore($selectedFranchiseId);
    const list = franchises.data ?? [];

    return {
        franchises: list,
        // the first franchise until one is picked, or if the picked one has since been deleted
        selected: list.find((franchise) => franchise.id === selectedId) ?? list[0],
        select: setSelectedFranchiseId,
        loading: franchises.loading,
        error: franchises.error,
        refetch: franchises.refetch,
    };
}
