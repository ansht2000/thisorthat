import { API_URL } from "../env";
import { type FranchisesType, type FranchiseType } from "../types/franchise";

export async function fetchFranchises(): Promise<FranchisesType> {
    const response = await fetch(`${API_URL}/lists`);
    if (!response.ok) {
        console.log(response);
        throw new Error("error fetching");
    }

    const franchises: FranchiseType[] = await response.json();
    console.log(franchises);
    const franchises_map = {} as FranchisesType;
    for (let franchise of franchises) {
        franchises_map[franchise.name] = franchise;
    }
    return franchises_map;
}