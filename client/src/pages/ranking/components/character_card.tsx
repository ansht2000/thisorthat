import CharacterAvatar from "../../../components/character_avatar";
import EloDelta from "../../../components/elo_delta";
import type { Character } from "../../../types/character";
import "./character_card.css";

export type Side = "left" | "right";

type Props = {
    character: Character;
    side: Side;
    onPick: () => void;
    disabled?: boolean;
    // set while a vote's result is on screen
    outcome?: { won: boolean; delta: number };
};

function CharacterCard({ character, side, onPick, disabled = false, outcome }: Props) {
    const isLeft = side === "left";
    const outcomeClass = outcome ? (outcome.won ? "cardWon" : "cardLost") : "";

    return (
        <button
            type="button"
            onClick={onPick}
            disabled={disabled}
            className={`card ${isLeft ? "cardLeft" : "cardRight"} ${outcomeClass}`}
            aria-label={`Choose ${character.name}`}
            aria-keyshortcuts={isLeft ? "ArrowLeft" : "ArrowRight"}
        >
            <div className={`cardImageWrap ${isLeft ? "cardImageWrapLeftGrad" : "cardImageWrapRightGrad"}`}>
                <CharacterAvatar name={character.name} pictureUrl={character.picture_url} />
                {outcome && <EloDelta value={outcome.delta} className="cardDelta" />}
            </div>
            <div className="cardInfo">
                <span className="cardName">{character.name}</span>
            </div>
            <div className={`cardPickLabel ${isLeft ? "cardPickLabelLeftGrad" : "cardPickLabelRightGrad"}`}>
                {outcome ? (outcome.won ? "Winner" : " ") : "Choose"}
                {!outcome && (
                    <kbd className="cardShortcut" aria-hidden="true">
                        {isLeft ? "←" : "→"}
                    </kbd>
                )}
            </div>
        </button>
    );
}

export default CharacterCard;
