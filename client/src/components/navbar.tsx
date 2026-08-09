import './navbar.css'

function Navbar({ active }: { active: string }) {
    const NAV_LINKS = [
        { label: "Rank", path: "/" },
        { label: "Leaderboard", path: "/leaderboard" },
        { label: "History", path: "/history" },
        { label: "About", path: "/about" },
    ];

    return (
        <nav className="nav">
            <div className="navInner">
                <a href="/" className="logo">
                    <span className="logoText">thisorthat</span>
                </a>
                <div className="navLinks">
                    {NAV_LINKS.map((link) => (
                        <a
                            key={link.path}
                            href={link.path}
                            className={`navlink ${active === link.path ? "navLinkActive" : ""}`}
                        >
                            {link.label}
                        </a>
                    ))}
                </div>
            </div>
        </nav>
    )
}

export default Navbar