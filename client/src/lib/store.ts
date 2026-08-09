import { atom } from "nanostores";
import { type FranchiseType } from "../types/franchise";
import { type FranchisesType } from "../types/franchise";

export const $franchise = atom<FranchiseType>({} as FranchiseType);
export const $franchises = atom<FranchisesType>({});

export function setFranchise(franchise: FranchiseType) {
    $franchise.set(franchise);
}

export function setFranchises(franchises: FranchisesType) {
    $franchises.set(franchises);
}

