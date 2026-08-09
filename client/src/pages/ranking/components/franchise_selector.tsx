import "./franchise_selector.css"

function FranchiseSelector({ selected, onChange }: {selected: string, onChange: (key: string) => void}) {
    const FRANCHISES = {
    invincible: {
        name: "Invincible",
        characters: [
        { id: 1, name: "Mark Grayson", alias: "Invincible", img: "https://i.imgur.com/8QlLfmj.png" },
        { id: 2, name: "Nolan Grayson", alias: "Omni-Man", img: "https://i.imgur.com/YFKoaAk.png" },
        { id: 3, name: "Rex Splode", alias: "Rex Splode", img: "" },
        { id: 4, name: "Atom Eve", alias: "Atom Eve", img: "" },
        ],
    },
    mha: {
        name: "My Hero Academia",
        characters: [
        { id: 10, name: "Izuku Midoriya", alias: "Deku", img: "" },
        { id: 11, name: "Katsuki Bakugo", alias: "Dynamight", img: "" },
        { id: 12, name: "Shoto Todoroki", alias: "Shoto", img: "" },
        { id: 13, name: "All Might", alias: "All Might", img: "" },
        ],
    },
    onepiece: {
        name: "One Piece",
        characters: [
        { id: 20, name: "Monkey D. Luffy", alias: "Straw Hat", img: "" },
        { id: 21, name: "Roronoa Zoro", alias: "Pirate Hunter", img: "" },
        { id: 22, name: "Sanji", alias: "Black Leg", img: "" },
        { id: 23, name: "Nico Robin", alias: "Devil Child", img: "" },
        ],
    },
    };

    return (
        <div className="selectorWrap">
            {Object.entries(FRANCHISES).map(([key, f]) => (
                <button
                    key={key}
                    onClick={() => onChange(key)}
                    className={`selectorBtn ${selected === key ? "selectorBtnActive" : ""}`}
                >
                    {f.name}
                </button>
            ))}
        </div>
    )
}

export default FranchiseSelector;