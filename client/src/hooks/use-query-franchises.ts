import { useStore } from "@nanostores/react"
import { $franchises, setFranchises } from "../lib/store"
import { useEffect } from "react"
import { fetchFranchises } from "../data/api"
import type { FranchisesType } from "../types/franchise"

export function useQueryFranchises(): FranchisesType {
    const franchises = useStore($franchises);

    useEffect(() => {
        fetchFranchises()
            .then(setFranchises)
            .catch((error) => console.log(error))
    }, []);

    return franchises;
}