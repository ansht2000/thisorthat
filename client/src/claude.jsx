import { useState } from "react";

// --- Mock data (replace with your backend calls) ---
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

const NAV_LINKS = [
  { label: "Rank", path: "/" },
  { label: "Leaderboard", path: "/leaderboard" },
  { label: "History", path: "/history" },
  { label: "About", path: "/about" },
];

// --- Utility: get initials for placeholder ---
function getInitials(name) {
  return name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

// --- Components ---
function Navbar({ active }) {
  return (
    <nav style={styles.nav}>
      <div style={styles.navInner}>
        <a href="/" style={styles.logo}>
          <span style={styles.logoIcon}>⚔</span>
          <span style={styles.logoText}>RANKR</span>
        </a>
        <div style={styles.navLinks}>
          {NAV_LINKS.map((link) => (
            <a
              key={link.path}
              href={link.path}
              style={{
                ...styles.navLink,
                ...(active === link.path ? styles.navLinkActive : {}),
              }}
            >
              {link.label}
            </a>
          ))}
        </div>
      </div>
    </nav>
  );
}

function CharacterCard({ character, side, onPick }) {
  const isLeft = side === "left";
  return (
    <button
      onClick={() => onPick(character)}
      style={{
        ...styles.card,
        ...(isLeft ? styles.cardLeft : styles.cardRight),
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = "scale(1.03)";
        e.currentTarget.style.boxShadow = isLeft
          ? "0 0 40px rgba(255,60,60,0.5)"
          : "0 0 40px rgba(60,130,255,0.5)";
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = "scale(1)";
        e.currentTarget.style.boxShadow = isLeft
          ? "0 0 20px rgba(255,60,60,0.25)"
          : "0 0 20px rgba(60,130,255,0.25)";
      }}
    >
      <div
        style={{
          ...styles.cardImageWrap,
          background: isLeft
            ? "linear-gradient(135deg, #1a0000 0%, #3d0a0a 100%)"
            : "linear-gradient(135deg, #00091a 0%, #0a1a3d 100%)",
        }}
      >
        {character.img ? (
          <img src={character.img} alt={character.name} style={styles.cardImage} />
        ) : (
          <span style={styles.cardInitials}>{getInitials(character.name)}</span>
        )}
      </div>
      <div style={styles.cardInfo}>
        <span style={styles.cardAlias}>{character.alias}</span>
        <span style={styles.cardName}>{character.name}</span>
      </div>
      <div
        style={{
          ...styles.cardPickLabel,
          background: isLeft
            ? "linear-gradient(90deg, #ff3c3c, #ff6b3c)"
            : "linear-gradient(90deg, #3c8cff, #3cdfff)",
        }}
      >
        PICK
      </div>
    </button>
  );
}

function VsBadge() {
  return (
    <div style={styles.vsBadge}>
      <span style={styles.vsText}>VS</span>
    </div>
  );
}

function FranchiseSelector({ selected, onChange }) {
  return (
    <div style={styles.selectorWrap}>
      {Object.entries(FRANCHISES).map(([key, f]) => (
        <button
          key={key}
          onClick={() => onChange(key)}
          style={{
            ...styles.selectorBtn,
            ...(selected === key ? styles.selectorBtnActive : {}),
          }}
        >
          {f.name}
        </button>
      ))}
    </div>
  );
}

// --- Main Export ---
export default function CharacterRanker() {
  const [franchise, setFranchise] = useState("invincible");
  const [pair, setPair] = useState([0, 1]);
  const [flash, setFlash] = useState(null); // "left" | "right" | null

  const chars = FRANCHISES[franchise].characters;
  const left = chars[pair[0]];
  const right = chars[pair[1]];

  function handlePick(picked) {
    const side = picked.id === left.id ? "left" : "right";
    setFlash(side);

    // TODO: Send pick to your backend here
    // e.g. fetch('/api/rank', { method: 'POST', body: JSON.stringify({ winner: picked.id, loser: ... }) })

    setTimeout(() => {
      setFlash(null);
      // Advance to next random pair
      let a, b;
      do {
        a = Math.floor(Math.random() * chars.length);
        b = Math.floor(Math.random() * chars.length);
      } while (a === b);
      setPair([a, b]);
    }, 500);
  }

  function handleFranchiseChange(key) {
    setFranchise(key);
    setPair([0, 1]);
    setFlash(null);
  }

  return (
    <div style={styles.page}>
      <Navbar active="/" />

      <main style={styles.main}>
        <div style={styles.header}>
          <h1 style={styles.title}>Who wins?</h1>
          <p style={styles.subtitle}>Pick the stronger character</p>
        </div>

        <FranchiseSelector selected={franchise} onChange={handleFranchiseChange} />

        <div style={styles.arena}>
          {/* Flash overlay */}
          {flash && (
            <div
              style={{
                ...styles.flashOverlay,
                background:
                  flash === "left"
                    ? "radial-gradient(ellipse at 30% 50%, rgba(255,60,60,0.25), transparent 70%)"
                    : "radial-gradient(ellipse at 70% 50%, rgba(60,130,255,0.25), transparent 70%)",
              }}
            />
          )}

          <CharacterCard character={left} side="left" onPick={handlePick} />
          <VsBadge />
          <CharacterCard character={right} side="right" onPick={handlePick} />
        </div>

        <p style={styles.hint}>Click a character to pick them as the winner</p>
      </main>
    </div>
  );
}

// --- Styles ---
const styles = {
  page: {
    minHeight: "100vh",
    background: "#0a0a0f",
    color: "#e8e8e8",
    overflow: "hidden",
  },

  // Nav
  nav: {
    position: "sticky",
    top: 0,
    zIndex: 100,
    background: "rgba(10,10,15,0.85)",
    backdropFilter: "blur(12px)",
    borderBottom: "1px solid rgba(255,255,255,0.06)",
  },
  navInner: {
    maxWidth: 1100,
    margin: "0 auto",
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    padding: "0 24px",
    height: 56,
  },
  logo: {
    display: "flex",
    alignItems: "center",
    gap: 8,
    textDecoration: "none",
  },
  logoIcon: { fontSize: 22 },
  logoText: {
    fontSize: 22,
    fontWeight: 800,
    letterSpacing: 3,
    color: "#fff",
    fontFamily: "'Barlow Condensed', 'Impact', sans-serif",
  },
  navLinks: { display: "flex", gap: 4 },
  navLink: {
    padding: "6px 16px",
    fontSize: 14,
    fontWeight: 600,
    letterSpacing: 1.5,
    textTransform: "uppercase",
    color: "rgba(255,255,255,0.45)",
    textDecoration: "none",
    borderRadius: 6,
    transition: "all 0.2s",
  },
  navLinkActive: {
    color: "#fff",
    background: "rgba(255,255,255,0.08)",
  },

  // Main
  main: {
    maxWidth: 1100,
    margin: "0 auto",
    padding: "40px 24px 60px",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
  },
  header: { textAlign: "center", marginBottom: 24 },
  title: {
    fontSize: 48,
    fontWeight: 900,
    letterSpacing: 2,
    textTransform: "uppercase",
    margin: 0,
    background: "linear-gradient(90deg, #ff3c3c, #fff 50%, #3c8cff)",
    WebkitBackgroundClip: "text",
    WebkitTextFillColor: "transparent",
  },
  subtitle: {
    fontSize: 16,
    color: "rgba(255,255,255,0.35)",
    letterSpacing: 3,
    textTransform: "uppercase",
    marginTop: 6,
  },

  // Franchise selector
  selectorWrap: {
    display: "flex",
    gap: 8,
    marginBottom: 40,
    flexWrap: "wrap",
    justifyContent: "center",
  },
  selectorBtn: {
    padding: "8px 20px",
    fontSize: 13,
    fontWeight: 700,
    letterSpacing: 1.5,
    textTransform: "uppercase",
    background: "rgba(255,255,255,0.04)",
    color: "rgba(255,255,255,0.4)",
    border: "1px solid rgba(255,255,255,0.08)",
    borderRadius: 8,
    cursor: "pointer",
    transition: "all 0.2s",
    fontFamily: "inherit",
  },
  selectorBtnActive: {
    background: "rgba(255,255,255,0.1)",
    color: "#fff",
    border: "1px solid rgba(255,255,255,0.2)",
    boxShadow: "0 0 16px rgba(255,255,255,0.05)",
  },

  // Arena
  arena: {
    position: "relative",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    gap: 24,
    width: "100%",
  },
  flashOverlay: {
    position: "absolute",
    inset: 0,
    borderRadius: 20,
    pointerEvents: "none",
    zIndex: 5,
    animation: "fadeOut 0.5s ease-out forwards",
  },

  // Character card
  card: {
    position: "relative",
    flex: "1 1 0",
    maxWidth: 380,
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    background: "rgba(255,255,255,0.03)",
    border: "1px solid rgba(255,255,255,0.07)",
    borderRadius: 16,
    padding: 0,
    cursor: "pointer",
    transition: "transform 0.25s ease, box-shadow 0.25s ease",
    overflow: "hidden",
    fontFamily: "inherit",
    color: "#e8e8e8",
  },
  cardLeft: {
    boxShadow: "0 0 20px rgba(255,60,60,0.25)",
  },
  cardRight: {
    boxShadow: "0 0 20px rgba(60,130,255,0.25)",
  },
  cardImageWrap: {
    width: "100%",
    aspectRatio: "1 / 1",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    overflow: "hidden",
  },
  cardImage: {
    width: "100%",
    height: "100%",
    objectFit: "cover",
  },
  cardInitials: {
    fontSize: 72,
    fontWeight: 900,
    color: "rgba(255,255,255,0.1)",
    letterSpacing: 4,
    userSelect: "none",
  },
  cardInfo: {
    padding: "16px 20px 12px",
    width: "100%",
    textAlign: "center",
  },
  cardAlias: {
    display: "block",
    fontSize: 24,
    fontWeight: 800,
    letterSpacing: 1.5,
    textTransform: "uppercase",
  },
  cardName: {
    display: "block",
    fontSize: 13,
    color: "rgba(255,255,255,0.35)",
    letterSpacing: 1,
    marginTop: 4,
  },
  cardPickLabel: {
    width: "100%",
    padding: "10px 0",
    fontSize: 14,
    fontWeight: 800,
    letterSpacing: 4,
    textAlign: "center",
    color: "#fff",
    textTransform: "uppercase",
  },

  // VS badge
  vsBadge: {
    width: 64,
    height: 64,
    borderRadius: "50%",
    background: "radial-gradient(circle, #1a1a2e 0%, #0a0a0f 100%)",
    border: "2px solid rgba(255,255,255,0.12)",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    flexShrink: 0,
    zIndex: 10,
    boxShadow: "0 0 30px rgba(0,0,0,0.8)",
  },
  vsText: {
    fontSize: 22,
    fontWeight: 900,
    letterSpacing: 2,
    color: "rgba(255,255,255,0.7)",
  },

  hint: {
    marginTop: 32,
    fontSize: 13,
    color: "rgba(255,255,255,0.2)",
    letterSpacing: 2,
    textTransform: "uppercase",
  },
};