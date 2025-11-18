import { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { supabase } from "./supabase";
import { apiLogout, User } from "./api";
import { AuthForms } from "./components/AuthForms";
import { PlushiesPage } from "./components/PlushiesPage";
import { PlushieDetail } from "./components/PlushieDetail";

export function App() {
  const [user, setUser] = useState<User | null>(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    // Check initial session
    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session?.user) {
        setUser({
          id: session.user.id,
          email: session.user.email || "",
        });
      }
      setChecking(false);
    });

    // Listen for auth changes
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, session) => {
      if (session?.user) {
        setUser({
          id: session.user.id,
          email: session.user.email || "",
        });
      } else {
        setUser(null);
      }
      setChecking(false);
    });

    return () => subscription.unsubscribe();
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


