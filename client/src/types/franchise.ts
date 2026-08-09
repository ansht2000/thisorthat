import type { Character } from "./character.ts";

export type FranchiseType = {
    id: string;
    name: string;
    characters: Character[];
}

export type FranchisesType = {
    [key: string]: FranchiseType;
}

