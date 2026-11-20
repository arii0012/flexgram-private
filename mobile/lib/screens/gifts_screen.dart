import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api.dart';

class GiftsScreen extends StatefulWidget {
  const GiftsScreen({super.key});
  @override
  State<GiftsScreen> createState() => _GiftsScreenState();
}

class _GiftsScreenState extends State<GiftsScreen> with SingleTickerProviderStateMixin {
  late TabController _tab;
  List<dynamic> _shop = [];
  List<dynamic> _my = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _tab = TabController(length: 2, vsync: this);
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    final api = context.read<ApiService>();
    final shop = await api.getShopGifts();
    final my = await api.getMyGifts();
    setState(() {
      _shop = shop;
      _my = my;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0E1621),
      appBar: AppBar(
        title: const Text('Подарки'),
        backgroundColor: const Color(0xFF17212B),
        bottom: TabBar(
          controller: _tab,
          indicatorColor: const Color(0xFF2AABEE),
          labelColor: const Color(0xFF2AABEE),
          unselectedLabelColor: Colors.grey,
          tabs: const [
            Tab(text: 'Магазин'),
            Tab(text: 'Мои подарки'),
          ],
        ),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFF2AABEE)))
          : TabBarView(
              controller: _tab,
              children: [
                _buildShop(),
                _buildMy(),
              ],
            ),
    );
  }

  // ---------- SHOP ----------
  Widget _buildShop() {
    if (_shop.isEmpty) {
      return const Center(child: Text('Магазин пуст', style: TextStyle(color: Colors.grey)));
    }
    return GridView.builder(
      padding: const EdgeInsets.all(12),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 0.72,
        crossAxisSpacing: 12,
        mainAxisSpacing: 12,
      ),
      itemCount: _shop.length,
      itemBuilder: (_, i) => _GiftCard(
        gift: _shop[i],
        isShop: true,
        onBuy: () async {
          final ok = await context.read<ApiService>().buyGift(_shop[i]['id']);
          if (!mounted) return;
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(ok ? 'Подарок куплен! 🎁' : 'Недостаточно Stars'),
              backgroundColor: ok ? Colors.green : Colors.redAccent,
            ),
          );
          if (ok) _load();
        },
      ),
    );
  }

  // ---------- MY GIFTS ----------
  Widget _buildMy() {
    if (_my.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('🎁', style: TextStyle(fontSize: 64)),
            const SizedBox(height: 12),
            Text('У вас пока нет подарков', style: TextStyle(color: Colors.grey[400])),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => _tab.animateTo(0),
              child: const Text('Перейти в магазин'),
            ),
          ],
        ),
      );
    }
    return GridView.builder(
      padding: const EdgeInsets.all(12),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 0.72,
        crossAxisSpacing: 12,
        mainAxisSpacing: 12,
      ),
      itemCount: _my.length,
      itemBuilder: (_, i) {
        final ug = _my[i];
        final gift = ug['gift'] ?? {};
        return _GiftCard(
          gift: gift,
          serial: ug['serial_number'],
          isEquipped: ug['is_equipped'] == true,
          isShop: false,
          onEquip: () async {
            // toggle equip via API if needed
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(ug['is_equipped'] == true ? 'Снято' : 'Надето')),
            );
          },
        );
      },
    );
  }
}

// ===================== VISUAL GIFT CARD =====================
class _GiftCard extends StatelessWidget {
  final Map gift;
  final int? serial;
  final bool isEquipped;
  final bool isShop;
  final VoidCallback? onBuy;
  final VoidCallback? onEquip;

  const _GiftCard({
    required this.gift,
    this.serial,
    this.isEquipped = false,
    this.isShop = true,
    this.onBuy,
    this.onEquip,
  });

  @override
  Widget build(BuildContext context) {
    final rarity = (gift['rarity'] ?? 'common').toString().toLowerCase();
    final name = gift['name'] ?? 'Gift';
    final price = gift['price'] ?? 0;
    final colors = _rarityGradient(rarity);
    final emoji = _rarityEmoji(rarity);
    final pattern = gift['pattern'] ?? '';

    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: colors,
        ),
        boxShadow: [
          BoxShadow(
            color: colors.first.withOpacity(0.4),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: isShop ? onBuy : onEquip,
          child: Stack(
            children: [
              // Pattern overlay
              if (pattern.isNotEmpty)
                Positioned.fill(
                  child: Opacity(
                    opacity: 0.15,
                    child: CustomPaint(painter: _PatternPainter(pattern)),
                  ),
                ),

              // Content
              Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Top row: rarity badge + serial
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                          decoration: BoxDecoration(
                            color: Colors.black26,
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Text(
                            rarity.toUpperCase(),
                            style: const TextStyle(
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ),
                        if (serial != null)
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                            decoration: BoxDecoration(
                              color: Colors.black38,
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Text(
                              '#$serial',
                              style: const TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                          ),
                      ],
                    ),

                    const Spacer(),

                    // Big emoji / icon
                    Center(
                      child: Text(emoji, style: const TextStyle(fontSize: 56)),
                    ),

                    const Spacer(),

                    // Name
                    Text(
                      name,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),

                    // Bottom: price or equipped
                    if (isShop)
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '$price ⭐',
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.amberAccent,
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.2),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: const Text(
                              'Купить',
                              style: TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w600),
                            ),
                          ),
                        ],
                      )
                    else
                      Row(
                        children: [
                          if (isEquipped)
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                              decoration: BoxDecoration(
                                color: Colors.white.withOpacity(0.25),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: const Text('Надето', style: TextStyle(fontSize: 11, color: Colors.white)),
                            )
                          else
                            Text(
                              'Нажми, чтобы надеть',
                              style: TextStyle(fontSize: 11, color: Colors.white.withOpacity(0.7)),
                            ),
                        ],
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  List<Color> _rarityGradient(String rarity) {
    switch (rarity) {
      case 'mythic':
        return [const Color(0xFFFF6B9D), const Color(0xFFC44569), const Color(0xFF6B2D5C)];
      case 'legendary':
        return [const Color(0xFFFF9F43), const Color(0xFFEE5A24), const Color(0xFFB33939)];
      case 'epic':
        return [const Color(0xFFA55EEA), const Color(0xFF8854D0), const Color(0xFF5F27CD)];
      case 'rare':
        return [const Color(0xFF54A0FF), const Color(0xFF2E86DE), const Color(0xFF1B4F72)];
      default:
        return [const Color(0xFF636E72), const Color(0xFF2D3436)];
    }
  }

  String _rarityEmoji(String rarity) {
    switch (rarity) {
      case 'mythic':
        return '👑';
      case 'legendary':
        return '🚀';
      case 'epic':
        return '💎';
      case 'rare':
        return '⭐';
      default:
        return '🎁';
    }
  }
}

// Simple pattern painter for visual interest
class _PatternPainter extends CustomPainter {
  final String pattern;
  _PatternPainter(this.pattern);

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;

    if (pattern.contains('star') || pattern.contains('stars')) {
      for (var i = 0; i < 8; i++) {
        final x = (i % 4) * size.width / 3;
        final y = (i ~/ 4) * size.height / 2 + 20;
        _drawStar(canvas, Offset(x + 20, y), 8, paint);
      }
    } else if (pattern.contains('heart')) {
      // simple circles as hearts substitute
      for (var i = 0; i < 6; i++) {
        canvas.drawCircle(Offset((i % 3) * size.width / 2 + 30, (i ~/ 3) * size.height / 2 + 40), 6, paint);
      }
    } else {
      // default dots
      for (var i = 0; i < 12; i++) {
        canvas.drawCircle(
          Offset((i % 4) * size.width / 3 + 15, (i ~/ 4) * size.height / 3 + 20),
          2,
          paint..style = PaintingStyle.fill,
        );
      }
    }
  }

  void _drawStar(Canvas canvas, Offset center, double r, Paint paint) {
    // simplified star as cross
    canvas.drawLine(Offset(center.dx - r, center.dy), Offset(center.dx + r, center.dy), paint);
    canvas.drawLine(Offset(center.dx, center.dy - r), Offset(center.dx, center.dy + r), paint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
