import { useState } from "react";
import { getInitials } from "../lib/utils";
import "./character_avatar.css";

type Props = {
    name: string;
    pictureUrl: string;
    // a round avatar this many pixels across. without it the avatar fills its container
    size?: number;
    className?: string;
};

// the character's picture, or their initials when there isn't one or it fails to load.
// always shown next to the name, so it's hidden from screen readers to avoid reading the name twice
function CharacterAvatar({ name, pictureUrl, size, className = "" }: Props) {
    // remembering which url failed, not just that one did, means a new picture gets a fresh try
    const [failedUrl, setFailedUrl] = useState<string | null>(null);
    const showPicture = pictureUrl !== "" && failedUrl !== pictureUrl;

    return (
        <span
            className={`avatar ${size ? "avatarRound" : "avatarFill"} ${className}`}
            style={size ? { width: size, height: size, fontSize: size * 0.38 } : undefined}
            aria-hidden="true"
        >
            {showPicture ? (
                <img src={pictureUrl} alt="" className="avatarImage" onError={() => setFailedUrl(pictureUrl)} />
            ) : (
                <span className="avatarInitials">{getInitials(name)}</span>
            )}
        </span>
    );
}

export default CharacterAvatar;
