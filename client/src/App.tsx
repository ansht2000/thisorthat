import { BrowserRouter, Route, Routes } from "react-router";
import AppShell from "./components/app_shell";
import AboutPage from "./pages/about/about_page";
import HistoryPage from "./pages/history/history_page";
import LeaderboardPage from "./pages/leaderboard/leaderboard_page";
import NotFoundPage from "./pages/not_found_page";
import CharacterRanker from "./pages/ranking/character_ranker";

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route element={<AppShell />}>
                    <Route index element={<CharacterRanker />} />
                    <Route path="leaderboard" element={<LeaderboardPage />} />
                    <Route path="history" element={<HistoryPage />} />
                    <Route path="about" element={<AboutPage />} />
                    <Route path="*" element={<NotFoundPage />} />
                </Route>
            </Routes>
        </BrowserRouter>
    );
}

export default App;
