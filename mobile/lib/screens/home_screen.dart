import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api.dart';
import '../services/auth.dart';
import 'chat_screen.dart';
import 'profile_screen.dart';
import 'gifts_screen.dart';
import 'god_mode_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});
  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<dynamic> _chats = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    final api = context.read<ApiService>();
    final chats = await api.getChats();
    setState(() { _chats = chats; _loading = false; });
  }

  void _newChat() async {
    final ctrl = TextEditingController();
    final username = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('New chat'),
        content: TextField(
          controller: ctrl,
          decoration: const InputDecoration(hintText: 'Username (@user)'),
          autofocus: true,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          TextButton(onPressed: () => Navigator.pop(ctx, ctrl.text.trim()), child: const Text('Start')),
        ],
      ),
    );
    if (username == null || username.isEmpty) return;
    final clean = username.replaceAll('@', '');
    final api = context.read<ApiService>();
    final chatId = await api.createDirectChat(clean);
    if (chatId != null && mounted) {
      await _load();
      Navigator.push(context, MaterialPageRoute(
        builder: (_) => ChatScreen(chatId: chatId, title: '@$clean'),
      ));
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('User not found')));
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthService>();
    return Scaffold(
      appBar: AppBar(
        title: const Text('FlexGram'),
        actions: [
          IconButton(icon: const Icon(Icons.card_giftcard), onPressed: () {
            Navigator.push(context, MaterialPageRoute(builder: (_) => const GiftsScreen()));
          }),
          IconButton(icon: const Icon(Icons.person), onPressed: () {
            Navigator.push(context, MaterialPageRoute(builder: (_) => const ProfileScreen()));
          }),
          PopupMenuButton(
            itemBuilder: (_) => [
              const PopupMenuItem(value: 'god', child: Text('Developer Panel')),
              const PopupMenuItem(value: 'logout', child: Text('Logout')),
            ],
            onSelected: (v) {
              if (v == 'logout') auth.logout();
              if (v == 'god') {
                Navigator.push(context, MaterialPageRoute(builder: (_) => const GodModeScreen()));
              }
            },
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _load,
              child: _chats.isEmpty
                  ? ListView(children: const [
                      SizedBox(height: 120),
                      Center(child: Text('No chats yet.\nTap + to start a conversation.', textAlign: TextAlign.center)),
                    ])
                  : ListView.builder(
                      itemCount: _chats.length,
                      itemBuilder: (_, i) {
                        final c = _chats[i];
                        final peer = c['peer'];
                        final title = c['title'] ?? (peer != null ? (peer['display_name'] ?? peer['username']) : 'Chat');
                        final last = c['last_message'];
                        final subtitle = last != null ? last['content'] ?? '' : '';
                        final online = peer != null && (peer['is_online'] == true);
                        return ListTile(
                          leading: CircleAvatar(
                            backgroundColor: const Color(0xFF2AABEE),
                            child: Text((title as String).isNotEmpty ? title[0].toUpperCase() : '?'),
                          ),
                          title: Row(
                            children: [
                              Expanded(child: Text(title, maxLines: 1, overflow: TextOverflow.ellipsis)),
                              if (online) const Icon(Icons.circle, size: 10, color: Colors.greenAccent),
                            ],
                          ),
                          subtitle: Text(subtitle, maxLines: 1, overflow: TextOverflow.ellipsis),
                          onTap: () {
                            Navigator.push(context, MaterialPageRoute(
                              builder: (_) => ChatScreen(chatId: c['id'], title: title, isOnline: peer != null && (peer['is_online'] == true)),
                            )).then((_) => _load());
                          },
                        );
                      },
                    ),
            ),
      floatingActionButton: FloatingActionButton(
        onPressed: _newChat,
        backgroundColor: const Color(0xFF2AABEE),
        child: const Icon(Icons.edit),
      ),
    );
  }
}
