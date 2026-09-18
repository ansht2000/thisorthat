import CharacterAvatar from "../../../components/character_avatar";
import type { LeaderboardEntry } from "../../../types/character";

// below this many matchups a rating is still moving fast, so it's marked as not settled yet
const SETTLED_GAMES = 10;

const MEDALS: Record<number, string> = { 1: "rankGold", 2: "rankSilver", 3: "rankBronze" };

function RankRow({ entry }: { entry: LeaderboardEntry }) {
    const settling = entry.games_played < SETTLED_GAMES;
    return (
        <tr className="rankRow">
            <td className="leaderboardRank">
                <span className={`rankBadge ${MEDALS[entry.rank] ?? ""}`}>{entry.rank}</span>
            </td>
            <td>
                <div className="rankCharacter">
                    <CharacterAvatar name={entry.name} pictureUrl={entry.picture_url} size={40} />
                    <span className="rankName">{entry.name}</span>
                    {settling && (
                        <span className="rankSettling" title={`Fewer than ${SETTLED_GAMES} matchups, so this rating is still settling`}>
                            New
                        </span>
                    )}
                </div>
            </td>
            <td className="leaderboardNumber rankElo">{entry.elo}</td>
            <td className="leaderboardNumber leaderboardGames">{entry.games_played}</td>
        </tr>
    );
}

export default RankRow;
