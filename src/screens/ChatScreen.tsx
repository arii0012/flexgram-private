import { useState, useRef, useEffect } from 'react';
import { useStore } from '../store/useStore';

export default function ChatScreen() {
  const {
    chats, currentChatId, replyToId, directorMode, directorFrom, user,
    setCurrentChat, sendMessage, setReplyTo, deleteMessage, editMessage,
    toggleDirectorMode, setDirectorFrom, insertSystemMessage, triggerCall,
    setFakeStatus,
  } = useStore();
  const chat = chats.find(c => c.id === currentChatId);
  const [text, setText] = useState('');
  const [ctxMsg, setCtxMsg] = useState<string | null>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    listRef.current?.scrollTo({ top: listRef.current.scrollHeight });
  }, [chat?.messages.length]);

  if (!chat) return null;

  const onSend = () => {
    if (!text.trim()) return;
    sendMessage(text.trim());
    setText('');
  };

  const replyMsg = replyToId ? chat.messages.find(m => m.id === replyToId) : null;

  return (
    <div className="screen">
      <div className="header">
        <button className="btn-icon" onClick={() => setCurrentChat(null)}>‹</button>
        <div className="avatar sm" style={{ background: chat.avatarColor }}>{chat.name[0]}</div>
        <div className="info" onDoubleClick={() => toggleDirectorMode()} title="Двойной клик = Director Mode">
          <div className="title">
            {chat.name}
            {chat.verified && <span style={{ color: 'var(--accent)' }}>✓</span>}
            {directorMode && <span style={{ fontSize: 11, background: 'var(--danger)', color: '#fff', padding: '1px 6px', borderRadius: 6, marginLeft: 6 }}>DIRECTOR</span>}
          </div>
          <div className={`sub ${chat.online ? 'online' : ''} ${chat.fakeTyping ? 'typing' : ''}`}>
            {chat.fakeTyping ? 'печатает...' : chat.online ? 'в сети' : 'был(а) недавно'}
          </div>
        </div>
        <button className="btn-icon" onClick={() => triggerCall('audio')}>📞</button>
        <button className="btn-icon" onClick={() => {
          const acts = ['Системное сообщение', 'Накрутить просмотры', 'Статус «печатает»', 'Режим режиссёра'];
          const i = prompt(acts.map((a, n) => `${n+1}. ${a}`).join('\n'));
          if (i === '1') insertSystemMessage(chat.id, prompt('Текст:') || 'Системное событие');
          if (i === '2') setFakeStatus(chat.id, { views: 1_000_000 + Math.floor(Math.random()*9e6) });
          if (i === '3') setFakeStatus(chat.id, { typing: true });
          if (i === '4') toggleDirectorMode();
        }}>⋮</button>
      </div>

      {directorMode && (
        <div style={{ padding: '8px 12px', background: 'var(--bg2)', display: 'flex', gap: 8, alignItems: 'center', fontSize: 13 }}>
          <span>Писать от:</span>
          <button onClick={() => setDirectorFrom('me')} style={{
            padding: '4px 12px', borderRadius: 8, border: 'none',
            background: directorFrom === 'me' ? 'var(--accent)' : 'var(--bg3)', color: '#fff',
          }}>Я</button>
          <button onClick={() => setDirectorFrom('them')} style={{
            padding: '4px 12px', borderRadius: 8, border: 'none',
            background: directorFrom === 'them' ? 'var(--accent)' : 'var(--bg3)', color: '#fff',
          }}>{chat.name}</button>
        </div>
      )}

      <div className="scroll" ref={listRef} style={{ padding: '8px 8px 6px', display: 'flex', flexDirection: 'column', gap: 3 }}>
        <div style={{ textAlign: 'center', margin: '8px 0' }}>
          <span style={{ background: 'rgba(30,44,58,0.75)', color: 'var(--text2)', fontSize: 12.5, padding: '4px 12px', borderRadius: 11 }}>Сегодня</span>
        </div>
        {chat.messages.map(m => {
          if (m.from === 'system' || m.type === 'system') {
            return <div key={m.id} style={{ textAlign: 'center', fontSize: 13, color: 'var(--text2)', padding: '6px 0' }}>{m.text}</div>;
          }
          const isOut = m.from === 'me';
          return (
            <div
              key={m.id}
              onContextMenu={e => { e.preventDefault(); setCtxMsg(m.id); }}
              onTouchStart={() => { const t = setTimeout(() => setCtxMsg(m.id), 450); (e.target as any)._t = t; }}
              onTouchEnd={e => clearTimeout((e.target as any)._t)}
              style={{
                maxWidth: '78%', padding: '7px 10px 5px', borderRadius: 14, fontSize: 15.5, lineHeight: 1.35,
                alignSelf: isOut ? 'flex-end' : 'flex-start',
                background: isOut ? 'var(--bubble-out)' : 'var(--bubble-in)',
                borderBottomRightRadius: isOut ? 4 : 14,
                borderBottomLeftRadius: isOut ? 14 : 4,
              }}
            >
              {m.replyTo && (() => {
                const orig = chat.messages.find(x => x.id === m.replyTo);
                return orig ? (
                  <div style={{ borderLeft: '2.5px solid var(--accent)', paddingLeft: 8, marginBottom: 4, fontSize: 13 }}>
                    <div style={{ color: 'var(--accent)', fontWeight: 600 }}>{orig.from === 'me' ? 'Вы' : chat.name}</div>
                    <div style={{ color: 'var(--text2)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{orig.text?.slice(0, 50)}</div>
                  </div>
                ) : null;
              })()}
              {m.type === 'voice' ? (
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 180 }}>
                  <div style={{ width: 36, height: 36, borderRadius: '50%', background: 'rgba(255,255,255,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>▶</div>
                  <div style={{ flex: 1, height: 20, display: 'flex', gap: 2, alignItems: 'center' }}>
                    {Array.from({ length: 18 }, (_, i) => (
                      <span key={i} style={{ width: 3, height: 6 + Math.random() * 14, background: 'rgba(255,255,255,0.5)', borderRadius: 2 }} />
                    ))}
                  </div>
                  <span style={{ fontSize: 12, opacity: 0.5 }}>{m.duration || '0:12'}</span>
                </div>
              ) : m.type === 'gift' ? (
                <div style={{ textAlign: 'center', padding: '8px 0' }}>
                  <div style={{ fontSize: 48 }}>{m.giftEmoji || '🎁'}</div>
                  <div style={{ fontSize: 13, opacity: 0.8 }}>{m.giftName}</div>
                </div>
              ) : (
                <span style={{ whiteSpace: 'pre-wrap' }}>{m.text}</span>
              )}
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 3, marginTop: 2, float: 'right', marginLeft: 10 }}>
                {m.edited && <span style={{ fontSize: 11, opacity: 0.4 }}>изм.</span>}
                <span style={{ fontSize: 11, opacity: 0.45 }}>{m.time}</span>
                {isOut && <span style={{ fontSize: 13, color: '#7ec8f0' }}>{m.status === 'read' ? '✓✓' : '✓'}</span>}
              </div>
            </div>
          );
        })}
        {chat.fakeViews != null && (
          <div style={{ textAlign: 'center', fontSize: 12, color: 'var(--text3)', padding: 4 }}>👁 {chat.fakeViews.toLocaleString()} просмотров</div>
        )}
      </div>

      {replyMsg && (
        <div style={{ padding: '8px 12px', background: 'var(--bg2)', borderLeft: '3px solid var(--accent)', display: 'flex', gap: 10, alignItems: 'center' }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 13, color: 'var(--accent)', fontWeight: 600 }}>{replyMsg.from === 'me' ? 'Вы' : chat.name}</div>
            <div style={{ fontSize: 13, color: 'var(--text2)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{replyMsg.text}</div>
          </div>
          <button className="btn-icon" style={{ width: 32, height: 32 }} onClick={() => setReplyTo(null)}>✕</button>
        </div>
      )}

      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 4, padding: '6px 6px 8px', background: 'var(--header)', borderTop: '1px solid var(--sep)', paddingBottom: 'calc(8px + var(--safe-bottom))' }}>
        <button className="btn-icon" onClick={() => sendMessage('', 'voice', { duration: '0:' + String(5 + Math.floor(Math.random()*20)).padStart(2, '0') })}>🎤</button>
        <textarea
          value={text}
          onChange={e => setText(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); onSend(); } }}
          placeholder="Сообщение"
          rows={1}
          style={{
            flex: 1, background: 'var(--bg2)', borderRadius: 22, padding: '10px 14px',
            border: 'none', fontSize: 16, outline: 'none', resize: 'none', maxHeight: 120, minHeight: 42, lineHeight: 1.35,
          }}
        />
        <button className="btn-icon" style={{ background: text.trim() ? 'var(--accent)' : 'transparent', color: text.trim() ? '#fff' : 'var(--text2)', borderRadius: '50%' }} onClick={onSend}>
          {text.trim() ? '➤' : '😊'}
        </button>
      </div>

      {/* Context menu */}
      {ctxMsg && (
        <div onClick={() => setCtxMsg(null)} style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', backdropFilter: 'blur(6px)',
          zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <div style={{ background: 'var(--bg2)', borderRadius: 16, minWidth: 240, overflow: 'hidden' }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', gap: 4, padding: 10, justifyContent: 'space-around', borderBottom: '1px solid var(--sep)' }}>
              {['❤️','🔥','👍','😂','😮','😢'].map(r => (
                <span key={r} style={{ fontSize: 26, cursor: 'pointer' }} onClick={() => setCtxMsg(null)}>{r}</span>
              ))}
            </div>
            {[
              { act: 'reply', label: '↩️ Ответить' },
              { act: 'copy', label: '📋 Копировать' },
              { act: 'edit', label: '✏️ Изменить' },
              { act: 'delete', label: '🗑 Удалить', danger: true },
            ].map(item => (
              <div key={item.act} style={{ padding: '13px 18px', cursor: 'pointer', color: item.danger ? 'var(--danger)' : undefined }}
                onClick={() => {
                  const m = chat.messages.find(x => x.id === ctxMsg);
                  if (item.act === 'reply') setReplyTo(ctxMsg);
                  if (item.act === 'copy' && m?.text) navigator.clipboard?.writeText(m.text);
                  if (item.act === 'edit' && m) {
                    const t = prompt('Изменить:', m.text);
                    if (t != null) editMessage(chat.id, m.id, t);
                  }
                  if (item.act === 'delete') deleteMessage(chat.id, ctxMsg);
                  setCtxMsg(null);
                }}
              >{item.label}</div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
