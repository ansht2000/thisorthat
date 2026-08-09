import type { Character } from "../../../types/character";
import { getInitials } from "../../../lib/utils";

import "./character_card.css"

function CharacterCard({ character, side, onPick }: { character: Character, side: string, onPick: (character: Character) => void }) {
    const isLeft = side === "left";
    return (
        <button
            onClick={() => onPick(character)}
            className={`card ${isLeft ? "cardLeft": "cardRight"}`}
            onMouseEnter={(e) => {
                e.currentTarget.style.transform="scale(1.03)";
                e.currentTarget.style.boxShadow=`
                    0 0 40px ${isLeft
                        ? "rgba(255,60,60,0.5)"
                        : "rgba(60,130,255,0.5)"
                    }
                `
            }}
            onMouseLeave={(e) => {
                e.currentTarget.style.transform = "scale(1)";
                e.currentTarget.style.boxShadow=`
                    0 0 20px ${isLeft
                        ? "rgba(255,60,60,0.25)"
                        : "rgba(60,130,255,0.25)"
                    }
                `
            }}
        >
            <div
                className={`
                    cardImageWrap ${isLeft ? "cardImageWrapLeftGrad" : "cardImageWrapRightGrad"}
                `}
            >
                {character.picture_url ? (
                    <img src={character.picture_url} alt={character.name} className="cardImage"/>
                ) : (
                    <span className="cardInitials">{getInitials(character.name)}</span>
                )}
            </div>
            <div className="cardInfo">
                <span className="cardName">{character.name}</span>
            </div>
            <div
                className={`
                    cardPickLabel ${isLeft ? "cardPickLabelLeftGrad" : "cardPickLabelRightGrad"}
                `}
            >
                CHOOSE
            </div>
        </button>
    )
}

export default CharacterCard;