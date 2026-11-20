import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api.dart';
import '../services/auth.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});
  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  Map<String, dynamic>? _user;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = context.read<ApiService>();
    final me = await api.getMe();
    if (me != null) {
      setState(() => _user = me);
      context.read<AuthService>().updateUser(me);
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = _user ?? context.watch<AuthService>().user ?? {};
    final color = _parseColor(user['profile_color'] ?? '#2AABEE');
    final name = user['display_name'] ?? user['username'] ?? '';
    final username = user['username'] ?? '';
    final bio = user['bio'] ?? '';
    final balance = user['balance'] ?? 0;
    final verified = user['is_verified'] == true;
    final premium = user['is_premium'] == true;

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 220,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [color, color.withOpacity(0.6), const Color(0xFF0E1621)],
                  ),
                ),
                child: SafeArea(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const SizedBox(height: 40),
                      CircleAvatar(
                        radius: 48,
                        backgroundColor: Colors.white24,
                        child: Text(
                          name.isNotEmpty ? name[0].toUpperCase() : '?',
                          style: const TextStyle(fontSize: 36, color: Colors.white),
                        ),
                      ),
                      const SizedBox(height: 12),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(name, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
                          if (verified) ...[
                            const SizedBox(width: 6),
                            const Icon(Icons.verified, color: Color(0xFF2AABEE), size: 20),
                          ],
                          if (premium) ...[
                            const SizedBox(width: 4),
                            const Icon(Icons.star, color: Colors.amber, size: 18),
                          ],
                        ],
                      ),
                      Text('@$username', style: TextStyle(color: Colors.white.withOpacity(0.8))),
                    ],
                  ),
                ),
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (bio.isNotEmpty) ...[
                    const Text('Bio', style: TextStyle(color: Colors.grey)),
                    Text(bio),
                    const SizedBox(height: 16),
                  ],
                  ListTile(
                    leading: const Icon(Icons.star, color: Colors.amber),
                    title: const Text('Stars balance'),
                    trailing: Text('$balance', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.card_giftcard),
                    title: const Text('My Gifts'),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () {},
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Color _parseColor(String hex) {
    hex = hex.replaceAll('#', '');
    if (hex.length == 6) hex = 'FF$hex';
    return Color(int.parse(hex, radix: 16));
  }
}
