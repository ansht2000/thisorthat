import FranchiseScope from "../../components/franchise_scope";
import PageHeader from "../../components/page_header";
import Arena from "./components/arena";
import "./character_ranker.css";

function CharacterRanker() {
    return (
        <>
            <PageHeader title="Who wins?" subtitle="You choose!" />
            {/* keyed so a new franchise starts from a clean slate instead of finishing the old one's vote */}
            <FranchiseScope>{(franchise) => <Arena key={franchise.id} franchiseId={franchise.id} />}</FranchiseScope>
            <p className="hint">
                Tap or click the character you think wins
                <span className="hintKeys"> · or press ← / →</span>
            </p>
        </>
    );
}

export default CharacterRanker;
