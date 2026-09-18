import { NavLink } from "react-router";
import "./navbar.css";

const NAV_LINKS = [
    { label: "Rank", path: "/" },
    { label: "Leaderboard", path: "/leaderboard" },
    { label: "History", path: "/history" },
    { label: "About", path: "/about" },
];

function Navbar() {
    return (
        <nav className="nav" aria-label="Main">
            <div className="navInner">
                <NavLink to="/" className="logo" aria-label="thisorthat home">
                    <span className="logoThis">this</span>
                    <span className="logoOr">or</span>
                    <span className="logoThat">that</span>
                </NavLink>
                <div className="navLinks">
                    {NAV_LINKS.map((link) => (
                        <NavLink
                            key={link.path}
                            to={link.path}
                            // without end, "/" would count as active on every page
                            end={link.path === "/"}
                            className={({ isActive }) => `navLink ${isActive ? "navLinkActive" : ""}`}
                        >
                            {link.label}
                        </NavLink>
                    ))}
                </div>
            </div>
        </nav>
    );
}

export default Navbar;
