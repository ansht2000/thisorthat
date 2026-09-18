import { atom } from "nanostores";

// shared by every page so switching pages keeps the same franchise.
// null until someone picks one, which shows the first franchise
export const $selectedFranchiseId = atom<string | null>(null);

export function setSelectedFranchiseId(id: string) {
    $selectedFranchiseId.set(id);
}
