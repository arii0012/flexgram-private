import 'dart:convert';
import 'package:http/http.dart' as http;
import 'auth.dart';

class ApiService {
  final AuthService auth;
  ApiService(this.auth);

  String get base => AuthService.baseUrl;

  Map<String, String> get _headers => {
        'Content-Type': 'application/json',
        if (auth.token != null) 'Authorization': 'Bearer ${auth.token}',
      };

  Future<List<dynamic>> getChats() async {
    final res = await http.get(Uri.parse('$base/api/chats'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['chats'] ?? [];
    return [];
  }

  Future<List<dynamic>> getMessages(String chatId) async {
    final res = await http.get(Uri.parse('$base/api/chats/$chatId/messages'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['messages'] ?? [];
    return [];
  }

  Future<Map<String, dynamic>?> sendMessage(String chatId, String content, {String? replyTo}) async {
    final body = {'content': content};
    if (replyTo != null) body['reply_to'] = replyTo;
    final res = await http.post(
      Uri.parse('$base/api/chats/$chatId/messages'),
      headers: _headers,
      body: jsonEncode(body),
    );
    if (res.statusCode == 201) return jsonDecode(res.body)['message'];
    return null;
  }

  Future<String?> createDirectChat(String username) async {
    final res = await http.post(
      Uri.parse('$base/api/chats'),
      headers: _headers,
      body: jsonEncode({'type': 'direct', 'username': username}),
    );
    final data = jsonDecode(res.body);
    if (res.statusCode == 201 || res.statusCode == 200) return data['chat_id'];
    return null;
  }

  Future<List<dynamic>> searchUsers(String q) async {
    final res = await http.get(Uri.parse('$base/api/users/search?q=$q'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['users'] ?? [];
    return [];
  }

  Future<Map<String, dynamic>?> getMe() async {
    final res = await http.get(Uri.parse('$base/api/users/me'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['user'];
    return null;
  }

  Future<List<dynamic>> getShopGifts() async {
    final res = await http.get(Uri.parse('$base/api/gifts/shop'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['gifts'] ?? [];
    return [];
  }

  Future<List<dynamic>> getMyGifts() async {
    final res = await http.get(Uri.parse('$base/api/gifts/my'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['gifts'] ?? [];
    return [];
  }

  Future<bool> buyGift(String giftId) async {
    final res = await http.post(Uri.parse('$base/api/gifts/buy'), headers: _headers, body: jsonEncode({'gift_id': giftId}));
    return res.statusCode == 200;
  }

  // --- Telegram extras ---
  Future<void> sendTyping(String chatId, bool isTyping) async {
    await http.post(Uri.parse('$base/api/chats/$chatId/typing'), headers: _headers, body: jsonEncode({'is_typing': isTyping}));
  }

  Future<void> markRead(String chatId) async {
    await http.post(Uri.parse('$base/api/chats/$chatId/read'), headers: _headers);
  }

  Future<Map<String, dynamic>?> react(String messageId, String emoji) async {
    final res = await http.post(
      Uri.parse('$base/api/messages/$messageId/react'),
      headers: _headers,
      body: jsonEncode({'emoji': emoji}),
    );
    if (res.statusCode == 200) return jsonDecode(res.body);
    return null;
  }

  Future<bool> pinMessage(String chatId, String messageId) async {
    final res = await http.post(
      Uri.parse('$base/api/chats/$chatId/pin'),
      headers: _headers,
      body: jsonEncode({'message_id': messageId}),
    );
    return res.statusCode == 200;
  }

  Future<List<dynamic>> getPinned(String chatId) async {
    final res = await http.get(Uri.parse('$base/api/chats/$chatId/pinned'), headers: _headers);
    if (res.statusCode == 200) return jsonDecode(res.body)['pinned'] ?? [];
    return [];
  }

  Future<void> setOnline(bool online) async {
    await http.post(Uri.parse('$base/api/users/online'), headers: _headers, body: jsonEncode({'online': online}));
  }

  // God Mode
  Future<int?> addStars(int amount) async {
    final res = await http.post(Uri.parse('$base/api/god/stars'), headers: _headers, body: jsonEncode({'amount': amount}));
    if (res.statusCode == 200) return jsonDecode(res.body)['balance'];
    return null;
  }

  Future<bool> setPremium({bool premium = true, bool verified = true, String color = '#2AABEE'}) async {
    final res = await http.post(Uri.parse('$base/api/god/premium'), headers: _headers,
        body: jsonEncode({'premium': premium, 'verified': verified, 'color': color}));
    return res.statusCode == 200;
  }

  Future<bool> grantGift(String giftId, {int serial = 0}) async {
    final res = await http.post(Uri.parse('$base/api/god/gift'), headers: _headers,
        body: jsonEncode({'gift_id': giftId, 'serial_number': serial}));
    return res.statusCode == 200;
  }

  Future<bool> changeUsername(String username) async {
    final res = await http.post(Uri.parse('$base/api/god/username'), headers: _headers, body: jsonEncode({'username': username}));
    return res.statusCode == 200;
  }
}
