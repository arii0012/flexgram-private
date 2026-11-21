import { useStore } from '../store/useStore';

const FOLDERS = [
  { id: 'all', label: 'Все' },
  { id: 'personal', label: 'Личные' },
  { id: 'groups', label: 'Группы' },
  { id: 'bots', label: 'Боты' },
  { id: 'archive', label: 'Архив' },
];

export default function MainScreen({ onOpenSettings, onOpenGifts }: { onOpenSettings: () => void; onOpenGifts: () => void }) {
  const { chats, currentFolder, searchQuery, stories, setCurrentChat, setFolder, setSearch, createChat, user } = useStore();

  let list = chats.filter(c => {
    if (currentFolder === 'archive') return c.archived;
    if (c.archived) return false;
    if (currentFolder === 'personal') return c.folder === 'personal';
    if (currentFolder === 'groups') return c.folder === 'groups' || c.type === 'group';
    if (currentFolder === 'bots') return c.folder === 'bots' || c.type === 'bot';
    return true;
  });
  if (searchQuery) {
    const q = searchQuery.toLowerCase();
    list = list.filter(c => c.name.toLowerCase().includes(q) || c.username.includes(q));
  }
  list = [...list].sort((a, b) => (b.pinned ? 1 : 0) - (a.pinned ? 1 : 0));

  const totalUnread = chats.filter(c => !c.archived).reduce((s, c) => s + c.unread, 0);

  return (
    <div className="screen" style={{ position: 'relative' }}>
      <div className="header">
        <div className="info" style={{ paddingLeft: 12 }}>
          <div className="title">FlexGram {user?.premium && <span style={{ color: '#e2a017' }}>⭐</span>}</div>
          <div className="sub">Приватный · локальный</div>
        </div>
        <button className="btn-icon" onClick={onOpenGifts}>🎁</button>
        <button className="btn-icon" onClick={onOpenSettings}>☰</button>
      </div>

      {/* Stories */}
      <div style={{ display: 'flex', gap: 14, padding: '12px 16px 8px', overflowX: 'auto' }}>
        {stories.map(s => (
          <div key={s.id} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4, minWidth: 64 }}>
            <div style={{
              width: 58, height: 58, borderRadius: '50%', padding: 2.5,
              background: s.seen ? 'var(--bg3)' : 'linear-gradient(45deg,#f09433,#dc2743,#bc1888)',
            }}>
              <div className="avatar" style={{ width: '100%', height: '100%', background: s.color, fontSize: 20, border: '2.5px solid var(--bg)' }}>
                {s.own ? '+' : s.name[0]}
              </div>
            </div>
            <span style={{ fontSize: 12, color: 'var(--text2)' }}>{s.name}</span>
          </div>
        ))}
      </div>

      {/* Folders */}
      <div style={{ display: 'flex', overflowX: 'auto', borderBottom: '1px solid var(--sep)', padding: '0 4px' }}>
        {FOLDERS.map(f => (
          <button key={f.id} onClick={() => setFolder(f.id)} style={{
            padding: '12px 14px 10px', fontSize: 14.5, fontWeight: 500, border: 'none', background: 'none',
            color: currentFolder === f.id ? 'var(--accent)' : 'var(--text2)',
            borderBottom: currentFolder === f.id ? '2.5px solid var(--accent)' : '2.5px solid transparent',
            whiteSpace: 'nowrap',
          }}>{f.label}</button>
        ))}
      </div>

      <div style={{ padding: '8px 12px' }}>
        <input
          value={searchQuery}
          onChange={e => setSearch(e.target.value)}
          placeholder="Поиск"
          style={{
            width: '100%', background: 'var(--bg2)', border: 'none', borderRadius: 11,
            padding: '10px 14px', fontSize: 15, outline: 'none',
          }}
        />
      </div>

      <div className="scroll">
        {list.map(c => (
          <div key={c.id} onClick={() => setCurrentChat(c.id)} style={{
            display: 'flex', gap: 12, padding: '9px 14px', alignItems: 'center', cursor: 'pointer',
          }}>
            <div className="avatar" style={{ background: c.avatarColor }}>
              {c.type === 'saved' ? '📌' : c.name[0]}
              {c.online && <span className="dot" />}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 3 }}>
                <span style={{ fontWeight: 600, fontSize: 16 }}>
                  {c.name} {c.verified && <span style={{ color: 'var(--accent)' }}>✓</span>}
                  {c.premium && <span style={{ color: '#e2a017' }}> ⭐</span>}
                  {c.type === 'secret' && ' 🔒'}
                </span>
                <span style={{ fontSize: 12.5, color: 'var(--text2)' }}>{c.lastTime}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ fontSize: 14.5, color: 'var(--text2)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {c.lastMessage}
                </span>
                {c.unread > 0 && (
                  <span style={{
                    background: 'var(--accent)', color: '#fff', fontSize: 12.5, fontWeight: 600,
                    minWidth: 20, height: 20, borderRadius: 10, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '0 6px',
                  }}>{c.unread}</span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      <button className="fab" onClick={() => {
        const name = prompt('Имя нового чата:');
        if (name?.trim()) {
          const id = createChat(name.trim());
          setCurrentChat(id);
        }
      }}>✏️</button>

      <div className="tabbar mobile-only">
        <button className="tab"><span className="ico">👤</span>Контакты</button>
        <button className="tab"><span className="ico">📞</span>Звонки</button>
        <button className="tab active">
          <span className="ico">💬</span>Чаты
          {totalUnread > 0 && <span className="badge">{totalUnread}</span>}
        </button>
        <button className="tab" onClick={onOpenSettings}><span className="ico">⚙️</span>Настройки</button>
      </div>
    </div>
  );
}
