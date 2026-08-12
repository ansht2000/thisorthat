import { useState } from "react";
import type { Character } from "../../types/character";

import "./character_ranker.css"
import Navbar from "../../components/navbar";
import FranchiseSelector from "./components/franchise_selector";
import CharacterCard from "./components/character_card";
import VSBadge from "./components/vs_badge";
import { useQueryFranchises } from "../../hooks/use-query-franchises";

function CharacterRanker() {
    const franchises = useQueryFranchises();

    const [franchise, setFranchise] = useState<string>("invincible");
    const [pair, setPair] = useState<number[]>([0, 1]);
    const [flash, setFlash] = useState<string | null>(null);

    const activeFranchise = franchise ?? Object.keys(franchises)[0];
    const chars = franchises[activeFranchise]?.characters ?? [];
    const left = chars[pair[0]];
    const right = chars[pair[1]];

    function handlePick(picked: Character) {
        if (chars.length < 2) return;
        const side = picked.id === left.id ? "left" : "right";
        setFlash(side);

        // TODO: update pick with backend

        setTimeout(() => {
            setFlash(null);
            let a: number;
            let b: number;
            do {
                a = Math.floor(Math.random() * chars.length);
                b = Math.floor(Math.random() * chars.length);
            } while (a === b);
            setPair([a, b]);
        // set a reasonable timeout, 500ms is good
        }, 500)
    }

    function handleFranchiseChange(franchise: string) {
        setFranchise(franchise);
        setPair([0, 1]);
        setFlash(null);
    }

    return (
        <div className="page">
            <Navbar active="/"/>
            <main className="main">
                <div className="header">
                    <h1 className="title">
                        Who wins?
                    </h1>
                    <p className="subtitle">
                       You choose! 
                    </p>
                </div>
                <FranchiseSelector selected={franchise} onChange={handleFranchiseChange}/>
                <div className="arena">
                    {flash && (
                        <div
                            className={`
                                flashOverlay ${flash === "left" ? "flashOverlayLeftGrad" : "flashOverlayRightGrad"}
                            `}
                        />
                    )}
                    {left && right && (
                        <>
                            <CharacterCard character={left} side="left" onPick={handlePick} />
                            <VSBadge />
                            <CharacterCard character={right} side="right" onPick={handlePick} />
                        </>
                    )}
                </div>
                <p className="hint">Click on a character to choose them as the winner</p>
            </main>
        </div>
    );
}

export default CharacterRanker;