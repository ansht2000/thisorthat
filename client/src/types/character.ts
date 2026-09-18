export type Character = {
    id: string;
    name: string;
    picture_url: string;
    elo: number;
    games_played: number;
    list_id: string;
};

// ties share a rank, so ranks can go 1, 2, 2, 4
export type LeaderboardEntry = Character & {
    rank: number;
};
