import type { LeaderboardEntry } from "../../../types/character";
import RankRow from "./rank_row";
import "./leaderboard_table.css";

function LeaderboardTable({ entries }: { entries: LeaderboardEntry[] }) {
    return (
        <table className="leaderboard">
            <thead>
                <tr>
                    <th scope="col" className="leaderboardRank">
                        <span className="visuallyHidden">Rank</span>
                        <span aria-hidden="true">#</span>
                    </th>
                    <th scope="col">Character</th>
                    <th scope="col" className="leaderboardNumber">Rating</th>
                    <th scope="col" className="leaderboardNumber leaderboardGames">Matchups</th>
                </tr>
            </thead>
            <tbody>
                {entries.map((entry) => (
                    <RankRow key={entry.id} entry={entry} />
                ))}
            </tbody>
        </table>
    );
}

export default LeaderboardTable;
