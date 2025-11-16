import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  apiCreatePlushie,
  apiDeletePlushie,
  apiListPlushies,
  apiUpdatePlushie,
  Plushie
} from "../api";

export function PlushiesPage() {
  const [items, setItems] = useState<Plushie[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [name, setName] = useState("");
  const [kind, setKind] = useState("");
  const [adoptedAt, setAdoptedAt] = useState("");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [saving, setSaving] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const list = await apiListPlushies();
      setItems(list);
    } catch (err) {
      if (err instanceof Error) setError(err.message);
      else setError("読み込みに失敗しました");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (editingId != null) {
        await apiUpdatePlushie(editingId, { name, kind, adoptedAt, imageFile });
      } else {
        await apiCreatePlushie({ name, kind, adoptedAt, imageFile });
      }
      setName("");
      setKind("");
      setAdoptedAt("");
      setImageFile(null);
      setEditingId(null);
      // reload list
      await load();
    } catch (err) {
      if (err instanceof Error) setError(err.message);
      else setError("登録に失敗しました");
    } finally {
      setSaving(false);
    }
  }

  function startEdit(p: Plushie) {
    setEditingId(p.id);
    setName(p.name);
    setKind(p.kind);
    setAdoptedAt(p.adopted_at ?? "");
    setImageFile(null);
    setError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setName("");
    setKind("");
    setAdoptedAt("");
    setImageFile(null);
    setError(null);
  }

  async function handleDelete(id: number) {
    if (!window.confirm("このぬいぐるみを削除しますか？")) return;
    try {
      await apiDeletePlushie(id);
      await load();
    } catch (err) {
      if (err instanceof Error) alert(err.message);
      else alert("削除に失敗しました");
    }
  }

  return (
    <div className="content-grid">
      <div>
        <div className="panel">
          <div className="panel-header">
            <div>
              <div className="panel-title">
                {editingId != null ? "ぬいぐるみを編集する" : "ぬいぐるみを登録する"}
              </div>
              <div className="panel-subtitle">
                名前・種類・お迎え日・写真を
                {editingId != null ? "更新します。" : "セットで記録します。"}
              </div>
            </div>
            {editingId != null ? (
              <button type="button" className="btn btn-ghost" onClick={cancelEdit}>
                編集をやめる
              </button>
            ) : (
              <span className="panel-badge">New</span>
            )}
          </div>
          <form onSubmit={handleSubmit}>
            <div className="field">
              <label htmlFor="name">名前 *</label>
              <input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                placeholder="しろくまさん"
              />
            </div>
            <div className="field">
              <label htmlFor="kind">種類 / キャラクター *</label>
              <input
                id="kind"
                value={kind}
                onChange={(e) => setKind(e.target.value)}
                required
                placeholder="くま / ポムポムプリン など"
              />
            </div>
            <div className="field">
              <label htmlFor="adoptedAt">お迎え日</label>
              <input
                id="adoptedAt"
                type="date"
                value={adoptedAt}
                onChange={(e) => setAdoptedAt(e.target.value)}
              />
            </div>
            <div className="field">
              <label htmlFor="image">写真</label>
              <input
                id="image"
                type="file"
                accept="image/*"
                onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
              />
              <div className="helper">スマホなどで撮った写真ファイルを選択してください。</div>
            </div>

            {error && <div className="error-text">{error}</div>}

            <button className="btn btn-primary" type="submit" disabled={saving}>
              {saving ? (editingId != null ? "更新中..." : "登録中...") : editingId != null ? "更新する" : "登録する"}
            </button>
          </form>
        </div>
      </div>

      <div>
        <div className="panel">
          <div className="list-header">
            <div>
              <h2>ぬいぐるみ一覧</h2>
              <span>登録したぬいぐるみをカード表示します。</span>
            </div>
          </div>
          {loading ? (
            <div className="helper">読み込み中...</div>
          ) : (items ?? []).length === 0 ? (
            <div className="empty-state">まだ登録はありません。左のフォームから追加してみてください。</div>
          ) : (
            <>
              <div className="plushies-list">
                {(items ?? []).map((p) => (
                  <article key={p.id} className="plush-card">
                    {p.image_url ? (
                      <img src={p.image_url} alt={p.name} className="plush-image" />
                    ) : (
                      <div className="plush-image" />
                    )}
                    <div className="plush-name">{p.name}</div>
                    <div className="plush-meta">
                      {p.adopted_at && <span>お迎え日: {p.adopted_at}</span>}
                      {!p.adopted_at && <span>お迎え日: -</span>}
                    </div>
                    <div className="plush-footer">
                      <span className="badge-kind">{p.kind}</span>
                      <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
                        <Link
                          to={`/plushies/${p.id}`}
                          className="btn btn-ghost"
                          style={{ textDecoration: "none" }}
                        >
                          詳細を見る
                        </Link>
                        <button
                          type="button"
                          className="btn btn-ghost"
                          onClick={() => startEdit(p)}
                        >
                          編集
                        </button>
                        <button
                          type="button"
                          className="btn btn-ghost"
                          onClick={() => handleDelete(p.id)}
                        >
                          削除
                        </button>
                      </div>
                    </div>
                  </article>
                ))}
              </div>

              <div style={{ marginTop: 12 }}>
                <div className="panel-subtitle" style={{ marginBottom: 6 }}>
                  一覧表ビュー
                </div>
                <div style={{ overflowX: "auto" }}>
                  <table
                    style={{
                      width: "100%",
                      borderCollapse: "collapse",
                      fontSize: 12,
                      background: "#f9fafb",
                      borderRadius: 12
                    }}
                  >
                    <thead>
                      <tr>
                        <th style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>名前</th>
                        <th style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>種類</th>
                        <th style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>お迎え日</th>
                        <th style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>画像</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(items ?? []).map((p) => (
                        <tr key={`row-${p.id}`}>
                          <td style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>
                            {p.name}
                          </td>
                          <td style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>
                            {p.kind}
                          </td>
                          <td style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>
                            {p.adopted_at || "-"}
                          </td>
                          <td style={{ padding: 6, borderBottom: "1px solid #e5e7eb" }}>
                            {p.image_url ? "あり" : "なし"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}


