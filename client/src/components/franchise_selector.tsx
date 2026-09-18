import type { FranchiseType } from "../types/franchise";
import "./franchise_selector.css";

type Props = {
    franchises: FranchiseType[];
    selectedId: string | undefined;
    onSelect: (id: string) => void;
};

function FranchiseSelector({ franchises, selectedId, onSelect }: Props) {
    return (
        <div className="selectorWrap" role="group" aria-label="Franchise">
            {franchises.map((franchise) => {
                const active = franchise.id === selectedId;
                return (
                    <button
                        key={franchise.id}
                        type="button"
                        onClick={() => onSelect(franchise.id)}
                        aria-pressed={active}
                        className={`selectorBtn ${active ? "selectorBtnActive" : ""}`}
                    >
                        {franchise.name}
                    </button>
                );
            })}
        </div>
    );
}

export default FranchiseSelector;
