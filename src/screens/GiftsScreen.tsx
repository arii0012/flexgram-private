import { useStore } from '../store/useStore';

export default function GiftsScreen({ onBack }: { onBack: () => void }) {
  const { user, giftCatalog, ownedGifts, buyGift, addStars, sendMessage, currentChatId } = useStore();

  return (
    <div className="screen">
      <div className="header">
        <button className="btn-icon" onClick={onBack}>‹</button>
        <div className="info"><div className="title">Подарки</div></div>
      </div>
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
        padding: 16, background: 'linear-gradient(135deg,#f5a623,#e2a017)', color: '#fff', fontSize: 20, fontWeight: 700,
      }}>
        <span>⭐</span>
        <span>{user?.stars ?? 0}</span>
        <button onClick={() => addStars(100)} style={{
          background: 'rgba(255,255,255,0.25)', border: 'none', color: '#fff',
          padding: '6px 14px', borderRadius: 20, fontSize: 13, fontWeight: 600, marginLeft: 12,
        }}>+ Пополнить</button>
      </div>
      <div className="scroll" style={{ padding: 12 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 12 }}>
          {giftCatalog.map(g => (
            <div key={g.id} onClick={() => {
              if (buyGift(g.id)) {
                alert(`Подарок «${g.name}» куплен!`);
                if (currentChatId) {
                  // send as gift message would need type support; simplified
                }
              } else {
                alert('Недостаточно звёзд');
              }
            }} style={{
              background: 'var(--bg2)', borderRadius: 14, padding: '14px 8px', textAlign: 'center', cursor: 'pointer',
            }}>
              <div style={{ fontSize: 36 }}>{g.emoji}</div>
              <div style={{ fontSize: 12, marginTop: 6, color: 'var(--text2)' }}>{g.name}</div>
              <div style={{ fontSize: 13, color: 'var(--accent)', marginTop: 3, fontWeight: 600 }}>⭐ {g.price}</div>
              <div style={{ fontSize: 10, color: 'var(--text3)', marginTop: 2 }}>{g.rarity}</div>
            </div>
          ))}
        </div>
        {ownedGifts.length > 0 && (
          <>
            <h3 style={{ margin: '20px 0 10px', fontSize: 15 }}>Мои подарки</h3>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 12 }}>
              {ownedGifts.map((g, i) => (
                <div key={i} style={{ background: 'var(--bg2)', borderRadius: 14, padding: '14px 8px', textAlign: 'center' }}>
                  <div style={{ fontSize: 36 }}>{g.emoji}</div>
                  <div style={{ fontSize: 12, marginTop: 6 }}>{g.name}</div>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
