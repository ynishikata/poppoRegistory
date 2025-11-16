import { FormEvent, useState } from "react";
import { apiLogin, apiRegister, User } from "../api";

type Props = {
  onAuthenticated: (user: User) => void;
};

export function AuthForms({ onAuthenticated }: Props) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const fn = mode === "login" ? apiLogin : apiRegister;
      const user = await fn(email.trim(), password);
      onAuthenticated(user);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("エラーが発生しました");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="panel">
      <div className="panel-header">
        <div>
          <div className="panel-title">サインイン</div>
          <div className="panel-subtitle">
            メールアドレスとパスワードでログイン / 新規登録します。
          </div>
        </div>
        <div className="auth-tabs">
          <button
            type="button"
            className={`auth-tab ${mode === "login" ? "active" : ""}`}
            onClick={() => setMode("login")}
          >
            ログイン
          </button>
          <button
            type="button"
            className={`auth-tab ${mode === "register" ? "active" : ""}`}
            onClick={() => setMode("register")}
          >
            新規登録
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="email">メールアドレス</label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className="field">
          <label htmlFor="password">パスワード</label>
          <input
            id="password"
            type="password"
            autoComplete={mode === "login" ? "current-password" : "new-password"}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={4}
          />
          <div className="helper">とりあえずの簡単なパスワードでOKです（開発用）。</div>
        </div>

        {error && <div className="error-text">{error}</div>}

        <button className="btn btn-primary" type="submit" disabled={loading}>
          {loading ? "送信中..." : mode === "login" ? "ログインする" : "登録してログイン"}
        </button>
      </form>
    </div>
  );
}


