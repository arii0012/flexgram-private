import { useStore } from '../store/useStore';

export default function SettingsScreen({ onBack, onOpenGifts }: { onBack: () => void; onOpenGifts: () => void }) {
  const { user, settings, setTheme, logout, resetDemo, clearHistory, exportData, addStars, togglePremium, wallet } = useStore();

  const items = [
    { icon: '👤', label: 'Профиль', value: user?.name, action: () => alert(`${user?.name}\n@${user?.username}\n${user?.bio}`) },
    { icon: '⭐', label: 'Звёзды и Premium', value: `${user?.stars ?? 0} ⭐`, action: onOpenGifts },
    { icon: '🎨', label: 'Тема', value: settings.theme === 'dark' ? 'Тёмная' : settings.theme === 'light' ? 'Светлая' : 'Системная', action: () => {
      const next = settings.theme === 'dark' ? 'light' : settings.theme === 'light' ? 'system' : 'dark';
      setTheme(next);
    }},
    { icon: '💼', label: 'Кошелёк', value: `${wallet.ton} TON`, action: () => alert(`TON: ${wallet.ton}\nUSDT: ${wallet.usdt}\nNotcoin: ${wallet.notcoin.toLocaleString()}`) },
    { icon: '🔔', label: 'Уведомления', action: () => alert('Локальные уведомления (демо)') },
    { icon: '🔒', label: 'Конфиденциальность', action: () => alert('Все данные только на этом устройстве') },
    { icon: '📤', label: 'Экспорт данных', action: () => {
      const data = exportData();
      const blob = new Blob([data], { type: 'application/json' });
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = 'flexgram-backup.json';
      a.click();
    }},
    { icon: '🗑', label: 'Очистить историю', action: () => { if (confirm('Удалить все сообщения?')) clearHistory(); } },
    { icon: '🔄', label: 'Сбросить демо', action: () => { if (confirm('Сбросить?')) resetDemo(); } },
    { icon: '⭐', label: 'Переключить Premium', action: () => togglePremium() },
    { icon: '💎', label: '+100 звёзд (демо)', action: () => addStars(100) },
    { icon: '🚪', label: 'Выйти', danger: true, action: logout },
  ];

  return (
    <div className="screen">
      <div className="header">
        <button className="btn-icon" onClick={onBack}>‹</button>
        <div className="info"><div className="title">Настройки</div></div>
      </div>
      <div className="scroll" style={{ padding: '8px 0' }}>
        <div style={{
          margin: '0 12px 12px', padding: '20px 16px', borderRadius: 14,
          background: 'linear-gradient(135deg,#1a6fa8,#2AABEE 60%,#5cc8ff)', color: '#fff', textAlign: 'center',
        }}>
          <div className="avatar" style={{ width: 80, height: 80, fontSize: 32, margin: '0 auto 10px', background: 'rgba(255,255,255,0.25)', border: '3px solid rgba(255,255,255,0.4)' }}>
            {user?.name?.[0] || '?'}
          </div>
          <div style={{ fontSize: 20, fontWeight: 700 }}>{user?.name} {user?.premium && '⭐'}</div>
          <div style={{ opacity: 0.9, marginTop: 2 }}>@{user?.username}</div>
        </div>
        <div style={{ margin: '0 12px', background: 'var(--bg2)', borderRadius: 14, overflow: 'hidden' }}>
          {items.map((item, i) => (
            <div key={i} onClick={item.action} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '13px 16px',
              borderBottom: i < items.length - 1 ? '1px solid var(--sep)' : 'none',
              cursor: 'pointer', color: item.danger ? 'var(--danger)' : undefined,
            }}>
              <div style={{ width: 32, height: 32, borderRadius: 9, background: 'var(--bg3)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16 }}>{item.icon}</div>
              <span style={{ flex: 1, fontSize: 15.5 }}>{item.label}</span>
              {item.value && <span style={{ color: 'var(--text2)', fontSize: 14 }}>{item.value}</span>}
              <span style={{ color: 'var(--text3)' }}>›</span>
            </div>
          ))}
        </div>
        <p style={{ textAlign: 'center', color: 'var(--text3)', fontSize: 12, margin: '20px 16px' }}>
          FlexGram Private v1.0 · полностью офлайн
        </p>
      </div>
    </div>
  );
}
