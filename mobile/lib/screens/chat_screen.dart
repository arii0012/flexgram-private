import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../services/api.dart';
import '../services/auth.dart';

class ChatScreen extends StatefulWidget {
  final String chatId;
  final String title;
  final bool isOnline;
  const ChatScreen({super.key, required this.chatId, required this.title, this.isOnline = false});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _ctrl = TextEditingController();
  final _scroll = ScrollController();
  List<dynamic> _messages = [];
  List<dynamic> _pinned = [];
  bool _loading = true;
  bool _peerTyping = false;
  String? _replyToId;
  String? _replyToText;
  Timer? _typingTimer;
  bool _sentTyping = false;

  static const _reactions = ['👍', '❤️', '🔥', '🥰', '👏', '😁', '🤔', '🤯', '😱', '🤬', '😢', '🎉', '🤩', '🤮', '💩', '🙏'];

  @override
  void initState() {
    super.initState();
    _load();
    context.read<ApiService>().setOnline(true);
  }

  @override
  void dispose() {
    _typingTimer?.cancel();
    _ctrl.dispose();
    _scroll.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final api = context.read<ApiService>();
    final msgs = await api.getMessages(widget.chatId);
    final pinned = await api.getPinned(widget.chatId);
    await api.markRead(widget.chatId);
    setState(() {
      _messages = msgs;
      _pinned = pinned;
      _loading = false;
    });
    _scrollToEnd();
  }

  void _scrollToEnd() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.jumpTo(_scroll.position.maxScrollExtent);
      }
    });
  }

  void _onTextChanged(String text) {
    final api = context.read<ApiService>();
    if (text.isNotEmpty && !_sentTyping) {
      _sentTyping = true;
      api.sendTyping(widget.chatId, true);
    }
    _typingTimer?.cancel();
    _typingTimer = Timer(const Duration(seconds: 2), () {
      if (_sentTyping) {
        api.sendTyping(widget.chatId, false);
        _sentTyping = false;
      }
    });
  }

  Future<void> _send() async {
    final text = _ctrl.text.trim();
    if (text.isEmpty) return;
    _ctrl.clear();
    final api = context.read<ApiService>();
    if (_sentTyping) {
      api.sendTyping(widget.chatId, false);
      _sentTyping = false;
    }
    final msg = await api.sendMessage(widget.chatId, text, replyTo: _replyToId);
    setState(() {
      _replyToId = null;
      _replyToText = null;
      if (msg != null) _messages.add(msg);
    });
    _scrollToEnd();
  }

  void _showReactionPicker(String messageId) {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF17212B),
      builder: (_) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Wrap(
            spacing: 12,
            runSpacing: 12,
            children: _reactions.map((e) {
              return GestureDetector(
                onTap: () async {
                  Navigator.pop(context);
                  await context.read<ApiService>().react(messageId, e);
                  HapticFeedback.lightImpact();
                  // optimistic: could reload
                },
                child: Text(e, style: const TextStyle(fontSize: 32)),
              );
            }).toList(),
          ),
        ),
      ),
    );
  }

  void _showMessageMenu(Map msg) {
    final isMe = msg['sender_id'] == context.read<AuthService>().user?['id'];
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF17212B),
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.reply, color: Colors.white70),
              title: const Text('Ответить'),
              onTap: () {
                Navigator.pop(context);
                setState(() {
                  _replyToId = msg['id'];
                  _replyToText = msg['content'];
                });
              },
            ),
            ListTile(
              leading: const Icon(Icons.add_reaction_outlined, color: Colors.white70),
              title: const Text('Реакция'),
              onTap: () {
                Navigator.pop(context);
                _showReactionPicker(msg['id']);
              },
            ),
            ListTile(
              leading: const Icon(Icons.push_pin, color: Colors.white70),
              title: const Text('Закрепить'),
              onTap: () async {
                Navigator.pop(context);
                await context.read<ApiService>().pinMessage(widget.chatId, msg['id']);
                final pinned = await context.read<ApiService>().getPinned(widget.chatId);
                setState(() => _pinned = pinned);
              },
            ),
            if (isMe)
              ListTile(
                leading: const Icon(Icons.delete_outline, color: Colors.redAccent),
                title: const Text('Удалить', style: TextStyle(color: Colors.redAccent)),
                onTap: () => Navigator.pop(context),
              ),
          ],
        ),
      ),
    );
  }

  String _formatTime(dynamic iso) {
    if (iso == null) return '';
    try {
      final dt = DateTime.parse(iso.toString()).toLocal();
      return '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }

  @override
  Widget build(BuildContext context) {
    final myId = context.read<AuthService>().user?['id'];
    final subtitle = _peerTyping
        ? 'печатает...'
        : (widget.isOnline ? 'в сети' : 'был(а) недавно');

    return Scaffold(
      backgroundColor: const Color(0xFF0E1621),
      appBar: AppBar(
        backgroundColor: const Color(0xFF17212B),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(widget.title, style: const TextStyle(fontSize: 16)),
            Text(
              subtitle,
              style: TextStyle(
                fontSize: 12,
                color: _peerTyping || widget.isOnline ? const Color(0xFF2AABEE) : Colors.grey,
              ),
            ),
          ],
        ),
        actions: [
          IconButton(icon: const Icon(Icons.search), onPressed: () {}),
          IconButton(icon: const Icon(Icons.more_vert), onPressed: () {}),
        ],
      ),
      body: Column(
        children: [
          // Pinned bar
          if (_pinned.isNotEmpty)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              color: const Color(0xFF1C2A3A),
              child: Row(
                children: [
                  const Icon(Icons.push_pin, size: 16, color: Color(0xFF2AABEE)),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _pinned.first['content'] ?? '',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                ],
              ),
            ),

          // Messages
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator(color: Color(0xFF2AABEE)))
                : ListView.builder(
                    controller: _scroll,
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                    itemCount: _messages.length,
                    itemBuilder: (_, i) {
                      final m = _messages[i];
                      final isMe = m['sender_id'] == myId;
                      final time = _formatTime(m['created_at']);
                      return GestureDetector(
                        onLongPress: () => _showMessageMenu(m),
                        child: Align(
                          alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
                          child: Container(
                            margin: const EdgeInsets.symmetric(vertical: 2),
                            padding: const EdgeInsets.fromLTRB(10, 6, 10, 4),
                            constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.78),
                            decoration: BoxDecoration(
                              color: isMe ? const Color(0xFF2B5278) : const Color(0xFF182533),
                              borderRadius: BorderRadius.only(
                                topLeft: const Radius.circular(14),
                                topRight: const Radius.circular(14),
                                bottomLeft: Radius.circular(isMe ? 14 : 4),
                                bottomRight: Radius.circular(isMe ? 4 : 14),
                              ),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                // reply preview
                                if (m['reply_to'] != null)
                                  Container(
                                    width: double.infinity,
                                    margin: const EdgeInsets.only(bottom: 4),
                                    padding: const EdgeInsets.only(left: 8),
                                    decoration: const BoxDecoration(
                                      border: Border(left: BorderSide(color: Color(0xFF2AABEE), width: 2)),
                                    ),
                                    child: Text(
                                      'Ответ на сообщение',
                                      style: TextStyle(fontSize: 12, color: Colors.white.withOpacity(0.6)),
                                    ),
                                  ),
                                Text(m['content'] ?? '', style: const TextStyle(fontSize: 15, height: 1.3)),
                                const SizedBox(height: 2),
                                Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Text(time, style: TextStyle(fontSize: 11, color: Colors.white.withOpacity(0.5))),
                                    if (isMe) ...[
                                      const SizedBox(width: 3),
                                      Icon(Icons.done_all, size: 14, color: Colors.white.withOpacity(0.6)),
                                    ],
                                  ],
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  ),
          ),

          // Reply bar
          if (_replyToId != null)
            Container(
              color: const Color(0xFF1C2A3A),
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Row(
                children: [
                  const Icon(Icons.reply, size: 18, color: Color(0xFF2AABEE)),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _replyToText ?? '',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, size: 18),
                    onPressed: () => setState(() {
                      _replyToId = null;
                      _replyToText = null;
                    }),
                  ),
                ],
              ),
            ),

          // Input
          Container(
            color: const Color(0xFF17212B),
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 6),
            child: SafeArea(
              child: Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.attach_file, color: Colors.grey),
                    onPressed: () {},
                  ),
                  Expanded(
                    child: TextField(
                      controller: _ctrl,
                      onChanged: _onTextChanged,
                      style: const TextStyle(fontSize: 15),
                      decoration: const InputDecoration(
                        hintText: 'Сообщение',
                        hintStyle: TextStyle(color: Colors.grey),
                        border: InputBorder.none,
                        contentPadding: EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                      ),
                      maxLines: 4,
                      minLines: 1,
                      textCapitalization: TextCapitalization.sentences,
                      onSubmitted: (_) => _send(),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.emoji_emotions_outlined, color: Colors.grey),
                    onPressed: () {},
                  ),
                  IconButton(
                    icon: const Icon(Icons.send_rounded, color: Color(0xFF2AABEE)),
                    onPressed: _send,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
