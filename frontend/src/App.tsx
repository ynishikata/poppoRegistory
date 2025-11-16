import { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { apiLogout, apiMe, User } from "./api";
import { AuthForms } from "./components/AuthForms";
import { PlushiesPage } from "./components/PlushiesPage";
import { PlushieDetail } from "./components/PlushieDetail";

export function App() {
  const [user, setUser] = useState<User | null>(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    async function run() {
      try {
        const me = await apiMe();
        setUser(me);
      } catch {
        // not logged in is fine
      } finally {
        setChecking(false);
      }
    }
    void run();
  }, []);

  async function handleLogout() {
    await apiLogout();
    setUser(null);
  }

  return (
    <BrowserRouter>
      <div className="app-root">
        <div className="app-card">
          <header className="app-header">
            <div className="app-title">
              <h1>ぬいぐるみレジストリ</h1>
              <span>家にいるぬいぐるみたちを、名前とお迎え日ごとに記録する小さなWebアプリ。</span>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap" }}>
              <span className="pill">
                <span>Go + React</span>
              </span>
              {user ? (
                <div className="user-chip">
                  <span>●</span>
                  <span>{user.email}</span>
                  <button type="button" className="btn btn-ghost" onClick={handleLogout}>
                    ログアウト
                  </button>
                </div>
              ) : (
                <span className="panel-subtitle">ログインすると自分専用の一覧が作られます。</span>
              )}
            </div>
          </header>

          {checking ? (
            <div className="helper">セッションを確認中です...</div>
          ) : user ? (
            <Routes>
              <Route path="/" element={<PlushiesPage />} />
              <Route path="/plushies/:id" element={<PlushieDetail />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          ) : (
            <AuthForms onAuthenticated={setUser} />
          )}
        </div>
      </div>
    </BrowserRouter>
  );
}


