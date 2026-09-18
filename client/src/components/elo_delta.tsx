import "./elo_delta.css";

type Props = {
    value: number;
    className?: string;
};

// a rating change like +29 or −29, in green or red. font size comes from where it's used
function EloDelta({ value, className = "" }: Props) {
    const gained = value >= 0;
    return (
        <span
            className={`eloDelta ${gained ? "eloDeltaGain" : "eloDeltaLoss"} ${className}`}
            aria-label={`${gained ? "gained" : "lost"} ${Math.abs(value)} rating points`}
        >
            {/* a real minus sign, the hyphen is noticeably shorter than the plus */}
            {gained ? "+" : "−"}
            {Math.abs(value)}
        </span>
    );
}

export default EloDelta;
