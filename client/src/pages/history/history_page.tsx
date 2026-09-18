import { Link } from "react-router";
import FranchiseScope from "../../components/franchise_scope";
import PageHeader from "../../components/page_header";
import { EmptyState, ErrorState, LoadingState } from "../../components/status";
import { useQueryCharacters } from "../../hooks/use-query-characters";
import { useQueryMatches } from "../../hooks/use-query-matches";
import MatchRow from "./components/match_row";
import "./history_page.css";

const HISTORY_LIMIT = 50;

function History({ franchiseId }: { franchiseId: string }) {
    const matches = useQueryMatches(franchiseId, HISTORY_LIMIT);
    // matches only carry ids, this fills in names and pictures
    const characters = useQueryCharacters(franchiseId);

    if (!matches.data || !characters.data) {
        const error = matches.error ?? characters.error;
        if (error) {
            return (
                <ErrorState
                    message={error.message}
                    onRetry={() => {
                        matches.refetch();
                        characters.refetch();
                    }}
                />
            );
        }
        return <LoadingState label="Loading recent votes" />;
    }
    if (matches.data.length === 0) {
        return (
            <EmptyState title="No votes yet" message="Nobody has voted on this franchise yet.">
                <Link to="/" className="historyCta">
                    Cast the first vote
                </Link>
            </EmptyState>
        );
    }

    const byId = new Map(characters.data.map((character) => [character.id, character]));
    return (
        <ol className="historyList">
            {matches.data.map((match) => (
                <MatchRow key={match.id} match={match} winner={byId.get(match.winner_id)} loser={byId.get(match.loser_id)} />
            ))}
        </ol>
    );
}

function HistoryPage() {
    return (
        <>
            <PageHeader title="History" subtitle={`The last ${HISTORY_LIMIT} votes`} />
            <FranchiseScope>{(franchise) => <History franchiseId={franchise.id} />}</FranchiseScope>
        </>
    );
}

export default HistoryPage;
