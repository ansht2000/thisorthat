import { useStore } from "@nanostores/react"
import { $franchises } from "../lib/store"
import { useEffect } from "react"
import { fetchFranchises } from "../data/api"

export function useQueryFranchises() {
    const franchises = useStore($franchises)

    useEffect(() => {
        fetchFranchises()
    }, [])
}