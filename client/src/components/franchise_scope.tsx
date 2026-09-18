import type { ReactNode } from "react";
import { useSelectedFranchise } from "../hooks/use-selected-franchise";
import type { FranchiseType } from "../types/franchise";
import FranchiseSelector from "./franchise_selector";
import { EmptyState, ErrorState, LoadingState } from "./status";

type Props = {
    // what to show for the picked franchise
    children: (franchise: FranchiseType) => ReactNode;
};

// the franchise picker every per franchise page starts with. it covers loading, failing,
// and there being no franchises at all, so pages only deal with the franchise they're given
function FranchiseScope({ children }: Props) {
    const { franchises, selected, select, loading, error, refetch } = useSelectedFranchise();

    if (!selected) {
        if (loading) return <LoadingState label="Loading franchises" />;
        if (error) return <ErrorState message={error.message} onRetry={refetch} />;
        return <EmptyState title="No franchises yet" message="Once some are added they'll show up here." />;
    }
    return (
        <>
            <FranchiseSelector franchises={franchises} selectedId={selected.id} onSelect={select} />
            {children(selected)}
        </>
    );
}

export default FranchiseScope;
