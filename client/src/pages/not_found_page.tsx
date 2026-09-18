import { Link } from "react-router";
import PageHeader from "../components/page_header";
import { EmptyState } from "../components/status";

function NotFoundPage() {
    return (
        <>
            <PageHeader title="Not found" />
            <EmptyState title="Nothing here" message="That page doesn't exist.">
                <Link to="/" className="stateButton">
                    Back to voting
                </Link>
            </EmptyState>
        </>
    );
}

export default NotFoundPage;
