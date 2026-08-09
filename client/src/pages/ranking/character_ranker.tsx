import { useState } from "react";
import type { FranchisesType } from "../../types/franchise";
import type { Character } from "../../types/character";

import "./character_ranker.css"
import Navbar from "../../components/navbar";
import FranchiseSelector from "./components/franchise_selector";
import CharacterCard from "./components/character_card";
import VSBadge from "./components/vs_badge";
import { useQueryFranchises } from "../../hooks/use-query-franchises";

function CharacterRanker() {
    const FRANCHISES: FranchisesType = {
        "invincible": {
            name: "Invincible",
            id: "1",
            characters: [
                { id: "1", name: "Invincible", picture_url: "https://i.imgur.com/8QlLfmj.png", list_id: "1", elo: 1200 },
                { id: "2", name: "Omni-Man", picture_url: "https://i.imgur.com/YFKoaAk.png", list_id: "1", elo: 1200 },
                { id: "3", name: "Rex Splode", picture_url: "", list_id: "1", elo: 1200 },
                { id: "4", name: "Atom Eve", picture_url: "", list_id: "1", elo: 1200 },
            ],
        },
        "mha": {
            name: "My Hero Academia",
            id: "2",
            characters: [
                { id: "10", name: "Deku", picture_url: "", list_id: "1", elo: 1200 },
                { id: "11", name: "Dynamight", picture_url: "", list_id: "1", elo: 1200 },
                { id: "12", name: "Shoto", picture_url: "", list_id: "1, elo: 1200", elo: 1200 },
                { id: "13", name: "All Might", picture_url: "", list_id: "1", elo: 1200 },
            ],
        },
        "onepiece": {
            name: "One Piece",
            id: "3",
            characters: [
                { id: "20", name: "Monkey D. Luffy", picture_url: "", list_id: "1", elo: 1200 },
                { id: "21", name: "Roronoa Zoro", picture_url: "", list_id: "1", elo: 1200 },
                { id: '22', name: "Sanji", picture_url: "", list_id: "1", elo: 1200 },
                { id: "23", name: "Nico Robin", picture_url: "", list_id: "1", elo: 1200 },
            ],
        },
    };

    const [franchise, setFranchise] = useState<string>("invincible");
    const [pair, setPair] = useState<number[]>([0, 1]);
    const [flash, setFlash] = useState<string | null>(null);

    const chars = FRANCHISES[franchise].characters;
    const left = chars[pair[0]];
    const right = chars[pair[1]];

    function handlePick(picked: Character) {
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

    useQueryFranchises();

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
                    <CharacterCard character={left} side="left" onPick={handlePick}/>
                    <VSBadge/>
                    <CharacterCard character={right} side="right" onPick={handlePick}/>
                </div>
                <p className="hint">Click on a character to choose them as the winner</p>
            </main>
        </div>
    );
}

export default CharacterRanker;