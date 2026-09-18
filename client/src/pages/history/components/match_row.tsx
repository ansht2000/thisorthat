import CharacterAvatar from "../../../components/character_avatar";
import EloDelta from "../../../components/elo_delta";
import { formatTimeAgo } from "../../../lib/utils";
import type { Character } from "../../../types/character";
import type { Match } from "../../../types/match";
import "./match_row.css";

type Props = {
    match: Match;
    // undefined if the character couldn't be found, which shouldn't happen since deleting one deletes its matches
    winner: Character | undefined;
    loser: Character | undefined;
};

function Contender({ character, delta }: { character: Character | undefined; delta: number }) {
    const name = character?.name ?? "Unknown";
    return (
        <div className="matchContender">
            <CharacterAvatar name={name} pictureUrl={character?.picture_url ?? ""} size={36} />
            <span className="matchName">{name}</span>
            <EloDelta value={delta} className="matchDelta" />
        </div>
    );
}

function MatchRow({ match, winner, loser }: Props) {
    return (
        <li className="matchRow">
            <Contender character={winner} delta={match.winner_elo_after - match.winner_elo_before} />
            <span className="matchVerb">beat</span>
            <Contender character={loser} delta={match.loser_elo_after - match.loser_elo_before} />
            <time className="matchTime" dateTime={match.created_at} title={new Date(match.created_at).toLocaleString()}>
                {formatTimeAgo(match.created_at)}
            </time>
        </li>
    );
}

export default MatchRow;
