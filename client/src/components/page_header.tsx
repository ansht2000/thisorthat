import { useEffect } from "react";
import "./page_header.css";

type Props = {
    title: string;
    subtitle?: string;
};

function PageHeader({ title, subtitle }: Props) {
    useEffect(() => {
        document.title = `${title} · thisorthat`;
    }, [title]);

    return (
        <header className="header">
            <h1 className="title">{title}</h1>
            {subtitle && <p className="subtitle">{subtitle}</p>}
        </header>
    );
}

export default PageHeader;
