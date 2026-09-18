import { Link } from "react-router";
import PageHeader from "../../components/page_header";
import "./about_page.css";

function AboutPage() {
    return (
        <>
            <PageHeader title="About" subtitle="How the rankings work" />
            <article className="about">
                <section>
                    <h2>Two characters, one vote</h2>
                    <p>
                        Pick a franchise and you'll get two of its characters. Choose the one you think wins. Every vote
                        nudges both ratings, and the <Link to="/leaderboard">leaderboard</Link> sorts everyone by rating.
                    </p>
                </section>

                <section>
                    <h2>Elo ratings</h2>
                    <p>
                        Ratings use the Elo system from chess. Every character starts at 1200, and the gap between two
                        ratings predicts how likely each is to win. Beating someone you were expected to beat barely
                        moves your rating. An upset moves it a lot.
                    </p>
                </section>

                <section>
                    <h2>Tuned for voting</h2>
                    <ul>
                        <li>
                            <strong>New characters move fast.</strong> A character's first votes count for up to four
                            times as much as later ones, so it finds its place quickly. The effect fades as its vote count
                            grows, so an established ranking can't be upended by one vote.
                        </li>
                        <li>
                            <strong>No points leak away.</strong> Rating changes are rounded, not cut off, so when two
                            characters have played equally often, what one gains the other loses exactly.
                        </li>
                        <li>
                            <strong>Every vote counts.</strong> Votes are applied one at a time, so two people voting at
                            the same moment can't overwrite each other.
                        </li>
                    </ul>
                </section>

                <section>
                    <h2>Picking matchups</h2>
                    <p>
                        Matchups aren't random. Characters with fewer votes come up more often, and they're usually paired
                        with someone close to their rating, since a close matchup tells us far more than a lopsided one.
                        You also won't see the same pair twice in a row.
                    </p>
                </section>

                <Link to="/" className="aboutCta">
                    Start voting
                </Link>
            </article>
        </>
    );
}

export default AboutPage;
