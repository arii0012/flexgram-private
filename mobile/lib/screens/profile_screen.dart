import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api.dart';
import '../services/auth.dart';
import 'gifts_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});
  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> with SingleTickerProviderStateMixin {
  Map<String, dynamic>? _user;
  List<dynamic> _myGifts = [];
  late TabController _tab;

  @override
  void initState() {
    super.initState();
    _tab = TabController(length: 3, vsync: this);
    _load();
  }

  Future<void> _load() async {
    final api = context.read<ApiService>();
    final me = await api.getMe();
    final gifts = await api.getMyGifts();
    if (me != null) {
      setState(() {
        _user = me;
        _myGifts = gifts;
      });
      context.read<AuthService>().updateUser(me);
    }
  }

  Color _parseColor(String hex) {
    hex = hex.replaceAll('#', '');
    if (hex.length == 6) hex = 'FF$hex';
    try {
      return Color(int.parse(hex, radix: 16));
    } catch (_) {
      return const Color(0xFF2AABEE);
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = _user ?? context.watch<AuthService>().user ?? {};
    final color = _parseColor(user['profile_color']?.toString() ?? '#2AABEE');
    final name = user['display_name'] ?? user['username'] ?? '';
    final username = user['username'] ?? '';
    final bio = user['bio'] ?? '';
    final balance = user['balance'] ?? 0;
    final verified = user['is_verified'] == true;
    final premium = user['is_premium'] == true;
    final emoji = user['status_emoji'] ?? '';

    return Scaffold(
      backgroundColor: const Color(0xFF0E1621),
      body: NestedScrollView(
        headerSliverBuilder: (context, inner) => [
          SliverAppBar(
            expandedHeight: 240,
            pinned: true,
            backgroundColor: const Color(0xFF17212B),
            flexibleSpace: FlexibleSpaceBar(
              background: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [color, color.withOpacity(0.7), const Color(0xFF0E1621)],
                  ),
                ),
                child: SafeArea(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const SizedBox(height: 36),
                      Stack(
                        alignment: Alignment.bottomRight,
                        children: [
                          CircleAvatar(
                            radius: 50,
                            backgroundColor: Colors.white24,
                            child: Text(
                              name.isNotEmpty ? name[0].toUpperCase() : '?',
                              style: const TextStyle(fontSize: 40, color: Colors.white, fontWeight: FontWeight.bold),
                            ),
                          ),
                          if (premium)
                            Container(
                              padding: const EdgeInsets.all(4),
                              decoration: const BoxDecoration(
                                color: Colors.amber,
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(Icons.star, size: 16, color: Colors.black),
                            ),
                        ],
                      ),
                      const SizedBox(height: 12),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          if (emoji.isNotEmpty) Text('$emoji ', style: const TextStyle(fontSize: 20)),
                          Text(name, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Colors.white)),
                          if (verified) ...[
                            const SizedBox(width: 6),
                            const Icon(Icons.verified, color: Color(0xFF2AABEE), size: 22),
                          ],
                        ],
                      ),
                      Text('@$username', style: TextStyle(color: Colors.white.withOpacity(0.85), fontSize: 15)),
                    ],
                  ),
                ),
              ),
            ),
          ),
          SliverPersistentHeader(
            pinned: true,
            delegate: _TabBarDelegate(
              TabBar(
                controller: _tab,
                indicatorColor: const Color(0xFF2AABEE),
                labelColor: const Color(0xFF2AABEE),
                unselectedLabelColor: Colors.grey,
                tabs: const [
                  Tab(text: 'О себе'),
                  Tab(text: 'Подарки'),
                  Tab(text: 'Архив'),
                ],
              ),
            ),
          ),
        ],
        body: TabBarView(
          controller: _tab,
          children: [
            // About
            ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (bio.toString().isNotEmpty) ...[
                  const Text('О себе', style: TextStyle(color: Colors.grey, fontSize: 13)),
                  const SizedBox(height: 4),
                  Text(bio.toString(), style: const TextStyle(fontSize: 15)),
                  const SizedBox(height: 20),
                ],
                _infoTile(Icons.star, 'Баланс Stars', '$balance ⭐', Colors.amber),
                _infoTile(Icons.workspace_premium, 'Premium', premium ? 'Активен' : 'Нет', premium ? Colors.amber : Colors.grey),
                _infoTile(Icons.verified, 'Верификация', verified ? 'Есть' : 'Нет', verified ? const Color(0xFF2AABEE) : Colors.grey),
                const SizedBox(height: 12),
                ListTile(
                  leading: const Icon(Icons.card_giftcard, color: Color(0xFF2AABEE)),
                  title: const Text('Открыть витрину подарков'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const GiftsScreen())),
                ),
              ],
            ),

            // Gifts tab - visual grid
            _myGifts.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Text('🎁', style: TextStyle(fontSize: 56)),
                        const SizedBox(height: 8),
                        Text('Нет подарков', style: TextStyle(color: Colors.grey[400])),
                        TextButton(
                          onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const GiftsScreen())),
                          child: const Text('В магазин'),
                        ),
                      ],
                    ),
                  )
                : GridView.builder(
                    padding: const EdgeInsets.all(12),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 3,
                      childAspectRatio: 0.85,
                      crossAxisSpacing: 10,
                      mainAxisSpacing: 10,
                    ),
                    itemCount: _myGifts.length,
                    itemBuilder: (_, i) {
                      final ug = _myGifts[i];
                      final g = ug['gift'] ?? {};
                      final rarity = (g['rarity'] ?? 'common').toString();
                      final serial = ug['serial_number'];
                      return Container(
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(14),
                          gradient: LinearGradient(
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                            colors: _rarityColors(rarity),
                          ),
                        ),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(_emoji(rarity), style: const TextStyle(fontSize: 32)),
                            const SizedBox(height: 4),
                            Text(
                              g['name'] ?? '',
                              style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.white),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              textAlign: TextAlign.center,
                            ),
                            if (serial != null)
                              Text('#$serial', style: const TextStyle(fontSize: 12, color: Colors.white70, fontWeight: FontWeight.w600)),
                          ],
                        ),
                      );
                    },
                  ),

            // Archive placeholder
            const Center(child: Text('Архив пуст', style: TextStyle(color: Colors.grey))),
          ],
        ),
      ),
    );
  }

  Widget _infoTile(IconData icon, String title, String value, Color color) {
    return ListTile(
      leading: Icon(icon, color: color),
      title: Text(title),
      trailing: Text(value, style: TextStyle(color: color, fontWeight: FontWeight.w600)),
    );
  }

  List<Color> _rarityColors(String r) {
    switch (r.toLowerCase()) {
      case 'mythic':
        return [const Color(0xFFFF6B9D), const Color(0xFF6B2D5C)];
      case 'legendary':
        return [const Color(0xFFFF9F43), const Color(0xFFB33939)];
      case 'epic':
        return [const Color(0xFFA55EEA), const Color(0xFF5F27CD)];
      case 'rare':
        return [const Color(0xFF54A0FF), const Color(0xFF1B4F72)];
      default:
        return [const Color(0xFF636E72), const Color(0xFF2D3436)];
    }
  }

  String _emoji(String r) {
    switch (r.toLowerCase()) {
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

class _TabBarDelegate extends SliverPersistentHeaderDelegate {
  final TabBar tabBar;
  _TabBarDelegate(this.tabBar);

  @override
  Widget build(BuildContext context, double shrinkOffset, bool overlapsContent) {
    return Container(color: const Color(0xFF17212B), child: tabBar);
  }

  @override
  double get maxExtent => tabBar.preferredSize.height;
  @override
  double get minExtent => tabBar.preferredSize.height;
  @override
  bool shouldRebuild(covariant _TabBarDelegate old) => false;
}
