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
    final api = context.read<ApiService>();
    final shop = await api.getShopGifts();
    final my = await api.getMyGifts();
    setState(() { _shop = shop; _my = my; _loading = false; });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Gifts'),
        bottom: TabBar(
          controller: _tab,
          tabs: const [Tab(text: 'Shop'), Tab(text: 'My Gifts')],
        ),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tab,
              children: [
                _buildShop(),
                _buildMy(),
              ],
            ),
    );
  }

  Widget _buildShop() {
    if (_shop.isEmpty) return const Center(child: Text('Shop is empty'));
    return GridView.builder(
      padding: const EdgeInsets.all(12),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2, childAspectRatio: 0.75, crossAxisSpacing: 12, mainAxisSpacing: 12,
      ),
      itemCount: _shop.length,
      itemBuilder: (_, i) {
        final g = _shop[i];
        return Card(
          color: const Color(0xFF17212B),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Center(
                    child: Text(
                      _rarityEmoji(g['rarity']),
                      style: const TextStyle(fontSize: 48),
                    ),
                  ),
                ),
                Text(g['name'] ?? '', style: const TextStyle(fontWeight: FontWeight.bold)),
                Text(g['rarity'] ?? '', style: TextStyle(color: _rarityColor(g['rarity']), fontSize: 12)),
                const SizedBox(height: 4),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('${g['price']} ⭐', style: const TextStyle(color: Colors.amber)),
                    TextButton(
                      onPressed: () async {
                        final ok = await context.read<ApiService>().buyGift(g['id']);
                        if (ok && mounted) {
                          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Gift purchased!')));
                          _load();
                        }
                      },
                      child: const Text('Buy'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildMy() {
    if (_my.isEmpty) return const Center(child: Text('You have no gifts yet'));
    return ListView.builder(
      itemCount: _my.length,
      itemBuilder: (_, i) {
        final ug = _my[i];
        final g = ug['gift'] ?? {};
        final serial = ug['serial_number'];
        return ListTile(
          leading: Text(_rarityEmoji(g['rarity']), style: const TextStyle(fontSize: 32)),
          title: Text('${g['name'] ?? 'Gift'} ${serial != null ? '#$serial' : ''}'),
          subtitle: Text(g['rarity'] ?? ''),
          trailing: ug['is_equipped'] == true
              ? const Chip(label: Text('Equipped'), backgroundColor: Color(0xFF2AABEE))
              : null,
        );
      },
    );
  }

  String _rarityEmoji(String? r) {
    switch (r) {
      case 'mythic': return '👑';
      case 'legendary': return '🚀';
      case 'epic': return '💎';
      case 'rare': return '⭐';
      default: return '🎁';
    }
  }

  Color _rarityColor(String? r) {
    switch (r) {
      case 'mythic': return Colors.pinkAccent;
      case 'legendary': return Colors.orangeAccent;
      case 'epic': return Colors.purpleAccent;
      case 'rare': return Colors.blueAccent;
      default: return Colors.grey;
    }
  }
}
