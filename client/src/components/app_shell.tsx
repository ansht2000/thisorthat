import { useEffect } from "react";
import { Outlet, useLocation } from "react-router";
import Navbar from "./navbar";
import "./app_shell.css";

// the frame every page renders inside, so the navbar stays put between pages
function AppShell() {
    const { pathname } = useLocation();
    // client side navigation keeps the old scroll position otherwise
    useEffect(() => {
        window.scrollTo(0, 0);
    }, [pathname]);

    return (
        <div className="page">
            <Navbar />
            <main className="main">
                <Outlet />
            </main>
        </div>
    );
}

export default AppShell;
