// one vote, with both ratings before and after it
export type Match = {
    id: string;
    list_id: string;
    winner_id: string;
    loser_id: string;
    winner_elo_before: number;
    winner_elo_after: number;
    loser_elo_before: number;
    loser_elo_after: number;
    created_at: string;
};
