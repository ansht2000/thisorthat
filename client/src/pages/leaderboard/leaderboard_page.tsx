import { Link } from "react-router";
import FranchiseScope from "../../components/franchise_scope";
import PageHeader from "../../components/page_header";
import { EmptyState, ErrorState, LoadingState } from "../../components/status";
import { useQueryLeaderboard } from "../../hooks/use-query-leaderboard";
import LeaderboardTable from "./components/leaderboard_table";

function Leaderboard({ franchiseId }: { franchiseId: string }) {
    const { data, error, refetch } = useQueryLeaderboard(franchiseId);

    if (!data) {
        return error ? <ErrorState message={error.message} onRetry={refetch} /> : <LoadingState label="Loading rankings" />;
    }
    if (data.length === 0) {
        return <EmptyState title="No characters yet" message="This franchise doesn't have anyone to rank yet." />;
    }
    return (
        <>
            <LeaderboardTable entries={data} />
            <p className="leaderboardFootnote">
                Every character starts at 1200. <Link to="/">Cast a few votes</Link> to move them around.
            </p>
        </>
    );
}

function LeaderboardPage() {
    return (
        <>
            <PageHeader title="Leaderboard" subtitle="Ranked by every vote" />
            <FranchiseScope>{(franchise) => <Leaderboard franchiseId={franchise.id} />}</FranchiseScope>
        </>
    );
}

export default LeaderboardPage;
