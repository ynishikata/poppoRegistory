import { FormEvent, useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { apiGetPlushie, apiUpdateConversation, apiChat, Plushie } from "../api";

export function PlushieDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [plushie, setPlushie] = useState<Plushie | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [conversationHistory, setConversationHistory] = useState("");
  const [savingHistory, setSavingHistory] = useState(false);
  const [chatMessage, setChatMessage] = useState<string | null>(null);
  const [chatting, setChatting] = useState(false);

  useEffect(() => {
    if (!id) {
      setError("Invalid ID");
      setLoading(false);
      return;
    }
    void load();
  }, [id]);

  async function load() {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const p = await apiGetPlushie(Number(id));
      setPlushie(p);
      setConversationHistory(p.conversation_history || "");
    } catch (err) {
      if (err instanceof Error) setError(err.message);
      else setError("読み込みに失敗しました");
    } finally {
      setLoading(false);
    }
  }

  async function handleSaveHistory(e: FormEvent) {
    e.preventDefault();
    if (!id) return;
    setSavingHistory(true);
    setError(null);
    try {
      await apiUpdateConversation(Number(id), conversationHistory);
      // reload to get updated data
      await load();
    } catch (err) {
      if (err instanceof Error) setError(err.message);
      else setError("保存に失敗しました");
    } finally {
      setSavingHistory(false);
    }
  }

  async function handleChat() {
    if (!id) return;
    setChatting(true);
    setChatMessage(null);
    setError(null);
    try {
      const result = await apiChat(Number(id));
      setChatMessage(result.message);
    } catch (err) {
      if (err instanceof Error) setError(err.message);
      else setError("会話の生成に失敗しました");
    } finally {
      setChatting(false);
    }
  }

  if (loading) {
    return (
      <div className="content-grid">
        <div className="helper">読み込み中...</div>
      </div>
    );
  }

  if (error && !plushie) {
    return (
      <div className="content-grid">
        <div className="error-text">{error}</div>
        <button className="btn btn-primary" onClick={() => navigate("/")}>
          一覧に戻る
        </button>
      </div>
    );
  }

  if (!plushie) {
    return (
      <div className="content-grid">
        <div className="error-text">ぬいぐるみが見つかりません</div>
        <button className="btn btn-primary" onClick={() => navigate("/")}>
          一覧に戻る
        </button>
      </div>
    );
  }

  return (
    <div className="content-grid">
      <div>
        <div className="panel">
          <div className="panel-header">
            <div>
              <div className="panel-title">{plushie.name}の詳細</div>
              <div className="panel-subtitle">種類: {plushie.kind}</div>
            </div>
            <button className="btn btn-ghost" onClick={() => navigate("/")}>
              一覧に戻る
            </button>
          </div>

          {plushie.image_url && (
            <div style={{ marginBottom: "1rem" }}>
              <img
                src={plushie.image_url}
                alt={plushie.name}
                style={{ maxWidth: "100%", borderRadius: "8px" }}
              />
            </div>
          )}

          <div className="field">
            <label>お迎え日</label>
            <div>{plushie.adopted_at || "-"}</div>
          </div>

          <div className="field">
            <label>作成日</label>
            <div>{plushie.created_at ? new Date(plushie.created_at).toLocaleDateString("ja-JP") : "-"}</div>
          </div>
        </div>
      </div>

      <div>
        <div className="panel">
          <div className="panel-header">
            <div>
              <div className="panel-title">会話履歴</div>
              <div className="panel-subtitle">このぬいぐるみが過去に話した内容を記録します</div>
            </div>
          </div>

          <form onSubmit={handleSaveHistory}>
            <div className="field">
              <label htmlFor="conversation_history">会話履歴</label>
              <textarea
                id="conversation_history"
                value={conversationHistory}
                onChange={(e) => setConversationHistory(e.target.value)}
                rows={10}
                placeholder="例: こんにちは！今日はいい天気ですね。一緒に遊びましょう。"
                style={{ width: "100%", padding: "0.5rem", borderRadius: "4px", border: "1px solid #ccc" }}
              />
              <div className="helper">
                このぬいぐるみの性格や過去の会話内容を自由に記録してください。LLMがこの内容を参考に「一言」を生成します。
              </div>
            </div>

            {error && <div className="error-text">{error}</div>}

            <button className="btn btn-primary" type="submit" disabled={savingHistory}>
              {savingHistory ? "保存中..." : "会話履歴を保存"}
            </button>
          </form>
        </div>

        <div className="panel" style={{ marginTop: "1rem" }}>
          <div className="panel-header">
            <div>
              <div className="panel-title">話す</div>
              <div className="panel-subtitle">このぬいぐるみに一言話してもらいます</div>
            </div>
          </div>

          <button
            className="btn btn-primary"
            onClick={handleChat}
            disabled={chatting}
            style={{ marginBottom: "1rem" }}
          >
            {chatting ? "生成中..." : "話す"}
          </button>

          {chatMessage && (
            <div
              style={{
                padding: "1rem",
                backgroundColor: "#f5f5f5",
                borderRadius: "8px",
                border: "1px solid #ddd",
                whiteSpace: "pre-wrap"
              }}
            >
              <strong>{plushie.name}:</strong> {chatMessage}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

