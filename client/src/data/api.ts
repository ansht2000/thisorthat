import { API_URL } from "../env";
import { type FranchisesType, type FranchiseType } from "../types/franchise";
import { setFranchises } from "../lib/store";

export async function fetchFranchises() {
    const response = await fetch(`${API_URL}/lists`);
    if (!response.ok) {
        throw new Error("error fetching");
    }

    const franchises: FranchiseType[] = await response.json();
    console.log(franchises);
    const franchises_map = {} as FranchisesType;
    for (let franchise of franchises) {
        franchises_map[franchise.name] = franchise;
    }
    setFranchises(franchises_map);
}